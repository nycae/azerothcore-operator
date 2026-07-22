package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type DatabaseStrategy string

const (
	StrategySelfManaged DatabaseStrategy = "SelfManaged"
	StrategyPlainText   DatabaseStrategy = "PlainText"
	StrategyUserSecret  DatabaseStrategy = "UserSecret"
	StrategyCertSecret  DatabaseStrategy = "CertSecret"
)

type ServerType string

const (
	ServerTypePvP   ServerType = "PvP"
	ServerTypePvE   ServerType = "PvE"
	ServerTypeRp    ServerType = "Rp"
	ServerTypeRpPvP ServerType = "RpPvP"
)

type WorldServerSpec struct {
	MaxPlayers int32                       `json:"maxPlayers,omitempty"`
	Resources  corev1.ResourceRequirements `json:"resources,omitempty"`
}

type RealmSpec struct {
	RealmType   ServerType       `json:"realmType"`
	Expansion   string           `json:"expansion"`
	Replicas    *int32           `json:"replicas,omitempty"`
	Database    DatabaseStrategy `json:"database"`
	WorldServer WorldServerSpec  `json:"worldServer,omitempty"`
}

type RealmStatus struct {
	Ready bool `json:"ready,omitempty"`
}

type Realm struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RealmSpec   `json:"spec,omitempty"`
	Status RealmStatus `json:"status,omitempty"`
}

type RealmList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Realm `json:"items"`
}

func (in *Realm) DeepCopyObject() runtime.Object {
	return new(*in)
}

func (in *RealmList) DeepCopyObject() runtime.Object {
	out := *in
	if in.Items != nil {
		out.Items = make([]Realm, len(in.Items))
		copy(out.Items, in.Items)
	}
	return &out
}
