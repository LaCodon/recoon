package initsystem

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/client"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
)

// InitSSHKeys initializes the SSH keys of recoon
func InitSSHKeys() error {
	privKeyFilePath := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PrivateKeyFile)
	pubKeyFilePath := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PublicKeyFile)
	if err := sshauth.CreateKeypairIfNotExists(privKeyFilePath, pubKeyFilePath, false); err != nil {
		return errors.WithMessage(err, "failed to generate server key pair")
	}

	clientPrivKeyFilePath := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ClientPrivateKeyFile)
	clientPubKeyFilePath := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ClientPublicKeyFile)
	if err := sshauth.CreateKeypairIfNotExists(clientPrivKeyFilePath, clientPubKeyFilePath, false); err != nil {
		return errors.WithMessage(err, "failed to generate client key pair")
	}

	return nil
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

// InitClientConfig generates the client config json
func InitClientConfig() error {
	hostBase := fmt.Sprintf("https://%s:%d/api/v1", config.Cfg.UI.Host, config.Cfg.UI.Port)

	serverCert, err := os.ReadFile(filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ServerCertFile))
	if err != nil {
		return errors.WithMessage(err, "failed to load server cert")
	}

	clientCert, err := tls.LoadX509KeyPair(filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ClientCertFile), filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ClientPrivateKeyFile))
	if err != nil {
		return errors.WithMessage(err, "failed to load client cert")
	}

	clientConfig, err := client.GenerateConfig(hostBase, serverCert, clientCert)
	if err != nil {
		return errors.WithMessage(err, "failed to generate client config json")
	}

	configOut, err := json.Marshal(clientConfig)
	if err != nil {
		return errors.WithMessage(err, "failed to marshal client config json")
	}

	return os.WriteFile(filepath.Join(config.Cfg.SSH.KeyDir, "client.json"), configOut, 0600)
}
