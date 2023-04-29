package handler

import (
	"errors"
	"fmt"
	"github.com/labstack/echo/v4"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"github.com/lacodon/recoon/pkg/compose"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type logLine struct {
	Timestamp time.Time
	Message   string
}

func ContainerList(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		project := &projectv1.Project{}
		if err := api.Get(metav1.NamespaceName{
			Name:      c.Param("project"),
			Namespace: "project-" + c.Param("project"),
		}, project); err != nil {
			if errors.Is(err, store.ErrNotFound) {
				logrus.Warn("adfsdfdf")
				return c.String(http.StatusNotFound, "not found")
			}

			return err
		}

		containers, err := compose.Status(c.Request().Context(), project.GetName())
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, containers)
	}
}

func ContainerGetLogs(api store.Getter) echo.HandlerFunc {
	return func(c echo.Context) error {
		containerId := c.Param("container")
		since := c.QueryParam("since")

		logs, err := compose.Logs(c.Request().Context(), containerId, since, "60")
		if err != nil {
			return err
		}

		rawLines := strings.Split(logs, "\n")
		logLines := make([]logLine, 0, len(rawLines))
		for _, l := range rawLines {
			parts := strings.SplitAfterN(l, "Z", 2)
			if len(parts) != 2 {
				continue
			}

			year := time.Now().Year()
			remainder := strings.Split(parts[0], strconv.Itoa(year))
			if len(remainder) != 2 {
				continue
			}

			rawTs := fmt.Sprintf("%d%s", time.Now().Year(), remainder[1])

			ts, _ := time.Parse(time.RFC3339Nano, rawTs)
			message := parts[1]

			logLines = append(logLines, logLine{
				Timestamp: ts,
				Message:   message[1:],
			})
		}

		return c.JSON(http.StatusOK, logLines)
	}
}
