package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrus.Info("welcome to recoon")

	logrus.SetLevel(logrus.DebugLevel)

	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("failed to run recoon")
		os.Exit(1)
	}
}
