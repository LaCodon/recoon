package gitrepo

import (
	"context"
	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

type GitRepository interface {
	Pull(ctx context.Context) error
	GetCurrentCommitId() string
	GetFS() (billy.Filesystem, error)
	GetLocalPath() string
}

type ReadonlyGitRepository interface {
	GetCurrentCommitId() string
	GetFS() (billy.Filesystem, error)
}

type gitRepo struct {
	url        string
	branchName string
	localPath  string
	repository *git.Repository
	auth       transport.AuthMethod
}

func NewReadOnlyGitRepository(localPath string) (ReadonlyGitRepository, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, err
	}

	auth, err := ssh.NewPublicKeysFromFile("git", filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PrivateKeyFile), "")
	if err != nil {
		return nil, err
	}

	return &gitRepo{
		localPath:  localPath,
		repository: repo,
		auth:       auth,
	}, nil
}

func NewGitRepository(ctx context.Context, cloneUrl, branchName string) (GitRepository, error) {
	var progressWriter io.Writer
	if logrus.GetLevel() == logrus.DebugLevel {
		progressWriter = os.Stdout
	}

	auth, err := ssh.NewPublicKeysFromFile("git", filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PrivateKeyFile), "")
	if err != nil {
		return nil, err
	}

	destinationPath := MakeLocalPath(cloneUrl, branchName)
	if err != nil {
		return nil, err
	}

	logrus.Debug("cloning into ", destinationPath)
	repo, err := git.PlainCloneContext(ctx, destinationPath, false, &git.CloneOptions{
		URL:           cloneUrl,
		Auth:          auth,
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		Progress:      progressWriter,
		SingleBranch:  true,
	})
	if err != nil {
		if err == git.ErrRepositoryAlreadyExists {
			repo, err = git.PlainOpen(destinationPath)
			if err != nil {
				_ = os.RemoveAll(destinationPath)
				return nil, err
			}
		} else {
			_ = os.RemoveAll(destinationPath)
			return nil, err
		}
	}

	return &gitRepo{
		url:        cloneUrl,
		branchName: branchName,
		localPath:  destinationPath,
		repository: repo,
		auth:       auth,
	}, nil
}

// Pull pulls all changes from remote
func (g *gitRepo) Pull(ctx context.Context) error {
	var progressWriter io.Writer
	if logrus.GetLevel() == logrus.DebugLevel {
		progressWriter = os.Stdout
	}

	if err := g.repository.FetchContext(ctx, &git.FetchOptions{
		RemoteName: "origin",
		RefSpecs:   []gitconfig.RefSpec{"+refs/heads/*:refs/remotes/origin/*"},
		Auth:       g.auth,
		Progress:   progressWriter,
		Force:      true,
	}); err != nil {
		if err != git.NoErrAlreadyUpToDate {
			return err
		}
	}

	worktree, err := g.repository.Worktree()
	if err != nil {
		return err
	}

	if err := worktree.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewRemoteReferenceName("origin", g.branchName),
		Force:  true,
	}); err != nil {
		return err
	}

	return nil
}

func (g *gitRepo) GetCurrentCommitId() string {
	head, err := g.repository.Head()
	if err != nil {
		return ""
	}

	return head.Hash().String()
}

func (g *gitRepo) GetFS() (billy.Filesystem, error) {
	worktree, err := g.repository.Worktree()
	if err != nil {
		return nil, err
	}

	return worktree.Filesystem, nil
}

func (g *gitRepo) GetLocalPath() string {
	return g.localPath
}
