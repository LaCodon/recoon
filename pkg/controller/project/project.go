package project

import (
	"context"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/retry"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/lacodon/recoon/pkg/watcher"
	"github.com/pkg/errors"
)

type Controller struct {
	events <-chan store.Event
	api    store.GetterSetter
}

func NewController(apiWatcher watcher.Watcher, api store.GetterSetter) *Controller {
	events := apiWatcher.Watch(projectv1.VersionKind)

	return &Controller{
		events: events,
		api:    api,
	}
}

func (c *Controller) Run(ctx context.Context) error {
	if err := c.reconcileEveryProject(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-c.events:
			if err := retry.KeepRetrying(ctx, event, c.handleProjectChangeEvent); err != nil {
				return err
			}
		}
	}
}

func (c *Controller) reconcileEveryProject(ctx context.Context) error {
	projectList, err := c.api.List(projectv1.VersionKind)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}

		return err
	}

	for _, p := range projectList {
		project := p.(*projectv1.Project)
		if err := c.api.Update(project); err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) handleProjectChangeEvent(ctx context.Context, event store.Event) error {
	switch event.Type {
	case store.EventTypeAdd:
		fallthrough
	case store.EventTypeUpdate:
		return c.handleProjectCreateUpdate(ctx, event)
	case store.EventTypeDelete:
		return c.handleProjectDelete(ctx, event)
	default:
		panic("unimplemented event type: " + event.Type)
	}
}
