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

func reconcileCmdRun(_ *cobra.Command, _ []string) error {
	if err := apiClient.Reconcile(); err != nil {
		return err
	}

	fmt.Println("OK")
	return nil
}
