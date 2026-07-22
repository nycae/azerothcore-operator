package account

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const accountFinalizer = "account.nycae.io/finalizer"

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Repo   Repository
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	slog.Info("Reconciliation in progress", "Account", req.Name)
	var account v1alpha1.Account
	if err := r.Get(ctx, types.NamespacedName{Name: req.Name}, &account); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if !account.ObjectMeta.DeletionTimestamp.IsZero() {
		return r.handleDelete(ctx, &account)
	}

	if err := r.ensureFinalizer(ctx, &account); err != nil {
		return ctrl.Result{}, err
	}

	return r.handleSync(ctx, &account)
}

func (r *Reconciler) fetchPasswordFromSecret(ctx context.Context, acc *v1alpha1.Account) (string, error) {
	var secret corev1.Secret

	secretNamespace := acc.Spec.PasswordSecretRef.Namespace
	if secretNamespace == "" {
		secretNamespace = "default"
	}

	secretKey := types.NamespacedName{
		Name:      acc.Spec.PasswordSecretRef.Name,
		Namespace: secretNamespace,
	}

	if err := r.Get(ctx, secretKey, &secret); err != nil {
		return "", fmt.Errorf("secret %s/%s not found: %w", secretNamespace, secretKey.Name, err)
	}

	passwordBytes, ok := secret.Data[acc.Spec.PasswordSecretRef.Key]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s", acc.Spec.PasswordSecretRef.Key, secretKey.Name)
	}

	return string(passwordBytes), nil
}

func (r *Reconciler) handleDelete(ctx context.Context, account *v1alpha1.Account) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(account, accountFinalizer) {
		if err := r.Repo.Delete(ctx, account); err != nil {
			return ctrl.Result{}, err
		}
		controllerutil.RemoveFinalizer(account, accountFinalizer)
		if err := r.Update(ctx, account); err != nil {
			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

func (r *Reconciler) handleCreate(ctx context.Context, account *v1alpha1.Account) (ctrl.Result, error) {
	password, err := r.fetchPasswordFromSecret(ctx, account)
	if err != nil {
		r.updateStatus(ctx, account, "Error", 0, err.Error())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}
	id, err := r.Repo.Create(ctx, account, password)
	if err != nil {
		r.updateStatus(ctx, account, "Error", 0, err.Error())
		return ctrl.Result{RequeueAfter: 10 * time.Second}, err
	}
	r.updateStatus(ctx, account, "Ready", id, "account created successfully")
	return ctrl.Result{}, nil
}

func (r *Reconciler) handleUpdate(ctx context.Context, account *v1alpha1.Account) (ctrl.Result, error) {
	if err := r.Repo.Update(ctx, account); err != nil {
		r.updateStatus(ctx, account, "Error", 0, err.Error())
		return ctrl.Result{}, err
	}
	r.updateStatus(ctx, account, "Ready", account.Status.AccountID, "account synced successfully")
	return ctrl.Result{}, nil
}

func (r *Reconciler) handleSync(ctx context.Context, account *v1alpha1.Account) (ctrl.Result, error) {
	exists, accountID, err := r.Repo.Exists(ctx, account)
	if err != nil {
		return ctrl.Result{}, err
	}

	if !exists {
		return r.handleCreate(ctx, account)
	}

	if account.Status.AccountID == 0 {
		account.Status.AccountID = accountID
	}

	return r.handleUpdate(ctx, account)
}

func (r *Reconciler) ensureFinalizer(ctx context.Context, account *v1alpha1.Account) error {
	if !controllerutil.ContainsFinalizer(account, accountFinalizer) {
		controllerutil.AddFinalizer(account, accountFinalizer)
		return r.Update(ctx, account)
	}
	return nil
}

func (r *Reconciler) updateStatus(ctx context.Context, acc *v1alpha1.Account, phase string, accountID int64, message string) {
	acc.Status.Phase = phase
	acc.Status.AccountID = accountID
	acc.Status.Message = message
	_ = r.Status().Update(ctx, acc)
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Account{}).
		Complete(r)
}
