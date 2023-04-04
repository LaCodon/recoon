package metav1

type VersionKind struct {
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (vk VersionKind) String() string {
	return vk.Version + "/" + vk.Kind
}
