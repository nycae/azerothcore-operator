package account

import (
	"context"

	_ "github.com/go-sql-driver/mysql"
	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Repo   Repository
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// TODO
	return ctrl.Result{}, nil
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
