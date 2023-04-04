package metav1

type TypeMeta struct {
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (t TypeMeta) GetVersionKind() VersionKind {
	return VersionKind{
		Version: t.Version,
		Kind:    t.Kind,
	}
}

func (t TypeMeta) DeepCopy() TypeMeta {
	return TypeMeta{
		Version: t.Version,
		Kind:    t.Kind,
	}
}
