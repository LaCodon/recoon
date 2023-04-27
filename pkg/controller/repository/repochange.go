package repository

import (
	"context"
	"encoding/json"
	"github.com/lacodon/recoon/pkg/api"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/gitrepo"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"os"
	"strings"
)

func (c *Controller) handleRepoChangeEvent(ctx context.Context, event store.Event) error {
	switch event.Type {
	case store.EventTypeAdd:
		return c.handleRepoCreate(ctx, event)
	case store.EventTypeUpdate:
		return c.handleRepoUpdate(ctx, event)
	case store.EventTypeDelete:
		return c.handleRepoDelete(event)
	default:
		panic("unimplemented repo event type: " + event.Type)
	}
}

func (c *Controller) handleRepoCreate(ctx context.Context, event store.Event) error {
	apiRepo := &repositoryv1.Repository{}
	if err := c.api.Get(event.ObjectNamespaceName, apiRepo); err != nil {
		return err
	}

	if apiRepo.Spec == nil {
		return nil
	}

	repo, err := gitrepo.NewGitRepository(ctx, apiRepo.Spec.Url, apiRepo.Spec.Branch)
	if err != nil {
		return errors.WithMessage(err, "failed to create app repo")
	}

	if err := repo.Pull(ctx); err != nil {
		return errors.WithMessage(err, "failed to pull app repo")
	}

	oldCommitId := ""
	if apiRepo.Status != nil {
		oldCommitId = apiRepo.Status.CurrentCommitId
	}

	if oldCommitId != repo.GetCurrentCommitId() {
		apiRepo.Status = &repositoryv1.Status{
			LocalPath:       repo.GetLocalPath(),
			CurrentCommitId: repo.GetCurrentCommitId(),
		}

		if err := c.api.Update(apiRepo); err != nil {
			return errors.WithMessage(err, "failed to update app repo")
		}
	}

	return nil
}

func (c *Controller) handleRepoUpdate(ctx context.Context, event store.Event) error {
	apiRepo := &repositoryv1.Repository{}
	if err := c.api.Get(event.ObjectNamespaceName, apiRepo); err != nil {
		return err
	}

	if apiRepo.Spec == nil {
		return nil
	}

	if apiRepo.Status == nil {
		return c.handleRepoCreate(ctx, event)
	}

	project := &projectv1.Project{}
	if err := c.api.Get(metav1.NamespaceName{
		Name:      apiRepo.Spec.ProjectName,
		Namespace: "project-" + apiRepo.Spec.ProjectName,
	}, project); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			project = &projectv1.Project{
				ObjectMeta: metav1.ObjectMeta{
					Name:      apiRepo.Spec.ProjectName,
					Namespace: "project-" + apiRepo.Spec.ProjectName,
				},
				Spec: &projectv1.Spec{
					LocalPath:   apiRepo.Status.LocalPath,
					CommitId:    apiRepo.Status.CurrentCommitId,
					ComposePath: apiRepo.Spec.Path,
					Repo: metav1.ObjectRef{
						Version:   apiRepo.Version,
						Kind:      apiRepo.Kind,
						Namespace: apiRepo.Namespace,
						Name:      apiRepo.Name,
					},
				},
			}
			return c.api.Create(project)
		} else {
			return errors.WithMessage(err, "failed to get project")
		}
	}

	if project.Spec.CommitId != apiRepo.Status.CurrentCommitId {
		project.Spec.CommitId = apiRepo.Status.CurrentCommitId
		project.Spec.ComposePath = apiRepo.Spec.Path
		if err := c.api.Update(project); err != nil {
			return errors.WithMessage(err, "failed to update project")
		}
	}

	return nil
}

func (c *Controller) handleRepoDelete(event store.Event) error {
	oldData := event.PreviousObject.(*api.GenericObject).Data

	oldRepo := &repositoryv1.Repository{}
	if err := json.Unmarshal(oldData, oldRepo); err != nil {
		return errors.WithMessage(err, "failed to unmarshal deleted repo")
	}

	if oldRepo.Status == nil {
		return nil
	}

	// don't remove repo folder if there are other repos with same local path
	parts := strings.Split(oldRepo.GetName(), ":")
	prefix := strings.Join(parts[:len(parts)-1], ":")

	repoList, err := c.api.List(repositoryv1.VersionKind, store.InNamespace("default"), store.WithNamePrefix(prefix))
	if err != nil {
		return errors.WithMessage(err, "failed to list repos")
	}

	if len(repoList) == 0 {
		return os.RemoveAll(oldRepo.Status.LocalPath)
	}

	if oldRepo.Spec == nil {
		return nil
	}

	_ = c.api.Delete(projectv1.VersionKind, metav1.NamespaceName{
		Name:      oldRepo.Spec.ProjectName,
		Namespace: "project-" + oldRepo.Spec.ProjectName,
	})

	return nil
}
