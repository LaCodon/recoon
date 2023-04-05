package compose

import (
	"fmt"
	"github.com/go-cmd/cmd"
	"github.com/sirupsen/logrus"
	"strings"
)

func Up(projectName, directory string) error {
	composeCmd := cmd.NewCmd("docker", "compose", "-p", projectName, "up", "-d", "--build", "--quiet-pull", "--remove-orphans")
	composeCmd.Dir = directory
	composeChan := composeCmd.Start()
	finalEvent := <-composeChan
	if finalEvent.Exit != 0 {
		return fmt.Errorf("error during docker-compose up: %s ;;; %s", strings.Join(finalEvent.Stdout, "\n"), strings.Join(finalEvent.Stderr, "\n"))
	}
	logrus.WithField("project", projectName).Debug("successfully ran docker-compose up")
	return nil
}

func Down(projectName string) error {
	composeCmd := cmd.NewCmd("docker", "compose", "-p", projectName, "down", "--remove-orphans", "--rmi", "all")
	composeChan := composeCmd.Start()
	finalEvent := <-composeChan
	if finalEvent.Exit != 0 {
		return fmt.Errorf("error during docker-compose down: %s ;;; %s", strings.Join(finalEvent.Stdout, "\n"), strings.Join(finalEvent.Stderr, "\n"))
	}
	logrus.WithField("project", projectName).Debug("successfully ran docker-compose down")
	return nil
}
