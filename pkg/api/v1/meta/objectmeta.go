package metav1

type ObjectMeta struct {
	Name             string `json:"name"`
	Namespace        string `json:"namespace"`
	RessourceVersion int64  `json:"ressourceVersion"`
}

func (o ObjectMeta) GetName() string {
	return o.Name
}

func (o ObjectMeta) GetNamespace() string {
	return o.Namespace
}

func (o ObjectMeta) GetNamespaceName() NamespaceName {
	return NamespaceName{
		Namespace: o.Namespace,
		Name:      o.Name,
	}
}

func (o ObjectMeta) GetRessourceVersion() int64 {
	return o.RessourceVersion
}

func (o ObjectMeta) DeepCopy() ObjectMeta {
	return ObjectMeta{
		Name:             o.Name,
		Namespace:        o.Namespace,
		RessourceVersion: o.RessourceVersion,
	}
}
