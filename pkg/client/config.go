package client

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"github.com/pkg/errors"
	"os"
)

type Config struct {
	RecoonHost string `json:"recoonHost"`
	ServerCert string `json:"serverCert"`
	ClientCert string `json:"clientCert"`
	ClientKey  string `json:"clientKey"`
}

func GenerateConfig(apiBase string, serverCert []byte, clientCert tls.Certificate) (*Config, error) {
	if len(clientCert.Certificate) != 1 {
		return nil, errors.New("client cert file must contain exactly one certificate")
	}

	clientPrivBytes, err := x509.MarshalPKCS8PrivateKey(clientCert.PrivateKey)
	if err != nil {
		return nil, errors.WithMessage(err, "unable to marshal client private key")
	}

	serverCertOut := base64.StdEncoding.EncodeToString(serverCert)
	clientCertOut := base64.StdEncoding.EncodeToString(clientCert.Certificate[0])
	clientKeyOut := base64.StdEncoding.EncodeToString(clientPrivBytes)

	return &Config{
		RecoonHost: apiBase,
		ServerCert: serverCertOut,
		ClientCert: clientCertOut,
		ClientKey:  clientKeyOut,
	}, nil
}

func LoadConfig(configPath string) (*Config, error) {
	fileBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, errors.WithMessage(err, "failed to open client config file")
	}

	clientConfig := &Config{}
	if err := json.Unmarshal(fileBytes, clientConfig); err != nil {
		return nil, errors.WithMessage(err, "failed to unmarshal client config")
	}

	return clientConfig, nil
}

func (c *Config) GetServerCert() []byte {
	b, _ := base64.StdEncoding.DecodeString(c.ServerCert)
	return b
}

func (c *Config) GetClientCert() tls.Certificate {
	clientKeyBytes, _ := base64.StdEncoding.DecodeString(c.ClientKey)
	clientCertBytes, _ := base64.StdEncoding.DecodeString(c.ClientCert)

	key, _ := x509.ParsePKCS8PrivateKey(clientKeyBytes)

	return tls.Certificate{
		PrivateKey:  key,
		Certificate: [][]byte{clientCertBytes},
	}
}
