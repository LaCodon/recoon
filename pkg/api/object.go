package api

import (
	"github.com/lacodon/recoon/pkg/api/v1/meta"
)

type Object interface {
	GetName() string
	GetNamespace() string
	GetNamespaceName() metav1.NamespaceName
	GetVersionKind() metav1.VersionKind
	GetRessourceVersion() int64
	DeepCopy() Object
}

type ObjectList []Object

func (o ObjectList) Index(namespaceName metav1.NamespaceName) int {
	for k, obj := range o {
		if obj.GetName() == namespaceName.Name && obj.GetNamespace() == namespaceName.Namespace {
			return k
		}
	}

	return -1
}

type GenericObject struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Data []byte `json:"data,omitempty"`
}

func (e *GenericObject) DeepCopy() Object {
	var d []byte

	if e.Data != nil {
		d = make([]byte, len(e.Data))
		copy(d, e.Data)
	}

	return &GenericObject{
		TypeMeta:   e.TypeMeta.DeepCopy(),
		ObjectMeta: e.ObjectMeta.DeepCopy(),
		Data:       d,
	}
}
