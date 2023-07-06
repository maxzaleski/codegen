package gen

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"math"
	"sync"
)

type (
	IQueue interface {
		// Enqueue enqueues a job.
		Enqueue(j *genJob)
		// Dequeue blocks until a job is available. Returns nil if the queue is closed.
		Dequeue(id int) *genJob

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

	queue struct {
		l ILogger

		mu        *sync.Mutex
		readyOnce sync.Once
		closeOnce sync.Once

		collection chan *genJob
		readyChan  chan any

		// isClosed means the queue is no longer accepting jobs.
		isClosed bool
		// isReady means the queue is ready to be read from.
		isReady bool

		capacity int
		metrics  *metrics.Metrics
	}

	genJob struct {
		*core.ScopeJob

		Metadata         Metadata
		Package          *core.Package
		DisableTemplates bool
	}

	Metadata struct {
		core.Metadata

		ScopeKey     string
		DomainType   core.DomainType
		AbsolutePath string
		Inline       bool
	}
)

var _ IQueue = (*queue)(nil)

// newQueue creates a new queue.
//
// The queue's capacity is set to `ceil(workerCount * 1.5)`.
func newQueue(l slog.ILogger, m *metrics.Metrics, c Config) IQueue {
	capacity := int(math.Ceil(float64(c.WorkerCount) * 1.5))
	if c.DebugMode && c.WorkerCount < 10 { // Omit; prevents deadlock in debug mode.
		capacity = 10
	}

	return &queue{
		l: newLogger(l, "queue", slog.None),

		mu:        &sync.Mutex{},
		closeOnce: sync.Once{},
		readyOnce: sync.Once{},

		collection: make(chan *genJob, capacity),
		readyChan:  make(chan any, 0),
		metrics:    m,
		capacity:   capacity,
	}
}

func (q *queue) Enqueue(j *genJob) {
	if q.isClosed {
		panic("queue is closed")
	}
	defer q.l.Ack("enqueue<-", j)

	q.collection <- j // channels are thread-safe by default.
}

func (q *queue) Dequeue(wid int) *genJob {
	j := <-q.collection // channels are thread-safe by default.
	if j == nil {
		if q.isClosed {
			return nil
		}
		panic("queue: race condition")
	}
	defer func() {
		q.l.Ack(
			fmt.Sprintf("dequeue%s",
				slog.Atom(slog.Purple, fmt.Sprintf("->worker_%d", wid))), j)
		q.metrics.CaptureWorker(wid)
	}()

	return j
}

func (q *queue) GetSize() int {
	return len(q.collection)
}

func (q *queue) GetCapacity() int {
	return q.capacity
}

func (q *queue) GetReady() bool {
	return q.isReady
}

func (q *queue) Ready() {
	q.readyOnce.Do(func() {
		defer q.logState("ready", "queue is ready")

		close(q.readyChan)
		q.isReady = true
	})
}

func (q *queue) ReadyListener() <-chan any {
	return q.readyChan
}

func (q *queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closeOnce.Do(func() {
		defer q.logState("close", "queue is closed")

		close(q.collection)
		q.isClosed = true
	})
}

func (q *queue) logState(state string, msg string) {
	q.l.Log(state, "msg", msg, "remaining", q.GetSize())
}
