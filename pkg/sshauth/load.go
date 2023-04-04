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

// GetPublicKeyOpenSshFormat loads the public key of recoon and returns it in OpenSSH format
func GetPublicKeyOpenSshFormat(path string) (string, error) {
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

	rsaPubKey, err := x509.ParsePKIXPublicKey(pemBlock.Bytes)
	if err != nil {
		return "", err
	}

	sshKey, ok := rsaPubKey.(*ecdsa.PublicKey)
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
