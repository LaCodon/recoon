package project

import (
	"context"
	composecli "github.com/compose-spec/compose-go/cli"
	conditionv1 "github.com/lacodon/recoon/pkg/api/v1/condition"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/compose"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func (c *Controller) handleProjectCreateUpdate(ctx context.Context, event store.Event) error {
	project := event.Object.(*projectv1.Project)

	if project.Spec == nil {
		return nil
	}

	if project.Status == nil {
		project.Status = &projectv1.Status{
			Conditions: make(map[conditionv1.Type]conditionv1.Condition),
		}
	}

	if project.Status.LastAppliedCommitId == project.Spec.CommitId {
		return nil
	}

	project.Status.LastAppliedCommitId = project.Spec.CommitId

	if err := compose.Up(project.Name, filepath.Join(project.Spec.LocalPath, project.Spec.ComposePath)); err != nil {
		project.Status.Conditions[projectv1.ConditionFailure] = conditionv1.Condition{
			LastTransitionTime: time.Now(),
			Status:             "failure",
			Message:            err.Error(),
		}

		updateProjectFailureConditions(project)

		logrus.WithError(err).WithField("project", project.Name).Warn("failed to run docker-compose")
	} else {
		project.Status.Conditions = make(map[conditionv1.Type]conditionv1.Condition)
		project.Status.Conditions[projectv1.ConditionSuccess] = conditionv1.Condition{
			LastTransitionTime: time.Now(),
			Status:             "success",
			Message:            "docker-compose up was successful",
		}
	}

	return c.api.Update(project)
}

func updateProjectFailureConditions(project *projectv1.Project) {
	workingDir := filepath.Join(project.Spec.LocalPath, project.Spec.ComposePath)
	_, err := composecli.ProjectFromOptions(&composecli.ProjectOptions{
		WorkingDir:  workingDir,
		ConfigPaths: []string{filepath.Join(workingDir, "docker-compose.yml")},
		Environment: make(map[string]string),
		EnvFiles:    []string{},
	})
	if err != nil {
		project.Status.Conditions[projectv1.ConditionSchema] = conditionv1.Condition{
			LastTransitionTime: time.Now(),
			Status:             "invalid",
			Message:            err.Error(),
		}
	}

}