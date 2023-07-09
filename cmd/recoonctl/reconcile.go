package main

import (
	"fmt"
	"github.com/spf13/cobra"
)

var reconcileCmd = &cobra.Command{
	Use:   "reconcile",
	Short: "Trigger an immediate repository reconciliation",
	RunE:  reconcileCmdRun,
}

func init() {
	rootCmd.AddCommand(reconcileCmd)
}

func reconcileCmdRun(cmd *cobra.Command, args []string) error {
	if err := apiClient.Reconcile(); err != nil {
		return err
	}

	fmt.Println("OK")
	return nil
}
