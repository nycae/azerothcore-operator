package realm

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	"github.com/nycae/azerothcore-operator/pkg/persistence"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const realmFinalizer = "realm.nycae.io/finalizer"

type Reconciler struct {
	client.Client
	Repo   Repository
	Scheme *runtime.Scheme
}

func (r *Reconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var realm v1alpha1.Realm
	if err := r.Get(ctx, req.NamespacedName, &realm); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	slog.Info("reconciliation in progress", "realm", realm.Name,
		"expansion", realm.Spec.Build, "type", realm.Spec.RealmType)

	if !realm.DeletionTimestamp.IsZero() {
		return r.handleDelete(ctx, &realm)
	}

	if !controllerutil.ContainsFinalizer(&realm, realmFinalizer) {
		controllerutil.AddFinalizer(&realm, realmFinalizer)
		if err := r.Update(ctx, &realm); err != nil {
			return ctrl.Result{}, err
		}
	}

	if err := r.ensureDatabase(ctx, &realm); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure database: %w", err)
	}

	cmName, err := r.ensureConfigMap(ctx, &realm)
	if err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure init conf: %w", err)
	}

	if err := r.ensureStatefulSet(ctx, cmName, &realm); err != nil {
		return ctrl.Result{}, fmt.Errorf("failed to ensure stateful set: %w", err)
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) ensureService(ctx context.Context, realm *v1alpha1.Realm) error {
	labels := map[string]string{
		"app.kubernetes.io/name":       "worldserver",
		"app.kubernetes.io/instance":   realm.Name,
		"app.kubernetes.io/component":  "game-server",
		"app.kubernetes.io/managed-by": "azerothcore-operator",
	}
	headlessSvcName := fmt.Sprintf("%s-headless", realm.Name)
	headlessSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      headlessSvcName,
			Namespace: realm.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, headlessSvc, func() error {
		if headlessSvc.Labels == nil {
			headlessSvc.Labels = make(map[string]string)
		}
		for k, v := range labels {
			headlessSvc.Labels[k] = v
		}
		headlessSvc.Spec.ClusterIP = "None"
		headlessSvc.Spec.Selector = labels
		headlessSvc.Spec.Ports = []corev1.ServicePort{
			{
				Name:     "worldserver",
				Port:     8085,
				Protocol: corev1.ProtocolTCP,
			},
		}
		return controllerutil.SetControllerReference(realm, headlessSvc, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to ensure headless service: %w", err)
	}

	port := realm.Spec.Routing.Port
	if port == 0 {
		port = 8085
	}

	svcName := fmt.Sprintf("%s-worldserver", realm.Name)
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      svcName,
			Namespace: realm.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, svc, func() error {
		if svc.Labels == nil {
			svc.Labels = make(map[string]string)
		}
		for k, v := range labels {
			svc.Labels[k] = v
		}
		svc.Spec.Selector = labels
		svc.Spec.Ports = []corev1.ServicePort{
			{
				Name:       "worldserver",
				Port:       int32(port),
				TargetPort: intstr.FromInt32(8085),
				Protocol:   corev1.ProtocolTCP,
			},
		}
		svc.Spec.Type = corev1.ServiceTypeLoadBalancer
		return controllerutil.SetControllerReference(realm, svc, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to ensure game service: %w", err)
	}

	return nil
}

func (r *Reconciler) ensureDatabase(ctx context.Context, realm *v1alpha1.Realm) error {
	oldID := realm.Status.RealmID

	if err := r.Repo.Ensure(ctx, realm); err != nil {
		return fmt.Errorf("failed to ensure realm in database: %w", err)
	}

	if realm.Status.RealmID != oldID {
		if err := r.Status().Update(ctx, realm); err != nil {
			return fmt.Errorf("failed to update realm status with realmId: %w", err)
		}
	}

	return nil
}

func (r *Reconciler) ensureStatefulSet(ctx context.Context, cmName string, realm *v1alpha1.Realm) error {
	name := fmt.Sprintf("%s-worldserver", realm.Name)
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: realm.Namespace,
		},
	}
	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, sts, func() error {
		labels := map[string]string{
			"app.kubernetes.io/name":       "worldserver",
			"app.kubernetes.io/instance":   realm.Name,
			"app.kubernetes.io/component":  "game-server",
			"app.kubernetes.io/managed-by": "azerothcore-operator",
		}
		if sts.Labels == nil {
			sts.Labels = make(map[string]string)
		}
		for k, v := range labels {
			sts.Labels[k] = v
		}
		replicas := int32(1)
		sts.Spec.Replicas = &replicas
		sts.Spec.Selector = &metav1.LabelSelector{MatchLabels: labels}
		sts.Spec.ServiceName = fmt.Sprintf("%s-headless", realm.Name)
		sts.Spec.Template.ObjectMeta = metav1.ObjectMeta{Labels: labels}

		sts.Spec.Template.Spec.InitContainers = []corev1.Container{
			{
				Name:  "init-client-data",
				Image: "acore/ac-client-data-init:master",
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "server-data",
						MountPath: "/env/dist/data",
					},
				},
			},
		}

		sts.Spec.Template.Spec.Containers = []corev1.Container{
			{
				Name:      "worldserver",
				Image:     "acore/ac-wotlk-worldserver:master",
				Resources: realm.Spec.WorldServer.Resources,
				VolumeMounts: []corev1.VolumeMount{
					{
						Name:      "config-volume",
						MountPath: "/env/dist/etc/worldserver.conf",
						SubPath:   "worldserver.conf",
					},
					{
						Name:      "server-data",
						MountPath: "/env/dist/data",
					},
				},
			},
		}

		sts.Spec.Template.Spec.Volumes = []corev1.Volume{
			{
				Name: "config-volume",
				VolumeSource: corev1.VolumeSource{
					ConfigMap: &corev1.ConfigMapVolumeSource{
						LocalObjectReference: corev1.LocalObjectReference{
							Name: cmName,
						},
					},
				},
			},
		}

		sts.Spec.VolumeClaimTemplates = []corev1.PersistentVolumeClaim{
			{
				ObjectMeta: metav1.ObjectMeta{
					Name: "server-data",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.VolumeResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: resource.MustParse("10Gi"),
						},
					},
				},
			},
		}
		return controllerutil.SetControllerReference(realm, sts, r.Scheme)
	}); err != nil {
		return fmt.Errorf("failed to ensure statefulset: %w", err)
	}
	return nil
}

func (r *Reconciler) ensureConfigMap(ctx context.Context, realm *v1alpha1.Realm) (string, error) {
	fetchPasswordSecret := func(passRef *v1alpha1.SecretKeySelector) (string, error) {
		ns := passRef.Namespace
		if ns == "" {
			ns = realm.Namespace
		}

		var secret corev1.Secret
		if err := r.Get(ctx, client.ObjectKey{Namespace: ns, Name: passRef.Name}, &secret); err != nil {
			return "", err
		}

		val, ok := secret.Data[passRef.Key]
		if !ok {
			return "", fmt.Errorf("secret %s does not contain key %s", passRef.Name, passRef.Key)
		}

		return string(val), nil
	}

	data := struct {
		RealmID                  int64
		PlayerLimit              int32
		AuthDatabaseConnStr      string
		WorldDatabaseConnStr     string
		CharacterDatabaseConnStr string
	}{
		RealmID:             realm.Status.RealmID,
		PlayerLimit:         realm.Spec.WorldServer.MaxPlayers,
		AuthDatabaseConnStr: persistence.DefaultDatabaseConnectionString(fmt.Sprintf("acore_auth")),
	}

	switch realm.Spec.Database.Strategy {
	case v1alpha1.DatabaseStrategySelfManaged:
		if realm.Spec.Database.WorldDB == nil || realm.Spec.Database.CharacterDB == nil {
			return "", errors.New("incomplete database configuration")
		}

		worldPass, err := fetchPasswordSecret(&realm.Spec.Database.WorldDB.PasswordSecretRef)
		if err != nil {
			return "", fmt.Errorf("failed to fetch password for world DB: %w", err)
		}

		charPass, err := fetchPasswordSecret(&realm.Spec.Database.CharacterDB.PasswordSecretRef)
		if err != nil {
			return "", fmt.Errorf("failed to fetch password for character DB: %w", err)
		}

		data.CharacterDatabaseConnStr = persistence.DatabaseConnectionString(
			realm.Spec.Database.CharacterDB.Username,
			charPass,
			realm.Spec.Database.CharacterDB.Hostname,
			fmt.Sprintf("%d", realm.Spec.Database.CharacterDB.Port),
			realm.Spec.Database.CharacterDB.Database)

		data.WorldDatabaseConnStr = persistence.DatabaseConnectionString(
			realm.Spec.Database.WorldDB.Username,
			worldPass,
			realm.Spec.Database.WorldDB.Hostname,
			fmt.Sprintf("%d", realm.Spec.Database.WorldDB.Port),
			realm.Spec.Database.WorldDB.Database)
	case v1alpha1.DatabaseStrategyAutomatic:
		data.CharacterDatabaseConnStr = persistence.DefaultDatabaseConnectionString(fmt.Sprintf("acore_character_%s", realm.Name))
		data.WorldDatabaseConnStr = persistence.DefaultDatabaseConnectionString(fmt.Sprintf("acore_world_%s", realm.Name))
	}

	var buff bytes.Buffer
	if err := WorldServerConfigTemplate.Execute(&buff, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	cmName := fmt.Sprintf("%s-worldserver-conf", realm.Name)
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cmName,
			Namespace: realm.Namespace,
		},
	}

	if _, err := controllerutil.CreateOrUpdate(ctx, r.Client, cm, func() error {
		if cm.Data == nil {
			cm.Data = make(map[string]string)
		}
		cm.Data["worldserver.conf"] = buff.String()
		return controllerutil.SetControllerReference(realm, cm, r.Scheme)
	}); err != nil {
		return "", err
	}

	return cmName, nil
}

func (r *Reconciler) handleDelete(ctx context.Context, realm *v1alpha1.Realm) (ctrl.Result, error) {
	if controllerutil.ContainsFinalizer(realm, realmFinalizer) {
		slog.Info("deleting realm from database", "realm", realm.Name)

		if err := r.Repo.Delete(ctx, realm); err != nil {
			return ctrl.Result{}, fmt.Errorf("failed to delete realm from database: %w", err)
		}

		controllerutil.RemoveFinalizer(realm, realmFinalizer)
		if err := r.Update(ctx, realm); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *Reconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.Realm{}).
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
