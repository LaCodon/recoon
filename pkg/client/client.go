package client

import (
	"github.com/go-resty/resty/v2"
	"time"
)

type Client struct {
	client *resty.Client
}

func New(baseUrl string) *Client {
	c := resty.New()
	c.SetBaseURL(baseUrl)
	c.SetRetryCount(3)
	c.SetTimeout(10 * time.Second)

	return &Client{
		client: c,
	}
}
