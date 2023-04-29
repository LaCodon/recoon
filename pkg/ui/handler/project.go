package handler

import (
	"github.com/labstack/echo/v4"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"net/http"
)

func ProjectList(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		list, err := api.List(projectv1.VersionKind)
		if err != nil {
			return err
		}

		resp := make([]*projectv1.Project, 0, len(list))
		for _, el := range list {
			project := el.(*projectv1.Project)
			if c.Param("namespace") == "" || c.Param("namespace") != "" && project.GetNamespace() == c.Param("namespace") {
				resp = append(resp, project)
			}
		}

		return c.JSON(http.StatusOK, resp)
	}
}

func ProjectGet(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		project := &projectv1.Project{}
		if err := api.Get(metav1.NamespaceName{
			Name:      c.Param("name"),
			Namespace: c.Param("namespace"),
		}, project); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				return c.String(http.StatusNotFound, "not found")
			}

			return err
		}

		return c.JSON(http.StatusOK, project)
	}
}
