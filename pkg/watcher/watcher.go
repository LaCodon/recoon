package watcher

import (
	"context"
	metav1 "github.com/lacodon/recoon/pkg/api/v1/meta"
	"github.com/lacodon/recoon/pkg/store"
	"sync"
)

type Watcher interface {
	Watch(kinds ...metav1.VersionKind) <-chan store.Event
}

type DefaultWatcher struct {
	subs  []subscriber
	input <-chan store.Event
	mu    sync.Mutex
}

type subscriber struct {
	Chan   chan store.Event
	Filter []metav1.VersionKind
}

func NewDefaultWatcher(input <-chan store.Event) *DefaultWatcher {
	return &DefaultWatcher{
		subs:  make([]subscriber, 0),
		input: input,
	}
}

func (w *DefaultWatcher) Watch(kinds ...metav1.VersionKind) <-chan store.Event {
	w.mu.Lock()
	defer w.mu.Unlock()

	sub := subscriber{
		Chan:   make(chan store.Event, 50),
		Filter: kinds,
	}
	w.subs = append(w.subs, sub)

	return sub.Chan
}

func (w *DefaultWatcher) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			w.unsubscribeAll()
			return nil
		case evt := <-w.input:
			w.fanOut(evt)
		}
	}
}

func (w *DefaultWatcher) unsubscribeAll() {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, sub := range w.subs {
		close(sub.Chan)
	}

	w.subs = make([]subscriber, 0)
}

func (w *DefaultWatcher) fanOut(event store.Event) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for _, sub := range w.subs {
		if w.shouldCollect(event.Object.GetVersionKind(), sub.Filter) {
			sub.Chan <- event.DeepCopy()
		}
	}
}

func (w *DefaultWatcher) shouldCollect(evtKind metav1.VersionKind, kinds []metav1.VersionKind) bool {
	if kinds == nil {
		return true
	}

	for _, kind := range kinds {
		if evtKind == kind {
			return true
		}
	}

	return false
}
