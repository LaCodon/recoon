package ui

import (
	"context"
	"fmt"
	"github.com/labstack/echo/v4"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
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
	e := echo.New()
	e.HideBanner = true
	u.setupRoutes(e)

	go func() {
		certFile := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.ServerCertFile)
		keyFile := filepath.Join(config.Cfg.SSH.KeyDir, sshauth.PrivateKeyFile)
		_ = e.StartTLS(fmt.Sprintf("0.0.0.0:%d", config.Cfg.UI.Port), certFile, keyFile)
	}()

	<-ctx.Done()
	return e.Shutdown(ctx)
}
