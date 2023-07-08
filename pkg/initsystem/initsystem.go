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
func InitSSHKeys(sshConfig config.Getter) error {
	sshKeyDir := sshConfig.GetString("keyDir")

	privKeyFilePath := filepath.Join(sshKeyDir, sshauth.PrivateKeyFile)
	pubKeyFilePath := filepath.Join(sshKeyDir, sshauth.PublicKeyFile)
	if err := sshauth.CreateKeypairIfNotExists(privKeyFilePath, pubKeyFilePath, false); err != nil {
		return errors.WithMessage(err, "failed to generate server key pair")
	}

	clientPrivKeyFilePath := filepath.Join(sshKeyDir, sshauth.ClientPrivateKeyFile)
	clientPubKeyFilePath := filepath.Join(sshKeyDir, sshauth.ClientPublicKeyFile)
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
func InitTLS(sshConfig config.Getter) error {
	return sshauth.CreateCertFilesIfNotExist(sshConfig.GetString("host"), sshConfig.GetString("keyDir"), false)
}

// InitClientConfig generates the client config json
func InitClientConfig(sshConfig, uiConfig config.Getter) error {
	hostBase := fmt.Sprintf("https://%s:%d/api/v1", uiConfig.GetString("host"), uiConfig.GetInt("port"))

	sshKeyDir := sshConfig.GetString("keyDir")

	serverCert, err := os.ReadFile(filepath.Join(sshKeyDir, sshauth.ServerCertFile))
	if err != nil {
		return errors.WithMessage(err, "failed to load server cert")
	}

	clientCert, err := tls.LoadX509KeyPair(filepath.Join(sshKeyDir, sshauth.ClientCertFile), filepath.Join(sshKeyDir, sshauth.ClientPrivateKeyFile))
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

	return os.WriteFile(filepath.Join(sshKeyDir, "client.json"), configOut, 0600)
}
