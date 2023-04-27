package event

import (
	"context"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	dockerclient "github.com/docker/docker/client"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Controller struct {
	api store.GetterSetter
}

func NewController(api store.GetterSetter) *Controller {
	return &Controller{
		api: api,
	}
}

func (c *Controller) Run(ctx context.Context) error {
	cli, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return fmt.Errorf("failed to connect to docker daemon: %s", err.Error())
	}

	eventsCh, errCh := cli.Events(ctx, dockertypes.EventsOptions{})

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-eventsCh:
			if event.Type != "container" {
				continue
			}

			logrus.
				WithField("type", event.Type).
				WithField("action", event.Action).
				WithField("actor", event.Actor.ID).
				Debug("new event")

			projectName := event.Actor.Attributes["com.docker.compose.project"]

			switch event.Action {
			case "stop":
				// container stopped
				fallthrough
			case "die":
				// container died
				fallthrough
			case "destroy":
				// container deleted
				if err := c.triggerReconcile(ctx, projectName); err != nil {
					logrus.
						WithError(err).
						WithField("project", projectName).
						Errorln("failed to trigger project reconciliation")
					return err
				}
			}
		case err := <-errCh:
			logrus.WithError(err).Error("got error from docker events channel; restart channel")
			time.Sleep(2 * time.Second)
			eventsCh, errCh = cli.Events(ctx, dockertypes.EventsOptions{})
		}
	}
}

func (c *Controller) triggerReconcile(ctx context.Context, projectName string) error {
	project := &projectv1.Project{}
	if err := c.api.Get(metav1.NamespaceName{
		Name:      projectName,
		Namespace: "project-" + projectName,
	}, project); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			// project deleted, don't reconcile
			return nil
		}
		return err
	}

	return c.api.Update(project)
}
