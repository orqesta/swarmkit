// Package watch exposes ordered in-process observations of SwarmKit state.
// All notifications use the local transport from jmeiracorbal/go-events.
package watch

import (
	"context"
	"fmt"
	"sync"

	"github.com/jmeiracorbal/go-events/adapters/local"
)

const topic = "swarmkit.state"

// Queue is a collection of filtered subscribers of one local event bus.
// It does not persist state, schedule tasks or introduce another bus.
type Queue struct {
	mu      sync.Mutex
	bus     *local.Bus
	watches map[*local.ValueWatcher]struct{}
	limit   int
	closed  bool
}

// NewQueue initializes a local subscription queue.
func NewQueue(options ...func(*Queue) error) *Queue {
	q := &Queue{bus: local.New(), watches: make(map[*local.ValueWatcher]struct{})}
	for _, option := range options {
		if err := option(q); err != nil {
			panic(fmt.Sprintf("invalid watch queue option: %s", err))
		}
	}
	return q
}

// WithLimit bounds not-yet-delivered events per watcher. Zero is unlimited.
// A full subscriber is terminated and Publish returns an explicit error.
func WithLimit(limit uint64) func(*Queue) error {
	return func(q *Queue) error {
		if limit > uint64(^uint(0)>>1) {
			return fmt.Errorf("watch limit %d exceeds platform capacity", limit)
		}
		q.limit = int(limit)
		return nil
	}
}

// Watch subscribes to subsequent events until cancel is called.
func (q *Queue) Watch() (<-chan any, func()) {
	return q.CallbackWatch(nil)
}

// WatchContext closes the observation when its context is cancelled.
func (q *Queue) WatchContext(ctx context.Context) <-chan any {
	stream, _ := q.register(ctx, nil)
	return stream
}

// CallbackWatch subscribes to events accepted by matcher.
func (q *Queue) CallbackWatch(matcher func(any) bool) (<-chan any, func()) {
	return q.register(context.Background(), matcher)
}

// CallbackWatchContext subscribes with cancellation and event filtering.
func (q *Queue) CallbackWatchContext(ctx context.Context, matcher func(any) bool) <-chan any {
	stream, _ := q.register(ctx, matcher)
	return stream
}

func (q *Queue) register(ctx context.Context, matcher func(any) bool) (<-chan any, func()) {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		ch := make(chan any)
		close(ch)
		return ch, func() {}
	}
	w, err := q.bus.WatchValues(ctx, topic, matcher, nil, q.limit)
	if err != nil {
		q.mu.Unlock()
		panic(fmt.Sprintf("cannot register SwarmKit watcher: %s", err))
	}
	q.watches[w] = struct{}{}
	q.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			_ = w.Close()
			q.mu.Lock()
			delete(q.watches, w)
			q.mu.Unlock()
		})
	}
	go func() {
		<-w.Done()
		cancel()
	}()
	return w.Events(), cancel
}

// Publish dispatches a native Go value to the current observers. A saturated
// subscriber closes and returns an error; events are never silently dropped.
func (q *Queue) Publish(item any) error {
	return q.bus.PublishValue(context.Background(), topic, item)
}

// Close ends all observations and prevents further publishing.
func (q *Queue) Close() error {
	q.mu.Lock()
	if q.closed {
		q.mu.Unlock()
		return nil
	}
	q.closed = true
	watches := make([]*local.ValueWatcher, 0, len(q.watches))
	for watcher := range q.watches {
		watches = append(watches, watcher)
	}
	clear(q.watches)
	q.mu.Unlock()
	for _, watcher := range watches {
		_ = watcher.Close()
	}
	return q.bus.Close()
}
