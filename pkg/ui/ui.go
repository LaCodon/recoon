package ui

import (
	"context"
	"github.com/labstack/echo/v4"
	"github.com/lacodon/recoon/pkg/store"
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
		_ = e.Start("0.0.0.0:3680")
	}()

	<-ctx.Done()
	return e.Shutdown(ctx)
}
