package retry

import (
	"context"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/sirupsen/logrus"
	"time"
)

type Retryable func(ctx context.Context, event store.Event) error

type Retryer interface {
	RetryOnError(ctx context.Context, event store.Event, handler Retryable)
}

type defaultRetryer struct {
	eventChan chan store.Event
}

func New(eventChan chan store.Event) Retryer {
	return &defaultRetryer{
		eventChan: eventChan,
	}
}

func (d *defaultRetryer) RetryOnError(ctx context.Context, event store.Event, handler Retryable) {
	err := handler(ctx, event.DeepCopy())
	if err == nil {
		return
	}

	logrus.WithError(err).Warn("failed to handle event")

	go func() {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			logrus.WithField("type", event.Type).WithField("nn", event.Object.GetNamespaceName()).Debug("retrying event...")
			d.eventChan <- event
		}
	}()
}
