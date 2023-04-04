package metav1

type NamespaceName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (nn NamespaceName) String() string {
	return nn.Namespace + "/" + nn.Name
}
