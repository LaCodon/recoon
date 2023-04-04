package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	dockertypes "github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/go-cmd/cmd"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	repositoryv1 "github.com/lacodon/recoon/pkg/api/v1/repository"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/gitrepo"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

type ContainerEvent struct {
	Action  string    `json:"action"`
	ID      string    `json:"id"`
	Service string    `json:"service"`
	Time    time.Time `json:"time"`
	Type    string    `json:"type"`
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	if err := sshauth.CreateKeypairIfNotExists(config.Cfg.SSH.KeyDir, false); err != nil {
		logrus.WithError(err).Fatal("failed to create SSH credentials")
	}

	cloneUrl := "git@github.com:LaCodon/recoon-test.git"
	branchName := "test"
	repo, err := gitrepo.NewGitRepository(context.Background(), cloneUrl, branchName)
	if err != nil {
		logrus.WithError(err).Fatal("failed to open git repo")
	}

	if err := repo.Pull(context.Background()); err != nil {
		logrus.WithError(err).Fatal("failed to pull repo")
	}

	logrus.Info(repo.GetCurrentCommitId())

	fs, _ := repo.GetFS()
	file, _ := fs.Open(".recoon.config.yml")
	data, _ := io.ReadAll(file)
	logrus.Debug(string(data))
	_ = file.Close()

	api, _ := store.NewDefaultStore()
	_ = api.Set(&repositoryv1.Repository{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "config-repo",
			Namespace: "recoon-system",
		},
		Spec: &repositoryv1.Spec{
			Url:    cloneUrl,
			Branch: branchName,
		},
		Status: &repositoryv1.Status{
			LocalPath:       repo.GetLocalPath(),
			CurrentCommitId: repo.GetCurrentCommitId(),
		},
	})
}

func _main() {
	logrus.SetLevel(logrus.DebugLevel)

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	s := make(chan os.Signal)
	signal.Notify(s, os.Interrupt, syscall.SIGTERM)

	project, err := loader.Load(types.ConfigDetails{
		Version:    "3",
		WorkingDir: "./test",
		ConfigFiles: []types.ConfigFile{{
			Filename: "./test/docker-compose.yml",
		}},
	}, func(options *loader.Options) {
		options.ResolvePaths = true
	})
	if err != nil {
		logrus.WithError(err).Fatalln("failed to load docker-compose.yml")
	}

	if err := loader.Normalize(project, true); err != nil {
		logrus.WithError(err).Fatalln("failed to normalize project")
	}

	svcs, err := project.GetServices()
	if err != nil {
		logrus.WithError(err).Fatalln("failed to get services")
	}

	for _, svc := range svcs {
		logrus.Println("service", svc.Name)
	}

	workingDir, err := os.Getwd()
	if err != nil {
		logrus.WithError(err).Fatalln("failed to get current working directory")
	}

	if err := runComposeUp(workingDir); err != nil {
		logrus.WithError(err).Fatalln("failed to run docker-compose")
	}

	//go collectEvents(ctx, workingDir)
	go dockerEvents(ctx, workingDir)

	<-s
	cancel()
	time.Sleep(500 * time.Millisecond)
}

func runComposeUp(workingDir string) error {
	composeCmd := cmd.NewCmd("docker", "compose", "up", "-d", "--build", "--quiet-pull", "--remove-orphans")
	composeCmd.Dir = filepath.Join(workingDir, "./test")
	composeChan := composeCmd.Start()
	finalEvent := <-composeChan
	if finalEvent.Exit != 0 {
		return fmt.Errorf("%#v", finalEvent.Stdout)
	}
	logrus.Println("successfully ran docker-compose up")
	return nil
}

func runDockerStart(containerId string) error {
	startCmd := cmd.NewCmd("docker", "start", containerId)
	composeChan := startCmd.Start()
	finalEvent := <-composeChan
	if finalEvent.Exit != 0 {
		return fmt.Errorf("%#v", finalEvent.Stderr)
	}
	logrus.WithField("containerId", containerId).Debug("restarted container")
	return nil
}

func dockerEvents(ctx context.Context, workingDir string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logrus.WithError(err).Error("failed to connect to docker daemon")
		return
	}

	eventsCh, errCh := cli.Events(ctx, dockertypes.EventsOptions{})

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventsCh:
			if event.Type != "container" {
				continue
			}

			logrus.
				WithField("type", event.Type).
				WithField("action", event.Action).
				WithField("actor", event.Actor.ID).
				WithField("attributes", event.Actor.Attributes).
				Info("new event")

			switch event.Action {
			case "stop":
				// container stopped
				if err := runDockerStart(event.ID); err != nil {
					logrus.WithError(err).Error("failed to restart container")
				}
			case "destroy":
				// container deleted
				if err := runComposeUp(workingDir); err != nil {
					logrus.WithError(err).Error("failed to reconcile project")
				}
			}
		case err := <-errCh:
			logrus.WithError(err).Error("got error from docker events channel")
			time.Sleep(2 * time.Second)
			eventsCh, errCh = cli.Events(ctx, dockertypes.EventsOptions{})
		}
	}
}

func collectEvents(ctx context.Context, workingDir string) {
	eventsCmd := cmd.NewCmdOptions(cmd.Options{
		Streaming: true,
	}, "docker", "compose", "events", "--json")
	eventsCmd.Dir = filepath.Join(workingDir, "./test")
	cmdStatus := eventsCmd.Start()

	go func() {
		<-ctx.Done()
		_ = eventsCmd.Stop()
		logrus.Println("stopped event streaming process")
	}()

	go func() {
		for rawEvent := range eventsCmd.Stdout {
			event := &ContainerEvent{}
			if err := json.Unmarshal([]byte(rawEvent), event); err != nil {
				logrus.WithError(err).Errorln("missed event because of invalid json format")
				continue
			}

			if event.Type != "container" {
				continue
			}

			logrus.
				WithField("action", event.Action).
				WithField("service", event.Service).
				Println("container event")

			switch event.Action {
			case "stop":
				// container stopped
				if err := runDockerStart(event.ID); err != nil {
					logrus.WithError(err).Error("failed to restart container")
				}
			case "destroy":
				// container deleted
				if err := runComposeUp(workingDir); err != nil {
					logrus.WithError(err).Error("failed to reconcile project")
				}
			}
		}
	}()

	<-cmdStatus
	logrus.Println("done printing events")
}
