package project

import (
	"github.com/lacodon/recoon/pkg/api"
	conditionv1 "github.com/lacodon/recoon/pkg/api/v1/condition"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/schema"
)

var VersionKind = metav1.VersionKind{Version: "v1", Kind: "Project"}

func init() {
	schema.Register(VersionKind, &Project{})
}

const (
	ConditionSuccess conditionv1.Type = "ComposeSuccess"
	ConditionFailure conditionv1.Type = "ComposeFailure"
	ConditionSchema  conditionv1.Type = "ComposeSchema"
)

type Project struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   *Spec   `json:"spec,omitempty"`
	Status *Status `json:"status,omitempty"`
}

type Spec struct {
	LocalPath   string           `json:"localPath,omitempty"`
	Repo        metav1.ObjectRef `json:"repo,omitempty"`
	CommitId    string           `json:"commitId,omitempty"`
	ComposePath string           `json:"composePath"`
}

type Status struct {
	Conditions          conditionv1.Conditions `json:"conditions,omitempty"`
	LastAppliedCommitId string                 `json:"lastAppliedCommitId"`
	ContainerCount      int                    `json:"containerCount"`
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

	if p.Status != nil {
		n.Status = &Status{
			Conditions:          p.Status.Conditions.DeepCopy(),
			LastAppliedCommitId: p.Status.LastAppliedCommitId,
			ContainerCount:      p.Status.ContainerCount,
		}
	}

	return n
}
