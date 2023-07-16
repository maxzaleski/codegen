package datastructure

import (
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"math"
	"sync"
)

type (
	IQueue[J any] interface {
		// Enqueue enqueues a job.
		Enqueue(j *J)
		// Dequeue blocks until a job is available. Returns nil if the queue is closed.
		Dequeue(wID int) *J

		// GetSize returns the number of jobs in the queue.
		GetSize() int
		// GetCapacity returns the capacity of the queue .
		GetCapacity() int
		// GetReady returns true if the queue is ready.
		GetReady() bool

		// Close closes the queue. This is a one-time operation and is safe to call multiple times.
		Close()
		// ReadyListener returns a channel that is closed when the queue is ready.
		ReadyListener() <-chan any
		// Ready marks the queue as ready. This is a one-time operation and is safe to call multiple times.
		Ready()
	}

	queue[T any] struct {
		mu       *sync.Mutex
		capacity int

		collection chan *T
		readyChan  chan any

		nl      slog.INamedLogger
		metrics *metrics.Metrics

		// isClosed means the queue is no longer accepting jobs.
		isClosed  bool
		closeOnce sync.Once
		// isReady means the queue is ready to be read from.
		isReady   bool
		readyOnce sync.Once
	}
)

var _ IQueue[any] = (*queue[any])(nil)

// NewQueue creates a new queue.
//
// The queue's capacity is set to `⌈workerCount * 1.5⌉`.
func NewQueue[J any](nl slog.INamedLogger, m *metrics.Metrics, workerCount int) IQueue[J] {
	capacity := int(math.Ceil(float64(workerCount) * 1.5))
	if workerCount < 10 { // Omit; prevents deadlock in debug mode.
		capacity = 10
	}

	return &queue[J]{
		mu:       &sync.Mutex{},
		capacity: capacity,

		nl:      nl,
		metrics: m,

		closeOnce: sync.Once{},
		readyOnce: sync.Once{},

		collection: make(chan *J, capacity),
		readyChan:  make(chan any, 0),
	}
}

func (q *queue[T]) Enqueue(j *T) {
	if q.isClosed {
		panic("queue is closed")
	}

	q.collection <- j // channels are thread-safe by default.
}

func (q *queue[T]) Dequeue(wID int) *T {
	j := <-q.collection // channels are thread-safe by default.
	if j == nil {
		if q.isClosed {
			return nil
		}
		panic("queue: race condition")
	}
	defer q.metrics.CaptureWorker(wID)

	return j
}

func (q *queue[T]) GetSize() int {
	return len(q.collection)
}

func (q *queue[T]) GetCapacity() int {
	return q.capacity
}

func (q *queue[T]) GetReady() bool {
	return q.isReady
}

func (q *queue[T]) Ready() {
	q.readyOnce.Do(func() {
		defer q.logState("ready", "queue is ready")

		close(q.readyChan)
		q.isReady = true
	})
}

func (q *queue[T]) ReadyListener() <-chan any {
	return q.readyChan
}

func (q *queue[T]) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closeOnce.Do(func() {
		defer q.logState("close", "queue is closed")

		close(q.collection)
		q.isClosed = true
	})
}

func (q *queue[any]) logState(state string, msg string) {
	q.nl.Log(state, "msg", msg, "remaining", q.GetSize())
}
