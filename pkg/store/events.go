package store

import (
	"github.com/lacodon/recoon/pkg/api"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
)

const (
	EventTypeAdd    = "add"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"
)

type Event struct {
	Type           string
	PreviousObject api.Object
	// only contains NamespaceName and VersionKind instead of complete object because otherwise retries will fail forever if outdated object
	ObjectNamespaceName metav1.NamespaceName
	ObjectVersionKind   metav1.VersionKind
}

func (e Event) DeepCopy() Event {
	n := Event{
		Type:                e.Type,
		ObjectNamespaceName: e.ObjectNamespaceName,
		ObjectVersionKind:   e.ObjectVersionKind,
	}

	if e.PreviousObject != nil {
		n.PreviousObject = e.PreviousObject.DeepCopy()
	}

	return n
}
