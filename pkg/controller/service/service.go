package service

import (
	"context"
)

type Controller struct {
}

func NewController() *Controller {
	return &Controller{}
}

func (c *Controller) Run(ctx context.Context) error {
	return nil
}
