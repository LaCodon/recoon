package metav1

type ObjectRef struct {
	Version   string `json:"version"`
	Kind      string `json:"kind"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

func (o ObjectRef) DeepCopy() ObjectRef {
	return ObjectRef{
		Version:   o.Version,
		Kind:      o.Kind,
		Namespace: o.Namespace,
		Name:      o.Name,
	}
}
