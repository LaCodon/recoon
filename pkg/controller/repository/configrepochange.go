package repository

import (
	"context"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/gitrepo"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io"
)

type ConfigRepoData struct {
	Repos []ConfigRepoMeta `yaml:"repos"`
}

type ConfigRepoMeta struct {
	Name   string `yaml:"name"`
	URL    string `yaml:"url"`
	Branch string `yaml:"branch"`
	Path   string `yaml:"path"`
}

func (c *Controller) handleConfigRepoChangeEvent(ctx context.Context, event store.Event) error {
	if event.Type == store.EventTypeDelete {
		return errors.New("deleted config-repo object, this should not happen!")
	}

	apiRepo := event.Object.(*repositoryv1.Repository)
	if apiRepo.Spec == nil {
		return nil
	}

	cloneUrl := apiRepo.Spec.Url
	branchName := apiRepo.Spec.Branch

	repo, err := gitrepo.NewReadOnlyGitRepository(gitrepo.MakeLocalPath(cloneUrl, branchName))
	if err != nil {
		return errors.WithMessage(err, "failed to initialize config repo")
	}

	fs, err := repo.GetFS()
	if err != nil {
		return errors.WithMessage(err, "failed to get filesystem of config repo")
	}

	file, err := fs.Open(".recoon.config.yml")
	if err != nil {
		return errors.WithMessage(err, "failed to open .recoon.config.yml")
	}

	data, err := io.ReadAll(file)
	_ = file.Close()
	if err != nil {
		return err
	}

	configRepoData := &ConfigRepoData{}
	if err := yaml.Unmarshal(data, configRepoData); err != nil {
		return errors.WithMessage(err, "failed to unmarshal .recoon.config.yml")
	}

	currentRepos, err := c.api.List(repositoryv1.VersionKind, store.InNamespace("default"))
	if err != nil {
		return errors.WithMessage(err, "failed to list repositories")
	}

	for _, repoMeta := range configRepoData.Repos {
		newRepo := &repositoryv1.Repository{
			TypeMeta: metav1.TypeMeta{
				Version: repositoryv1.VersionKind.Version,
				Kind:    repositoryv1.VersionKind.Kind,
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      gitrepo.MakeAPIName(repoMeta.URL, repoMeta.Branch, repoMeta.Path),
				Namespace: "default",
			},
			Spec: &repositoryv1.Spec{
				ProjectName: repoMeta.Name,
				Url:         repoMeta.URL,
				Branch:      repoMeta.Branch,
				Path:        repoMeta.Path,
			},
		}

		oldIxd := currentRepos.Index(metav1.NamespaceName{
			Name:      newRepo.GetName(),
			Namespace: newRepo.GetNamespace(),
		})

		if oldIxd < 0 {
			if err := c.api.Create(newRepo); err != nil {
				logrus.WithError(err).Warn("failed to create repo")
			}
		} else {
			currentRepos = append(currentRepos[:oldIxd], currentRepos[oldIxd+1:]...)
			// no api update required because if spec changes, name changes as well -> replacement instead of update
		}

	}

	for _, oldRepo := range currentRepos {
		_ = c.api.Delete(oldRepo.GetVersionKind(), oldRepo.GetNamespaceName())
	}

	return nil
}
