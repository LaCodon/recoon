package ui

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"path/filepath"
)

type UI struct {
	api store.Getter
}

func New(api store.Getter) *UI {
	return &UI{
		api: api,
	}
}

func (u *UI) Run(ctx context.Context) error {
	s := http.Server{}

	e := echo.New()
	e.HideBanner = true
	u.setupRoutes(e)

	go func() {
		clientCert, err := os.ReadFile(filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ClientCertFile))
		if err != nil {
			logrus.WithError(err).Error("failed to load recoon client cert")
			os.Exit(1)
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(clientCert)

		s.Addr = fmt.Sprintf("0.0.0.0:%d", config.Cfg.UI.Port)
		s.Handler = e
		s.TLSConfig = &tls.Config{
			ClientCAs:  caCertPool,
			ClientAuth: tls.RequireAndVerifyClientCert,
		}

		certFile := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ServerCertFile)
		keyFile := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PrivateKeyFile)
		_ = s.ListenAndServeTLS(certFile, keyFile)
	}()

	<-ctx.Done()
	return s.Shutdown(ctx)
}
