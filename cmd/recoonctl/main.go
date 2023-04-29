package main

import (
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	if err := rootCmd.Execute(); err != nil {
		logrus.WithError(err).Error("failed to run recoonctl")
		os.Exit(1)
	}
}
