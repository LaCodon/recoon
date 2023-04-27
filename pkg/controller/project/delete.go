package project

import (
	"context"
	"github.com/lacodon/recoon/pkg/compose"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
)

func (c *Controller) handleProjectDelete(ctx context.Context, event store.Event) error {
	logrus.WithField("project", event.PreviousObject.GetNamespaceName()).Debug("run compose down")

	if err := compose.Down(event.PreviousObject.GetName()); err != nil {
		logrus.WithError(err).WithField("project", event.PreviousObject.GetNamespaceName()).Error("error during docker-compose down")
	}

	return nil
}
