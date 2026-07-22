package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type SecretKeySelector struct {
	Name string `json:"name"`
	Key  string `json:"key"`
}

type AccountSpec struct {
	Username          string            `json:"username"`
	PasswordSecretRef SecretKeySelector `json:"passwordSecretRef"`
	Email             string            `json:"email,omitempty"`
	GmLevel           int               `json:"gmLevel,omitempty"`
	Expansion         int               `json:"expansion,omitempty"`
}

type AccountStatus struct {
	Phase     string `json:"phase,omitempty"`
	AccountID int64  `json:"accountId,omitempty"`
	Message   string `json:"message,omitempty"`
}

type Account struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   AccountSpec   `json:"spec,omitempty"`
	Status AccountStatus `json:"status,omitempty"`
}

func (in *Account) DeepCopyObject() runtime.Object {
	out := *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return &out
}

type AccountList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Account `json:"items"`
}

func (in *AccountList) DeepCopyObject() runtime.Object {
	out := *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		out.Items = make([]Account, len(in.Items))
		for i := range in.Items {
			out.Items[i] = *in.Items[i].DeepCopyObject().(*Account)
		}
	}
	return &out
}
