package ui

import (
	"github.com/labstack/echo/v4"
	"github.com/lacodon/recoon/pkg/ui/handler"
	"net/http"
)

func (u *UI) setupRoutes(e *echo.Echo) {
	e.GET("/", func(c echo.Context) error {
		return c.JSON(http.StatusOK, "ok")
	})

	apiGroup := e.Group("/api/v1")

	repoGroup := apiGroup.Group("/repository")
	repoGroup.GET("", handler.RepositoryList(u.api))
	repoGroup.GET("/:namespace", handler.RepositoryList(u.api))
	repoGroup.GET("/:namespace/:name", handler.RepositoryGet(u.api))

	projectGroup := apiGroup.Group("/project")
	projectGroup.GET("", handler.ProjectList(u.api))
	projectGroup.GET("/:namespace", handler.ProjectList(u.api))
	projectGroup.GET("/:namespace/:name", handler.ProjectGet(u.api))
}
