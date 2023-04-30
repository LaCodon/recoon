package main

import (
	"github.com/lacodon/recoon/pkg/client"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"path/filepath"
)

var Version = "dev-build"
var apiClient *client.Client

var rootCmd = &cobra.Command{
	Use:               "recoonctl",
	Short:             "RecoonCtl is a remote ctl tool for recoon",
	Version:           Version,
	RunE:              rootCmdRun,
	PersistentPreRunE: loadConfig,
}

func loadConfig(cmd *cobra.Command, args []string) error {
	clientConfig, err := client.LoadConfig(filepath.Join(config.Cfg.SSH.KeyDir, "client.json"))
	if err != nil {
		logrus.WithError(err).Error("failed to load client config")
		return err
	}

	apiClient = client.New(clientConfig.RecoonHost, clientConfig.GetClientCert(), clientConfig.GetServerCert())

	return nil
}

func rootCmdRun(cmd *cobra.Command, _ []string) error {
	return nil
}
