package store

import "github.com/lacodon/recoon/pkg/api"

const (
	EventTypeAdd    = "add"
	EventTypeUpdate = "update"
	EventTypeDelete = "delete"
)

type Event struct {
	Type           string
	PreviousObject api.Object
	Object         api.Object
}

func (e Event) DeepCopy() Event {
	n := Event{
		Type: e.Type,
	}

	if e.PreviousObject != nil {
		n.PreviousObject = e.PreviousObject.DeepCopy()
	}

	if e.Object != nil {
		n.Object = e.Object.DeepCopy()
	}

	return n
}
