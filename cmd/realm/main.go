package main

import (
	"log"
	"log/slog"
	"os"

	"github.com/go-logr/logr"
	"github.com/nycae/azerothcore-operator/api/v1alpha1"
	"github.com/nycae/azerothcore-operator/pkg/account"
	"github.com/nycae/azerothcore-operator/pkg/persistence"
	"github.com/nycae/azerothcore-operator/pkg/realm"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
		&v1alpha1.Realm{},
		&v1alpha1.RealmList{},
		&v1alpha1.Account{},
		&v1alpha1.AccountList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
}

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{Scheme: scheme})
	if err != nil {
		log.Fatalf("unable to start manager: %v", err)
	}

	ctrl.SetLogger(logr.FromSlogHandler(logger.Handler()))
	if err = (&realm.Reconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme()}).
		SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to setup realm controller: %v", err)
	}

	if err = (&account.Reconciler{Client: mgr.GetClient(), Scheme: mgr.GetScheme(), Repo: account.NewRepository(persistence.AuthDB())}).
		SetupWithManager(mgr); err != nil {
		log.Fatalf("unable to setup realm controller: %v", err)
	}

	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		log.Fatalf("unable to start manager: %v", err)
	}
}
