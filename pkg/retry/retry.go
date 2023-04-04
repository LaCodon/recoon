package retry

import (
	"context"
	"github.com/lacodon/recoon/pkg/store"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"time"
)

type Retryable func(ctx context.Context, event store.Event) error

func KeepRetrying(ctx context.Context, event store.Event, handler Retryable) error {
	for i := 0; i < 5; i++ {
		err := handler(ctx, event.DeepCopy())
		if err == nil {
			return nil
		}

		logrus.WithError(err).Warn()

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(5 * time.Second):
			continue
		}
	}

	return errors.New("failed after 4 retries")
}
