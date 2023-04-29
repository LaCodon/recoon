package client

import "github.com/go-resty/resty/v2"

type Client struct {
	client *resty.Client
}

func New(baseUrl string) *Client {
	c := resty.New()
	c.SetBaseURL(baseUrl)
	c.SetRetryCount(3)

	return &Client{
		client: c,
	}
}
