package realm

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

type Reconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	realm := &v1alpha1.AzerothRealm{}
	if err := r.Get(ctx, req.NamespacedName, realm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	slog.Info("reconciliation in progress", "realm", realm.Name,
		"expansion", realm.Spec.Expansion, "type", realm.Spec.RealmType)

	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("%s-worldserver", realm.Name),
			Namespace: realm.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(ctx, r.Client, deployment, func() error {
		replicas := int32(1)
		if realm.Spec.Replicas != nil {
			replicas = *realm.Spec.Replicas
		}

		labels := map[string]string{"app": "azerothcore-realm", "realm": realm.Name}

		deployment.Labels = labels
		deployment.Spec.Replicas = &replicas
		deployment.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		deployment.Spec.Template.ObjectMeta.Labels = labels

		deployment.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:      "worldserver",
				Image:     "nginx:alpine",
				Resources: realm.Spec.WorldServer.Resources,
			},
		}

		return controllerutil.SetControllerReference(realm, deployment, r.Scheme)
	})

	if err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.AzerothRealm{}).
		Owns(&appsv1.Deployment{}).
		Complete(r)
}
