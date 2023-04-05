package repository

import (
	"github.com/lacodon/recoon/pkg/api"
	conditionv1 "github.com/lacodon/recoon/pkg/api/v1/condition"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/schema"
)

var VersionKind = metav1.VersionKind{Version: "v1", Kind: "Repository"}

func init() {
	schema.Register(VersionKind, &Repository{})
}

type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   *Spec   `json:"spec,omitempty"`
	Status *Status `json:"status,omitempty"`
}

type Spec struct {
	// ProjectName of the compose project
	ProjectName string `json:"projectName"`
	// Url for git cloning this repository
	Url string `json:"url,omitempty"`
	// Branch which should be reconciled
	Branch string `json:"branch,omitempty"`
	// Path where the docker-compose.yml can be found
	Path string `json:"path,omitempty"`
}

type Status struct {
	// Conditions represent the current status of this object
	Conditions conditionv1.Conditions `json:"conditions,omitempty"`

	// LocalPath tells us where this repo lays in the local file system
	LocalPath string `json:"localPath,omitempty"`
	// CurrentCommitId is the id of the currently checked out git commit
	CurrentCommitId string `json:"currentCommitId,omitempty"`
}

func (r *Repository) DeepCopy() api.Object {
	n := &Repository{
		TypeMeta:   r.TypeMeta.DeepCopy(),
		ObjectMeta: r.ObjectMeta.DeepCopy(),
	}

	if r.Spec != nil {
		n.Spec = &Spec{
			ProjectName: r.Spec.ProjectName,
			Url:         r.Spec.Url,
			Branch:      r.Spec.Branch,
			Path:        r.Spec.Path,
		}
	}

	if r.Status != nil {
		n.Status = &Status{
			Conditions:      r.Status.Conditions.DeepCopy(),
			LocalPath:       r.Status.LocalPath,
			CurrentCommitId: r.Status.CurrentCommitId,
		}
	}

	return n
}
