package client

import (
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"net/url"
)

func (c *Client) GetRepositories() ([]*repositoryv1.Repository, error) {
	resp, err := c.client.R().SetResult([]*repositoryv1.Repository{}).Get("/repository")
	if err != nil {
		return nil, err
	}

	return *resp.Result().(*[]*repositoryv1.Repository), nil
}

func (c *Client) GetRepository(name string) (*repositoryv1.Repository, error) {
	resp, err := c.client.R().SetResult(&repositoryv1.Repository{}).Get("/repository/default/" + url.PathEscape(name))
	if err != nil {
		return nil, err
	}

	return resp.Result().(*repositoryv1.Repository), nil
}
