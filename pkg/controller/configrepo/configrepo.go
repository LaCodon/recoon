package configrepo

import (
	"context"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/gitrepo"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

const ConfigRepoName = "config-repo"

type Controller struct {
	cloneURL               string
	branchName             string
	api                    store.GetterSetter
	reconciliationInterval time.Duration
	localGitDir            string
	sshKeyDir              string

	repo gitrepo.GitRepository
}

func NewController(api store.GetterSetter, localGitDir, cloneURL, branchName string, reconciliationInterval time.Duration, sshKeyDir string) *Controller {
	return &Controller{
		cloneURL:               cloneURL,
		branchName:             branchName,
		api:                    api,
		reconciliationInterval: reconciliationInterval,
		localGitDir:            localGitDir,
		sshKeyDir:              sshKeyDir,
	}
}

func (c *Controller) Run(ctx context.Context) error {
	var err error
	c.repo, err = gitrepo.NewGitRepository(ctx, c.localGitDir, c.cloneURL, c.branchName, c.sshKeyDir)
	if err != nil {
		return errors.WithMessage(err, "failed to initialize config repo")
	}

	for {
		if err := c.runOnce(ctx); err != nil {
			logrus.WithError(err).Warn("failed to update config repo")
		}

		timer := time.NewTimer(c.reconciliationInterval)
		select {
		case <-ctx.Done():
			return nil
		case <-timer.C:
			continue
		}
	}
}

func (c *Controller) runOnce(ctx context.Context) error {
	if err := c.repo.Pull(ctx); err != nil {
		return errors.WithMessage(err, "failed to pull config repo")
	}

	apiRepo := &repositoryv1.Repository{}
	if err := c.api.Get(metav1.NamespaceName{
		Name:      ConfigRepoName,
		Namespace: "recoon-system",
	}, apiRepo); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			apiRepo = &repositoryv1.Repository{
				ObjectMeta: metav1.ObjectMeta{
					Name:      ConfigRepoName,
					Namespace: "recoon-system",
				},
				Spec: &repositoryv1.Spec{
					Url:    c.cloneURL,
					Branch: c.branchName,
				},
				Status: &repositoryv1.Status{
					LocalPath:       c.repo.GetLocalPath(),
					CurrentCommitId: c.repo.GetCurrentCommitId(),
				},
			}

			if err := c.api.Create(apiRepo); err != nil {
				return err
			}

			return nil
		} else {
			return err
		}
	}

	if apiRepo.Status.CurrentCommitId != c.repo.GetCurrentCommitId() {
		apiRepo.Spec = &repositoryv1.Spec{
			Url:    c.cloneURL,
			Branch: c.branchName,
		}
		apiRepo.Status = &repositoryv1.Status{
			LocalPath:       c.repo.GetLocalPath(),
			CurrentCommitId: c.repo.GetCurrentCommitId(),
		}

		return c.api.Update(apiRepo)
	}

	return nil
}
