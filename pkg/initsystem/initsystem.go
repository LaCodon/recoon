package initsystem

import (
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
)

// InitSSHKeys initializes the SSH keys of recoon
func InitSSHKeys() error {
	return sshauth.CreateKeypairIfNotExists(config.Cfg.SSH.KeyDir, false)
}

// InitStore creates the bbolt files and inits the buckets
func InitStore(api *store.DefaultStore) error {
	if err := api.CreateBucket(projectv1.VersionKind.String()); err != nil {
		return err
	}

	if err := api.CreateBucket(repositoryv1.VersionKind.String()); err != nil {
		return err
	}

	return nil
}

// InitTLS generates a TLS server and client certificate; requires InitSSHKeys
func InitTLS() error {
	return sshauth.CreateCertFilesIfNotExist(config.Cfg.SSH.Host, config.Cfg.SSH.KeyDir, false)
}
