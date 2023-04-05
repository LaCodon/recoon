package initsystem

import (
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
)

// InitSshKeys initializes the SSH keys of recoon
func InitSshKeys() error {
	return sshauth.CreateKeypairIfNotExists(config.Cfg.SSH.KeyDir, false)
}

func InitStore(api *store.DefaultStore) error {
	if err := api.CreateBucket(projectv1.VersionKind.String()); err != nil {
		return err
	}

	if err := api.CreateBucket(repositoryv1.VersionKind.String()); err != nil {
		return err
	}

	return nil
}
