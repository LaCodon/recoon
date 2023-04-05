package puller

import (
	"context"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/gitrepo"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"os"
	"time"
)

type Puller struct {
	api store.GetterSetter
}

func NewPuller(api store.GetterSetter) *Puller {
	return &Puller{
		api: api,
	}
}

func (p *Puller) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-time.After(config.Cfg.AppRepo.ReconciliationIntervall):
			if err := p.runOnce(ctx); err != nil {
				logrus.WithError(err).Warn("failed to update/pull app repositories")
			}
		}
	}
}

func (p *Puller) runOnce(ctx context.Context) error {
	projects, err := p.api.List(projectv1.VersionKind)
	if err != nil {
		return errors.WithMessage(err, "failed to list projects")
	}

	// maps localPath to apiRepos
	repoMap := make(map[string][]*repositoryv1.Repository)
	for _, rawProject := range projects {
		project := rawProject.(*projectv1.Project)

		if project.Spec == nil {
			continue
		}

		repo := &repositoryv1.Repository{}
		if err := p.api.Get(metav1.NamespaceName{
			Name:      project.Spec.Repo.Name,
			Namespace: project.Spec.Repo.Namespace,
		}, repo); err != nil {
			_ = p.api.Delete(projectv1.VersionKind, metav1.NamespaceName{
				Name:      project.Name,
				Namespace: project.Namespace,
			})
			continue
		}

		if repo.Status == nil {
			continue
		}

		if _, ok := repoMap[repo.Status.LocalPath]; !ok {
			repoMap[repo.Status.LocalPath] = make([]*repositoryv1.Repository, 0)
		}

		repoMap[repo.Status.LocalPath] = append(repoMap[repo.Status.LocalPath], repo)
	}

	for _, repos := range repoMap {
		// only pull once but update all api objects
		pullRepo := repos[0]
		localRepo, err := gitrepo.NewGitRepository(ctx, pullRepo.Spec.Url, pullRepo.Spec.Branch)
		if err != nil {
			logrus.WithError(err).Warn("failed to init git repo")
			continue
		}

		if err := localRepo.Pull(ctx); err != nil {
			logrus.WithError(err).Warn("failed to pull repo")
			continue
		}

		for _, repo := range repos {
			repo.Status.CurrentCommitId = localRepo.GetCurrentCommitId()
			if err := p.api.Update(repo); err != nil {
				if errors.Is(err, store.ErrNotFound) {
					// maybe repo has been deleted in the meantime -> remove files on disk
					_ = os.RemoveAll(localRepo.GetLocalPath())
					continue
				}

				logrus.WithError(err).Warn("failed to update repository")
				continue
			}
		}
	}

	return nil
}
