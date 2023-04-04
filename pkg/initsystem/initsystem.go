package initsystem

import (
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/sshauth"
)

// InitSshKeys initializes the SSH keys of recoon
func InitSshKeys() error {
	return sshauth.CreateKeypairIfNotExists(config.Cfg.SSH.KeyDir, false)
}
