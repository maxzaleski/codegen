package gen

import (
	"fmt"
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/core/slog"
	"math"
	"sync"
)

type (
	IQueue interface {
		// Enqueue enqueues a job.
		Enqueue(j *genJob)
		// Dequeue blocks until a job is available. Returns nil if the queue is closed.
		Dequeue(id int) (*genJob, bool)

		// GetSize returns the number of jobs in the queue.
		GetSize() int
		// GetCapacity returns the capacity of the queue .
		GetCapacity() int

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
		isClosed   bool
		capacity   int
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
func newQueue(l slog.ILogger, c Config) IQueue {
	capacity := int(math.Ceil(float64(c.WorkerCount) * 1.5))
	if c.DebugMode && c.WorkerCount < 10 { // Omit; prevents deadlock in debug mode.
		capacity = 10
	}

	return &queue{
		l: newLogger(l, "queue"),

		mu:        &sync.Mutex{},
		closeOnce: sync.Once{},
		readyOnce: sync.Once{},

		collection: make(chan *genJob, capacity),
		readyChan:  make(chan any, 0),
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

func (q *queue) Dequeue(wid int) (*genJob, bool) {
	j := <-q.collection // channels are thread-safe by default.
	if j == nil {
		if q.isClosed {
			return nil, false
		}
		panic("queue: race condition")
	}
	defer q.l.Ack(fmt.Sprintf("dequeue->worker_%d", wid), j)

	return j, true
}

func (q *queue) GetSize() int {
	return len(q.collection)
}

func (q *queue) GetCapacity() int {
	return q.capacity
}

func (q *queue) Ready() {
	q.readyOnce.Do(func() {
		defer q.logState("ready", "queue is ready")

		close(q.readyChan)
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
