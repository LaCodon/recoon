package client

import (
	"crypto/tls"
	"crypto/x509"
	"github.com/go-resty/resty/v2"
	"net/http"
	"time"
)

type Client struct {
	client *resty.Client
}

func New(baseUrl string, clientCert tls.Certificate, serverCert []byte) *Client {
	c := resty.New()
	c.SetBaseURL(baseUrl)
	c.SetRetryCount(3)
	c.SetTimeout(10 * time.Second)

	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(serverCert)

	c.SetTransport(&http.Transport{
		TLSClientConfig: &tls.Config{
			RootCAs:      caCertPool,
			Certificates: []tls.Certificate{clientCert},
		},
	})

	return &Client{
		client: c,
	}
}
