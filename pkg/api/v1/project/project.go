package project

import (
	"github.com/lacodon/recoon/pkg/api"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/schema"
)

var VersionKind = metav1.VersionKind{Version: "v1", Kind: "Project"}

func init() {
	schema.Register(VersionKind, &Project{})
}

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec *Spec `json:"spec,omitempty"`
}

type Spec struct {
	LocalPath   string           `json:"localPath,omitempty"`
	Repo        metav1.ObjectRef `json:"repo,omitempty"`
	CommitId    string           `json:"commitId,omitempty"`
	ComposePath string           `json:"composePath"`
}

func (p *Project) DeepCopy() api.Object {
	n := &Project{
		TypeMeta:   p.TypeMeta.DeepCopy(),
		ObjectMeta: p.ObjectMeta.DeepCopy(),
	}

	if p.Spec != nil {
		n.Spec = &Spec{
			LocalPath:   p.Spec.LocalPath,
			Repo:        p.Spec.Repo.DeepCopy(),
			CommitId:    p.Spec.CommitId,
			ComposePath: p.Spec.ComposePath,
		}
	}

	return n
}
