package client

import (
	"fmt"
	projectv1 "github.com/lacodon/recoon/pkg/api/v1/project"
	"net/http"
	"net/url"
)

func (c *Client) GetProjects() ([]*projectv1.Project, error) {
	resp, err := c.client.R().SetResult([]*projectv1.Project{}).Get("/project")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status(), string(resp.Body()))
	}

	return *resp.Result().(*[]*projectv1.Project), nil
}

func (c *Client) GetProject(name string) (*projectv1.Project, error) {
	resp, err := c.client.R().SetResult(&projectv1.Project{}).Get(fmt.Sprintf("/project/project-%s/%s", url.PathEscape(name), url.PathEscape(name)))
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("%s: %s", resp.Status(), string(resp.Body()))
	}

	return resp.Result().(*projectv1.Project), nil
}
