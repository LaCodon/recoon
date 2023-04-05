package main

import (
	"github.com/lacodon/recoon/pkg/initsystem"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func initRecoon(api *store.DefaultStore) error {
	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)

	if err := initsystem.InitSshKeys(); err != nil {
		return errors.WithMessage(err, "failed to init ssh keys")
	}

	if err := initsystem.InitStore(api); err != nil {
		return errors.WithMessage(err, "failed to init store")
	}

	return nil
}
