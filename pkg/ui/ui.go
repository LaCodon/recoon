package ui

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

type UI struct {
	api       store.Getter
	port      int
	sshKeyDir string
}

func New(api store.Getter, port int, sshKeyDir string) *UI {
	return &UI{
		api:       api,
		port:      port,
		sshKeyDir: sshKeyDir,
	}
}

func (u *UI) Run(ctx context.Context) error {
	s := http.Server{}

	e := echo.New()
	e.HideBanner = true
	u.setupRoutes(e)

	go func() {
		clientCert, err := os.ReadFile(filepath.Join(u.sshKeyDir, sshauth.ClientCertFile))
		if err != nil {
			logrus.WithError(err).Error("failed to load recoon client cert")
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(clientCert)

		s.Addr = fmt.Sprintf("0.0.0.0:%d", u.port)
		s.Handler = e
		s.TLSConfig = &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}

		certFile := filepath.Join(u.sshKeyDir, sshauth.ServerCertFile)
		keyFile := filepath.Join(u.sshKeyDir, sshauth.PrivateKeyFile)
		_ = s.ListenAndServeTLS(certFile, keyFile)
	}()

	<-ctx.Done()
	return s.Shutdown(ctx)
}
