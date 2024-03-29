package main

import (
	"context"
	"github.com/lacodon/recoon/pkg/config"
	"github.com/lacodon/recoon/pkg/controller/configrepo"
	"github.com/lacodon/recoon/pkg/controller/event"
	"github.com/lacodon/recoon/pkg/controller/project"
	"github.com/lacodon/recoon/pkg/controller/repository"
	"github.com/lacodon/recoon/pkg/puller"
	"github.com/lacodon/recoon/pkg/runner"
	"github.com/lacodon/recoon/pkg/sshauth"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/lacodon/recoon/pkg/ui"
	"github.com/lacodon/recoon/pkg/watcher"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"syscall"
)

var Version = "dev-build"

var rootCmd = &cobra.Command{
	Use:     "recoon",
	Short:   "Recoon is a container orchestrator with integrated gitops controller",
	Version: Version,
	RunE:    rootCmdRun,
}

func rootCmdRun(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Setup()
	if err != nil {
		return err
	}

	api, err := store.NewDefaultStore(cfg.GetString("store.databaseFile"))
	if err != nil {
		return err
	}

	if err := initRecoon(api, cfg); err != nil {
		return err
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	logrus.Println(sshauth.GetPublicKeyOpenSSHFormat(cfg.GetString("ssh.keyDir")))

	immediateRepoReconcileTrigger := make(chan bool)

	apiWatcher := watcher.NewDefaultWatcher(api.EventsChan())
	repoPuller := puller.NewPuller(api,
		immediateRepoReconcileTrigger,
		cfg.GetString("store.gitDir"),
		cfg.GetString("ssh.keyDir"),
		cfg.GetDuration("appRepo.reconciliationInterval"))
	repoConfigController := configrepo.NewController(api,
		cfg.GetString("store.gitDir"),
		cfg.GetString("configRepo.cloneURL"),
		cfg.GetString("configRepo.branchName"),
		cfg.GetDuration("configRepo.reconciliationInterval"),
		cfg.GetString("ssh.keyDir"))
	repositoryController := repository.NewController(apiWatcher, api,
		cfg.GetString("store.gitDir"),
		cfg.GetString("ssh.keyDir"))
	projectController := project.NewController(apiWatcher, api)
	eventController := event.NewController(api)
	recoonUI := ui.New(api,
		immediateRepoReconcileTrigger,
		cfg.GetInt("ui.port"),
		cfg.GetString("ssh.keyDir"))

	ctx, cancel := context.WithCancel(cmd.Context())

	// run subsystems
	logrus.Debugf("start all subsystems")
	taskManager := new(runner.Runner)
	taskManager.AddTask(recoonUI)
	taskManager.AddTask(repositoryController)
	taskManager.AddTask(repoConfigController)
	taskManager.AddTask(projectController)
	taskManager.AddTask(eventController)
	taskManager.AddTask(apiWatcher)
	taskManager.AddTask(repoPuller)
	taskManager.StartAll(ctx)

	select {
	case <-taskManager.Done():
		logrus.Error("some runner aborted its execution, shutting down...")
		cancel()
		taskManager.Wait()
		return errors.New("error in runner")
	case <-sigChan:
		logrus.Info("shutting down...")
		cancel()
		taskManager.Wait()
	}

	return nil
}
