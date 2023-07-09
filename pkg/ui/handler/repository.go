package handler

import (
	"github.com/labstack/echo/v4"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"net/http"
)

func RepositoryList(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		list, err := api.List(repositoryv1.VersionKind)
		if err != nil {
			return err
		}

		resp := make([]*repositoryv1.Repository, 0, len(list))
		for _, el := range list {
			repo := el.(*repositoryv1.Repository)
			if c.Param("namespace") == "" || c.Param("namespace") != "" && repo.GetNamespace() == c.Param("namespace") {
				resp = append(resp, repo)
			}
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func RepositoryGet(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		repo := &repositoryv1.Repository{}
		if err := api.Get(metav1.NamespaceName{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}, repo); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return c.String(http.StatusNotFound, "not found")
			}

			return err
		}

		return c.JSON(http.StatusOK, repo)
	}
}

func RepositoryReconcile(repoReconcileTrigger chan<- bool) echo.HandlerFunc {
	return func(c echo.Context) error {
		repoReconcileTrigger <- true
		return c.JSON(http.StatusOK, nil)
	}
}
