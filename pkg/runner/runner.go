package runner

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"sync"
)

type Runner struct {
	tasks []taskWithConfig

	taskWG       sync.WaitGroup
	childContext context.Context
	childCancel  context.CancelFunc
}

type Task interface {
	Run(ctx context.Context) error
}

type taskWithConfig struct {
	name string
	task Task
}

func (r *Runner) AddTask(task Task) {
	if r.tasks == nil {
		r.tasks = make([]taskWithConfig, 0, 1)
	}

	r.tasks = append(r.tasks, taskWithConfig{
		task: task,
		name: fmt.Sprintf("%T", task),
	})
}

func (r *Runner) StartAll(ctx context.Context) {
	r.childContext, r.childCancel = context.WithCancel(ctx)

	for _, taskConfig := range r.tasks {
		r.taskWG.Add(1)
		go func(ctx context.Context, wg *sync.WaitGroup, taskConfig taskWithConfig, cancel context.CancelFunc) {
			defer wg.Done()

			if err := taskConfig.task.Run(ctx); err != nil {
				cancel()
				if !errors.Is(err, context.Canceled) {
					logrus.WithError(err).WithField("task", taskConfig.name).Error("error during runner execution")
				}
			}
		}(r.childContext, &r.taskWG, taskConfig, r.childCancel)
	}
}

func (r *Runner) Done() <-chan struct{} {
	return r.childContext.Done()
}

func (r *Runner) Wait() {
	r.taskWG.Wait()
}
