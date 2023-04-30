package main

import (
	"errors"
	"github.com/spf13/cobra"
)

var logsCmd = &cobra.Command{
	Use:   "logs",
	Short: "Get logs from a container",
	RunE:  logsCmdRun,
}

func init() {
	rootCmd.AddCommand(logsCmd)
}

func logsCmdRun(cmd *cobra.Command, args []string) error {
	if len(args) != 1 {
		return errors.New("must pass container id")
	}

	return apiClient.StreamContainerLogs(args[0])
}
