package watch

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jmeiracorbal/go-events/adapters/local"
)

type observation struct {
	Name string
}

func receive(t *testing.T, ch <-chan any) any {
	t.Helper()
	select {
	case value, ok := <-ch:
		if !ok {
			t.Fatal("observation channel closed unexpectedly")
		}
		return value
	case <-time.After(3 * time.Second):
		t.Fatal("timed out waiting for event")
		return nil
	}
}

func TestWatchFanoutFilteringAndCancel(t *testing.T) {
	q := NewQueue()
	defer q.Close()
	all, stopAll := q.Watch()
	defer stopAll()
	matched, stopMatched := q.CallbackWatch(func(value any) bool {
		return value.(observation).Name == "first"
	})
	defer stopMatched()

	if err := q.Publish(observation{Name: "first"}); err != nil {
		t.Fatal(err)
	}
	if err := q.Publish(observation{Name: "second"}); err != nil {
		t.Fatal(err)
	}
	if got := []any{receive(t, all), receive(t, all)}; !reflect.DeepEqual(got, []any{
		observation{Name: "first"}, observation{Name: "second"},
	}) {
		t.Fatalf("observation order mismatch: %v", got)
	}
	if got := receive(t, matched); got != (observation{Name: "first"}) {
		t.Fatalf("matcher received wrong notification: %v", got)
	}
	stopMatched()
	select {
	case _, ok := <-matched:
		if ok {
			t.Fatal("cancelled watcher delivered a value")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("cancelled watcher remained open")
	}
}

func TestWatchContextCancellation(t *testing.T) {
	q := NewQueue()
	defer q.Close()
	ctx, cancel := context.WithCancel(context.Background())
	ch := q.WatchContext(ctx)
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("cancelled watcher must stop")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("context cancellation not observed")
	}
}

func TestWatchRejectsOverflow(t *testing.T) {
	q := NewQueue(WithLimit(1))
	defer q.Close()
	ch, cancel := q.Watch()
	defer cancel()
	var overflow error
	for i := range 30 {
		if err := q.Publish(i); err != nil {
			overflow = err
			break
		}
	}
	if !errors.Is(overflow, local.ErrWatchFull) {
		t.Fatalf("watch queue overflow must reach publisher: %v", overflow)
	}
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("overflowed watcher must be closed")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("overflowed watcher remained open")
	}
}

func TestWatchCloseIsIdempotent(t *testing.T) {
	q := NewQueue()
	ch, stop := q.Watch()
	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
	stop()
	if err := q.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("closed queue delivered a value")
		}
	case <-time.After(3 * time.Second):
		t.Fatal("queue was not closed")
	}
	if err := q.Publish("after-close"); err == nil {
		t.Fatal("publishing after close must fail")
	}
}
