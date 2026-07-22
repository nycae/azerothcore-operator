package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/nycae/azerothcore-operator/api/v1alpha1"
)

type Repository interface {
	Exists(ctx context.Context, account *v1alpha1.Account) (bool, int64, error)
	Create(ctx context.Context, account *v1alpha1.Account, password string) (int64, error)
	Update(ctx context.Context, account *v1alpha1.Account) error
	Delete(ctx context.Context, account *v1alpha1.Account) error
}

type repo struct {
	db *sql.DB
}

func (r *repo) Delete(ctx context.Context, account *v1alpha1.Account) error {
	accountID := account.Status.AccountID
	if accountID == 0 {
		exists, id, err := r.Exists(ctx, account)
		if err != nil {
			return fmt.Errorf("failed to check if account exists: %w", err)
		}
		if !exists {
			return nil
		}
		accountID = id
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction for deletion: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, "DELETE FROM account_access WHERE id = ?", accountID); err != nil {
		return fmt.Errorf("failed to delete from account_access: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM account_banned WHERE id = ?", accountID); err != nil {
		return fmt.Errorf("failed to delete from account_banned: %w", err)
	}

	if _, err := tx.ExecContext(ctx, "DELETE FROM account_muted WHERE guid = ?", accountID); err != nil {
		return fmt.Errorf("failed to delete from account_muted: %w", err)
	}

	res, err := tx.ExecContext(ctx, "DELETE FROM account WHERE id = ?", accountID)
	if err != nil {
		return fmt.Errorf("failed to delete account record: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return nil
	}

	return tx.Commit()
}

func (r *repo) Exists(ctx context.Context, account *v1alpha1.Account) (bool, int64, error) {
	var id int64
	err := r.db.QueryRowContext(ctx, "SELECT id FROM account WHERE USERNAME = ?", strings.ToUpper(account.Spec.Username)).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("failed to check if account exists: %w", err)
	}
	return true, id, nil
}

func (r *repo) Create(ctx context.Context, account *v1alpha1.Account, password string) (int64, error) {
	username := strings.ToUpper(account.Spec.Username)
	salt, verifier, err := CalculateSRP6(username, password)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate SRP6 creds: %w", err)
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	res, err := tx.ExecContext(ctx, "INSERT INTO account (username, salt, verifier, email, expansion) VALUES (?, ?, ?, ?, ?)",
		username, salt, verifier, account.Spec.Email, account.Spec.Expansion)
	if err != nil {
		return 0, fmt.Errorf("failed to insert account: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("failed to get account id: %w", err)
	}
	if account.Spec.GmLevel > 0 {
		if _, err := tx.ExecContext(ctx, "INSERT INTO account_access (id, gmlevel, RealmID) VALUES (?, ?, -1)",
			id, account.Spec.GmLevel); err != nil {
			return 0, fmt.Errorf("failed to insert account access: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}
	return id, nil
}

func (r *repo) Update(ctx context.Context, account *v1alpha1.Account) error {
	if account.Status.AccountID == 0 {
		return errors.New("cannot update account: status.accountId is missing")
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx,
		"UPDATE account SET email = ?, expansion = ? WHERE id = ?",
		account.Spec.Email, account.Spec.Expansion, account.Status.AccountID,
	)
	if err != nil {
		return fmt.Errorf("failed to update account info: %w", err)
	}

	if account.Spec.GmLevel > 0 {
		_, err = tx.ExecContext(ctx,
			"INSERT INTO account_access (id, gmlevel, RealmID) VALUES (?, ?, -1) "+
				"ON DUPLICATE KEY UPDATE gmlevel = VALUES(gmlevel)",
			account.Status.AccountID, account.Spec.GmLevel,
		)
		if err != nil {
			return fmt.Errorf("failed to update GM access level: %w", err)
		}
	} else {
		_, err = tx.ExecContext(ctx, "DELETE FROM account_access WHERE id = ?", account.Status.AccountID)
		if err != nil {
			return fmt.Errorf("failed to remove GM access level: %w", err)
		}
	}

	return tx.Commit()
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}
