package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	"github.com/nycae/azerothcore-operator/pkg/realm"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
)

var scheme = runtime.NewScheme()

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	GroupVersion := schema.GroupVersion{Group: "nycae.io", Version: "v1alpha1"}
	scheme.AddKnownTypes(GroupVersion,
		&v1alpha1.AzerothRealm{},
		&v1alpha1.AzerothRealmList{},
	)
	metav1GroupVersion := schema.GroupVersion{Group: "nycae.io", Version: "v1alpha1"}
	scheme.AddKnownTypes(metav1GroupVersion, &v1alpha1.AzerothRealm{})
}

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("unable to start manager: %v", err)
	}

	if err = (&realm.AzerothRealmReconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).
		SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to setup realm controller: %v", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("unable to start manager: %v", err)
	}
}
