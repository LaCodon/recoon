package main

import (
	"github.com/lacodon/recoon/pkg/initsystem"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func initRecoon() error {
	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)

	if err := initsystem.InitSshKeys(); err != nil {
		return errors.WithMessage(err, "failed to init ssh keys")
	}

	return nil
}
