package compose

import (
	"context"
	"fmt"
	dockertypes "github.com/docker/docker/api/types"
	dockerfilters "github.com/docker/docker/api/types/filters"
	dockerclient "github.com/docker/docker/client"
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

func Status(ctx context.Context, projectName string) ([]dockertypes.Container, error) {
	client, err := dockerclient.NewClientWithOpts(dockerclient.FromEnv, dockerclient.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker socket: %s", err.Error())
	}

	return client.ContainerList(ctx, dockertypes.ContainerListOptions{
		All:     true,
		Filters: dockerfilters.NewArgs(dockerfilters.Arg("label", fmt.Sprintf("com.docker.compose.project=%s", projectName))),
	})
}
