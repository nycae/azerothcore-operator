package realm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/nycae/azerothcore-operator/api/v1alpha1"
)

type Repository interface {
	Ensure(ctx context.Context, realm *v1alpha1.Realm) error
	Exists(ctx context.Context, realm *v1alpha1.Realm) (bool, int64, error)
	Create(ctx context.Context, realm *v1alpha1.Realm) (int64, error)
	Update(ctx context.Context, realm *v1alpha1.Realm) error
	Delete(ctx context.Context, realm *v1alpha1.Realm) error
}

type repo struct {
	db *sql.DB
}

func (r *repo) Ensure(ctx context.Context, realm *v1alpha1.Realm) error {
	exists, id, err := r.Exists(ctx, realm)
	if err != nil {
		return fmt.Errorf("failed to check realm existence: %w", err)
	}

	if !exists {
		newID, err := r.Create(ctx, realm)
		if err != nil {
			return err
		}
		realm.Status.RealmID = newID
		return nil
	}

	realm.Status.RealmID = id
	return r.Update(ctx, realm)
}

func (r *repo) Exists(ctx context.Context, realm *v1alpha1.Realm) (bool, int64, error) {
	var id int64
	var err error

	if realm.Status.RealmID > 0 {
		err = r.db.QueryRowContext(ctx, `SELECT id FROM realmlist WHERE id = ?`, realm.Status.RealmID).Scan(&id)
	} else {
		err = r.db.QueryRowContext(ctx, `SELECT id FROM realmlist WHERE name = ?`, realm.Name).Scan(&id)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return false, 0, nil
	}
	if err != nil {
		return false, 0, fmt.Errorf("unable to lookup realm: %w", err)
	}

	return true, id, nil
}

func (r *repo) Create(ctx context.Context, realm *v1alpha1.Realm) (int64, error) {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO realmlist 
    	(name, address, localAddress, localSubnetMask, port, icon, flag, timezone, allowedSecurityLevel, gamebuild)
		VALUES
		  (?, ?, ?, '255.255.255.0', ?, ?, ?, ?, ?, ?)`,
		realm.Spec.RealmName,
		realm.Spec.Routing.Address,
		realm.Spec.Routing.LocalAddress,
		realm.Spec.Routing.Port,
		realm.Spec.RealmType.ToInt(),
		realm.Spec.Status.ToInt(),
		realm.Spec.Timezone.ToInt(),
		realm.Spec.GmOnly.ToInt(),
		realm.Spec.Build.Build())
	if err != nil {
		return 0, fmt.Errorf("failed to insert realm: %w", err)
	}
	return res.LastInsertId()
}

func (r *repo) Update(ctx context.Context, realm *v1alpha1.Realm) error {
	if realm.Status.RealmID == 0 {
		return errors.New("cannot update realm: status.realmId is missing")
	}

	if _, err := r.db.ExecContext(ctx, `
		UPDATE realmlist SET
			name = ?,
			address = ?,
			localAddress = ?,
			port = ?,
			icon = ?,
			flag = ?,
			timezone = ?,
			allowedSecurityLevel = ?
		WHERE id = ?`,
		realm.Name,
		realm.Spec.Routing.Address,
		realm.Spec.Routing.LocalAddress,
		realm.Spec.Routing.Port,
		realm.Spec.RealmType.ToInt(),
		realm.Spec.Status.ToInt(),
		realm.Spec.Timezone.ToInt(),
		realm.Spec.GmOnly.ToInt(),
		realm.Status.RealmID,
	); err != nil {
		return fmt.Errorf("failed to update realm: %w", err)
	}

	return nil
}

func (r *repo) Delete(ctx context.Context, realm *v1alpha1.Realm) error {
	realmID := realm.Status.RealmID
	if realmID == 0 {
		exists, id, err := r.Exists(ctx, realm)
		if err != nil {
			return fmt.Errorf("failed to check if realm exists before delete: %w", err)
		}
		if !exists {
			return nil
		}
		realmID = id
	}

	if _, err := r.db.ExecContext(ctx, `DELETE FROM realmlist WHERE id = ?`, realmID); err != nil {
		return fmt.Errorf("failed to delete realm: %w", err)
	}
	return nil
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}
