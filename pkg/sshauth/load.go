package sshauth

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
	"path/filepath"
)

// GetPublicKeyOpenSSHFormat loads the public key of recoon and returns it in OpenSSH format
func GetPublicKeyOpenSSHFormat(path string) (string, error) {
	data, err := os.ReadFile(filepath.Join(path, PublicKeyFile))
	if err != nil {
		return "", fmt.Errorf("failed to load public key file: %s", err.Error())
	}

	pemBlock, rest := pem.Decode(data)
	if pemBlock == nil {
		return "", fmt.Errorf("public key has invalid format")
	}
	if len(rest) > 0 {
		return "", fmt.Errorf("public key file contains too much information")
	}

	if pemBlock.Type != "EC PUBLIC KEY" {
		return "", fmt.Errorf("got unsupported key type %q", pemBlock.Type)
	}

	pubKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return "", err
	}

	sshKey, ok := pubKey.(*ecdsa.PublicKey)
	if !ok {
		return "", err
	}

	pub, err := ssh.NewPublicKey(sshKey)
	if err != nil {
		return "", err
	}

	sshPubKey := base64.StdEncoding.EncodeToString(pub.Marshal())

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "recoon-host"
	}

	return fmt.Sprintf("ecdsa-sha2-nistp384 %s recoon@%s", sshPubKey, hostname), nil
}

// GetPrivateKey from file
func GetPrivateKey(path string) (*ecdsa.PrivateKey, error) {
	data, err := os.ReadFile(filepath.Join(path, PrivateKeyFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load private key file: %s", err.Error())
	}

	pemBlock, rest := pem.Decode(data)
	if pemBlock == nil {
		return nil, fmt.Errorf("private key has invalid format")
	}
	if len(rest) > 0 {
		return nil, fmt.Errorf("private key file contains too much information")
	}

	if pemBlock.Type != "EC PRIVATE KEY" {
		return nil, fmt.Errorf("got unsupported key type %q", pemBlock.Type)
	}

	return x509.ParseECPrivateKey(pemBlock.Bytes)
}
