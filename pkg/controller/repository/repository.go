package repository

import (
	"context"
	"fmt"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/controller/configrepo"
	"github.com/lacodon/recoon/pkg/retry"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/lacodon/recoon/pkg/watcher"
	"github.com/pkg/errors"
)

type Controller struct {
	events      <-chan store.Event
	api         store.GetterSetter
	retryer     retry.Retryer
	localGitDir string
	sshKeyDir   string
}

func NewController(apiWatcher watcher.Watcher, api store.GetterSetter, localGitDir, sshKeyDir string) *Controller {
	events := apiWatcher.Watch(repositoryv1.VersionKind)

	return &Controller{
		events:      events,
		api:         api,
		retryer:     retry.New(events),
		localGitDir: localGitDir,
		sshKeyDir:   sshKeyDir,
	}
}

func (c *Controller) Run(ctx context.Context) error {
	// startup reconciliation
	if err := c.reconcileEveryRepo(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		case event := <-c.events:
			if err := c.handleEvent(ctx, event); err != nil {
				return err
			}
		}
	}
}

func (c *Controller) handleEvent(ctx context.Context, event store.Event) error {
	switch event.ObjectVersionKind {
	case repositoryv1.VersionKind:
		if event.ObjectNamespaceName.Name == configrepo.ConfigRepoName && event.ObjectNamespaceName.Namespace == "recoon-system" {
			c.retryer.RetryOnError(ctx, event, c.handleConfigRepoChangeEvent)
			return nil
		} else {
			c.retryer.RetryOnError(ctx, event, c.handleRepoChangeEvent)
			return nil
		}
	default:
		return fmt.Errorf("unknown event object kind: %s/%s", event.ObjectVersionKind, event.ObjectNamespaceName)
	}
}

func (c *Controller) reconcileEveryRepo(ctx context.Context) error {
	repoList, err := c.api.List(repositoryv1.VersionKind)
	if err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return nil
		}

		return err
	}

	for _, r := range repoList {
		repo := r.(*repositoryv1.Repository)
		if err := c.api.Update(repo); err != nil {
			return err
		}
	}

	return nil
}
