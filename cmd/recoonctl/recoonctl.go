package main

import (
	"github.com/spf13/cobra"
)

var Version = "dev-build"

var rootCmd = &cobra.Command{
	Use:     "recoonctl",
	Short:   "RecoonCtl is a remote ctl tool for recoon",
	Version: Version,
	RunE:    rootCmdRun,
}

func rootCmdRun(cmd *cobra.Command, _ []string) error {
	return nil
}
