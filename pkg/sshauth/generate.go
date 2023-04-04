package sshauth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
)

const (
	PublicKeyFile  = "public.key"
	PrivateKeyFile = "private.key"
)

// CreateKeypairIfNotExists creates a new SSH key pair if none could be found at the given path or if force is true
func CreateKeypairIfNotExists(path string, force bool) error {
	privKeyFilePath := filepath.Join(path, PrivateKeyFile)
	pubKeyFilePath := filepath.Join(path, PublicKeyFile)

	if !force && isExistent(privKeyFilePath) {
		logrus.Debug("SSH keys already exist")
		return nil
	}

	logrus.Debug("Generating SSH keys")

	privFile, err := os.OpenFile(privKeyFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open private key file: %s", err.Error())
	}
	defer privFile.Close()

	pubFile, err := os.OpenFile(pubKeyFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open public key file: %s", err.Error())
	}
	defer pubFile.Close()

	_ = privFile.Truncate(0)
	_ = pubFile.Truncate(0)

	return generateKeyPair(privFile, pubFile)
}

// generateKeyPair generates an ECDSA key pair and writes them to the given streams
func generateKeyPair(keyFile io.Writer, pubFile io.Writer) error {
	key, err := ecdsa.GenerateKey(elliptic.P384(), rand.Reader)
	if err != nil {
		return err
	}

	x509Encoded, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		return err
	}
	if err := pem.Encode(keyFile, &pem.Block{
		Type:  "EC PRIVATE KEY",
		Bytes: x509Encoded,
	}); err != nil {
		return err
	}

	x509EncodedPub, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		return err
	}
	if err := pem.Encode(pubFile, &pem.Block{
		Type:  "EC PUBLIC KEY",
		Bytes: x509EncodedPub,
	}); err != nil {
		return err
	}

	return nil
}

func isExistent(fPath string) bool {
	fileInfo, err := os.Stat(fPath)
	if os.IsNotExist(err) {
		return false
	}
	return !fileInfo.IsDir()
}
