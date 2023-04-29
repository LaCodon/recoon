package client

import (
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/lacodon/recoon/pkg/ui/handler"
	"github.com/sirupsen/logrus"
	"net/http"
	"net/url"
	"time"
)

func (c *Client) GetContainers(projectName string) ([]*dockertypes.Container, error) {
	suffix := ""
	if projectName != "" {
		suffix = "/" + url.PathEscape(projectName)
	}

	resp, err := c.client.R().SetResult([]*dockertypes.Container{}).Get("/container" + suffix)
	if err != nil {
		return nil, err
	}

	return *resp.Result().(*[]*dockertypes.Container), nil
}

func (c *Client) StreamContainerLogs(containerId string) error {
	since := ""

	for {
		resp, err := c.client.R().SetResult([]handler.LogLine{}).SetQueryParam("since", since).Get("/container/logs/" + containerId)
		if err != nil {
			return err
		}

		if resp.StatusCode() != http.StatusOK {
			logrus.Error("failed to load logs for given container: ", resp.Status())
			return nil
		}

		first := true

		logLines := *resp.Result().(*[]handler.LogLine)
		for _, l := range logLines {
			if first {
				first = false
				continue
			}

			fmt.Printf("[%s] %s\n", l.Timestamp.Format(time.RFC3339), l.Message)
			since = l.Timestamp.Format(time.RFC3339Nano)
		}

		time.Sleep(1 * time.Second)
	}
}
