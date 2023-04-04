package store

import (
	"github.com/lacodon/recoon/pkg/api"
	"github.com/lacodon/recoon/pkg/api/v1/meta"
)

type Getter interface {
	Get(namespaceName metav1.NamespaceName, object api.Object) error
	List(vk metav1.VersionKind, opts ...ListOption) (api.ObjectList, error)
}

type Setter interface {
	Create(object api.Object) error
	Update(object api.Object) error
	Delete(vk metav1.VersionKind, namespaceName metav1.NamespaceName) error
}

type GetterSetter interface {
	Getter
	Setter
}

type listConfig struct {
	Prefix string
}

type ListOption func(cfg *listConfig)

func InNamespace(namespace string) ListOption {
	return func(cfg *listConfig) {
		cfg.Prefix = namespace + "/"
	}
}

func WithNamePrefix(prefix string) ListOption {
	return func(cfg *listConfig) {
		cfg.Prefix += prefix
	}
}
