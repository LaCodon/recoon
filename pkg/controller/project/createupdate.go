package project

import (
	"context"
	composecli "github.com/compose-spec/compose-go/cli"
	conditionv1 "github.com/lacodon/recoon/pkg/api/v1/condition"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/compose"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"path/filepath"
	"time"
)

func (c *Controller) handleProjectCreateUpdate(ctx context.Context, event store.Event) error {
	project := &projectv1.Project{}
	if err := c.api.Get(event.ObjectNamespaceName, project); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}
		return err
	}

	if project.Spec == nil {
		return nil
	}

	if project.Status == nil {
		project.Status = &projectv1.Status{
			Conditions: make(map[conditionv1.Type]conditionv1.Condition),
		}
	}

	projectContainers, err := compose.Status(ctx, project.Name)
	if err != nil {
		return err
	}

	requireRestart := false
	for _, container := range projectContainers {
		if container.State != "running" {
			requireRestart = true
			break
		}
	}

	if project.Status.LastAppliedCommitId != project.Spec.CommitId {
		requireRestart = true
	}

	if project.Status.ContainerCount != len(projectContainers) {
		requireRestart = true
	}

	if !requireRestart {
		return nil
	}

	project.Status.LastAppliedCommitId = project.Spec.CommitId

	project.Status.Conditions = make(map[conditionv1.Type]conditionv1.Condition)
	if err := compose.Up(project.Name, filepath.Join(project.Spec.LocalPath, project.Spec.ComposePath)); err != nil {
		project.Status.Conditions[projectv1.ConditionFailure] = conditionv1.Condition{
			LastTransitionTime: time.Now(),
			Status:             "failure",
			Message:            err.Error(),
		}

		if err := checkComposeSchema(project); err != nil {
			project.Status.Conditions[projectv1.ConditionSchema] = conditionv1.Condition{
				LastTransitionTime: time.Now(),
				Status:             "invalid",
				Message:            err.Error(),
			}
		}

		logrus.WithError(err).WithField("project", project.Name).Warn("failed to run docker-compose")
	} else {
		project.Status.Conditions[projectv1.ConditionSuccess] = conditionv1.Condition{
			LastTransitionTime: time.Now(),
			Status:             "success",
			Message:            "docker-compose up was successful",
		}
	}

	project.Status.ContainerCount = len(projectContainers)

	logrus.WithField("status", project.Status).Debug("update project")

	if err := c.api.Update(project); err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return err
		}
	}

	return nil
}

func checkComposeSchema(project *projectv1.Project) error {
	workingDir := filepath.Join(project.Spec.LocalPath, project.Spec.ComposePath)
	_, err := composecli.ProjectFromOptions(&composecli.ProjectOptions{
		WorkingDir:  workingDir,
		ConfigPaths: []string{filepath.Join(workingDir, "docker-compose.yml")},
		Environment: make(map[string]string),
		EnvFiles:    []string{},
	})

	return err
}
