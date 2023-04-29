package main

import (
	"fmt"
	"github.com/lacodon/recoon/pkg/client"
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
		return fmt.Errorf("must pass container id")
	}

	c := client.New("http://localhost:3680/api/v1")
	return c.StreamContainerLogs(args[0])
}
