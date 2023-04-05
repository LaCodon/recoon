package project

import (
	"context"
	"encoding/json"
	"github.com/lacodon/recoon/pkg/api"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/compose"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func (c *Controller) handleProjectDelete(ctx context.Context, event store.Event) error {
	oldData := event.Object.(*api.GenericObject).Data

	project := &projectv1.Project{}
	if err := json.Unmarshal(oldData, project); err != nil {
		return errors.WithMessage(err, "failed to unmarshal deleted object")
	}

	if project.Spec == nil {
		return nil
	}

	if err := compose.Down(project.Name); err != nil {
		logrus.WithError(err).Error("error during docker-compose down for project " + project.Name)
	}

	return nil
}
