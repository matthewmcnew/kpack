package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const CustomBuilderKind = "CustomBuilder"

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object,k8s.io/apimachinery/pkg/apis/meta/v1.ObjectMetaAccessor

type CustomBuilder struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CustomBuilderSpec `json:"spec"`
	Status BuilderStatus     `json:"status"`
}

type CustomBuilderSpec struct {
	Tag            string  `json:"tag"`
	Stack          Stack   `json:"stack"`
	Store          Store   `json:"store"`
	Order          []Group `json:"order"`
	ServiceAccount string  `json:"serviceAccount"`
}

type Stack struct {
	BaseBuilderImage string `json:"baseBuilderImage"`
}

type Store struct {
	Image string `json:"image"`
}

type Group struct {
	Group []Buildpack `json:"group"`
}

type Buildpack struct {
	ID string `json:"id"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type CustomBuilderList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []CustomBuilder `json:"items"`
}

func (*CustomBuilder) GetGroupVersionKind() schema.GroupVersionKind {
	return SchemeGroupVersion.WithKind(ClusterBuilderKind)
}
