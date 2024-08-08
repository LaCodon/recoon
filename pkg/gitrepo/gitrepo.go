package gitrepo

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-billy/v5"
	"github.com/go-git/go-git/v5"
	gitconfig "github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/transport"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/sirupsen/logrus"
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

func NewReadOnlyGitRepository(localPath, sshKeyDir string) (ReadonlyGitRepository, error) {
	repo, err := git.PlainOpen(localPath)
	if err != nil {
		return nil, err
	}

	auth, err := ssh.NewPublicKeysFromFile("git", filepath.Join(sshKeyDir, sshauth.PrivateKeyFile), "")
	if err != nil {
		return nil, err
	}

	return &gitRepo{
		localPath:  localPath,
		repository: repo,
		auth:       auth,
	}, nil
}

func NewGitRepository(ctx context.Context, localDir, cloneUrl, branchName, sshKeyDir string) (GitRepository, error) {
	var progressWriter io.Writer
	if logrus.GetLevel() == logrus.DebugLevel {
		progressWriter = os.Stdout
	}

	destinationPath := MakeLocalPath(localDir, cloneUrl, branchName)

	options := &git.CloneOptions{
		URL:           cloneUrl,
		RemoteName:    "origin",
		ReferenceName: plumbing.NewBranchReferenceName(branchName),
		Progress:      progressWriter,
		SingleBranch:  true,
	}

	if strings.HasPrefix(cloneUrl, "git@") {
		auth, err := ssh.NewPublicKeysFromFile("git", filepath.Join(sshKeyDir, sshauth.PrivateKeyFile), "")
		if err != nil {
			return nil, err
		}

		options.Auth = auth
	}

	logrus.Debug("cloning into ", destinationPath)
	repo, err := git.PlainCloneContext(ctx, destinationPath, false, options)
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
		auth:       options.Auth,
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
		if errors.Is(err, git.ErrRemoteNotFound) {
			if _, err := g.repository.CreateRemote(&gitconfig.RemoteConfig{
				Name:  "origin",
				URLs:  []string{g.url},
				Fetch: []gitconfig.RefSpec{gitconfig.RefSpec(g.url)},
			}); err != nil {
				return err
			}

			return err
		}

		if !errors.Is(err, git.NoErrAlreadyUpToDate) {
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
