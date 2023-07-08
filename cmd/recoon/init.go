package main

import (
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/initsystem"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func initRecoon(api *store.DefaultStore, config config.Getter) error {
	formatter := new(logrus.TextFormatter)
	formatter.FullTimestamp = true
	logrus.SetFormatter(formatter)

	if err := initsystem.InitSSHKeys(config.Sub("ssh")); err != nil {
		return errors.WithMessage(err, "failed to init ssh keys")
	}

	if err := initsystem.InitTLS(config.Sub("ssh")); err != nil {
		return errors.WithMessage(err, "failed to init TLS certs")
	}

	if err := initsystem.InitClientConfig(config.Sub("ssh"), config.Sub("ui")); err != nil {
		return errors.WithMessage(err, "failed to create client config json")
	}

	if err := initsystem.InitStore(api); err != nil {
		return errors.WithMessage(err, "failed to init store")
	}

	return nil
}
