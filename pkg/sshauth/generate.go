package sshauth

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"io"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const (
	PublicKeyFile        = "public.key"
	PrivateKeyFile       = "private.key"
	ServerCertFile       = "server.cert"
	ClientPublicKeyFile  = "client-public.key"
	ClientPrivateKeyFile = "client-private.key"
	ClientCertFile       = "client.cert"
)

// CreateKeypairIfNotExists creates a new SSH key pair if none could be found at the given path or if force is true
func CreateKeypairIfNotExists(privKeyFilePath, pubKeyFilePath string, force bool) error {
	if !force && isExistent(privKeyFilePath) {
		logrus.Debug("SSH keys already exist")
		return nil
	}

	logrus.Debug("Generating SSH keys")

	privFile, err := os.OpenFile(privKeyFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return errors.WithMessage(err, "failed to open private key file")
	}
	defer privFile.Close()

	pubFile, err := os.OpenFile(pubKeyFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return errors.WithMessage(err, "failed to open public key file")
	}
	defer pubFile.Close()

	_ = privFile.Truncate(0)
	_ = pubFile.Truncate(0)

	return generateKeyPair(privFile, pubFile)
}

// CreateCertFilesIfNotExist creates client and server TLS certificates if none could be found at the given path or if force is true
func CreateCertFilesIfNotExist(host string, path string, force bool) error {
	serverCertFilePath := filepath.Join(path, ServerCertFile)
	clientCertFilePath := filepath.Join(path, ClientCertFile)

	if !force && isExistent(serverCertFilePath) {
		logrus.Debug("Certificates already exist")
		return nil
	}

	logrus.Debug("Generating Certificates")

	priv, err := GetPrivateKey(filepath.Join(path, PrivateKeyFile))
	if err != nil {
		return fmt.Errorf("failed to load private key: %s", err)
	}

	clienPriv, err := GetPrivateKey(filepath.Join(path, ClientPrivateKeyFile))
	if err != nil {
		return fmt.Errorf("failed to load client private key: %s", err)
	}

	serverCertFile, err := os.OpenFile(serverCertFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open server cert file: %s", err.Error())
	}
	defer serverCertFile.Close()

	clientCertFile, err := os.OpenFile(clientCertFilePath, os.O_RDWR|os.O_CREATE, 0600)
	if err != nil {
		return fmt.Errorf("failed to open client cert file: %s", err.Error())
	}
	defer clientCertFile.Close()

	_ = serverCertFile.Truncate(0)
	_ = clientCertFile.Truncate(0)

	if err := generateCert(priv, []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth}, strings.Split(host, ","), serverCertFile); err != nil {
		return err
	}

	if err := generateCert(clienPriv, []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}, []string{}, clientCertFile); err != nil {
		return err
	}

	return nil
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

func generateCert(priv *ecdsa.PrivateKey, extKeyUsage []x509.ExtKeyUsage, hosts []string, certFile io.Writer) error {
	keyUsage := x509.KeyUsageDigitalSignature
	notBefore := time.Now()
	notAfter := notBefore.Add(10 * 365 * 24 * time.Hour)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return fmt.Errorf("failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"de.lacodon.recoon"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              keyUsage,
		ExtKeyUsage:           extKeyUsage,
		BasicConstraintsValid: true,
	}

	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			template.IPAddresses = append(template.IPAddresses, ip)
		} else {
			template.DNSNames = append(template.DNSNames, h)
		}
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, priv.Public(), priv)
	if err != nil {
		return fmt.Errorf("failed to create certificate: %v", err)
	}

	if err := pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: derBytes}); err != nil {
		return fmt.Errorf("failed to write data to cert.pem: %v", err)
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
