package condition

import "time"

type Type string

type Status string

type Condition struct {
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Status             Status    `json:"status"`
	Message            string    `json:"message"`
}

type Conditions map[Type]Condition

func (c Conditions) DeepCopy() Conditions {
	n := make(map[Type]Condition, len(c))

	for typ, cond := range c {
		n[typ] = cond
	}

	return n
}
