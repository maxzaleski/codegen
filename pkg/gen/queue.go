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
		Enqueue(j *genJob)
		Dequeue(id int) (*genJob, bool)
		Size() int
		Close()
		// WaitReadiness blocks until the queue is ready.
		WaitReadiness()
		// Ready marks the queue as ready. This is a one-time operation and is safe to call multiple times.
		Ready()
	}

	queue struct {
		l ILogger

		mu        *sync.Mutex
		readyOnce sync.Once
		closeOnce sync.Once

		collection chan *genJob
		readyChan  chan bool
		isClosed   bool
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
		readyChan:  make(chan bool, 1),
	}
}

func (q *queue) Enqueue(j *genJob) {
	if q.isClosed {
		panic("queue is closed")
	}
	defer q.l.Ack("enqueue<-", j)

	q.collection <- j
}

func (q *queue) Dequeue(wid int) (*genJob, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()

	j := <-q.collection
	if j == nil {
		if q.isClosed {
			return nil, false
		}
		panic("queue: race condition")
	}
	defer q.l.Ack(fmt.Sprintf("dequeue->worker_%d", wid), j)

	return j, true
}

func (q *queue) Size() int {
	return len(q.collection)
}

func (q *queue) Ready() {
	q.readyOnce.Do(func() {
		defer q.l.Log("ready", "msg", "queue is ready", "size", q.Size())

		q.readyChan <- true
	})
}

func (q *queue) WaitReadiness() {
	<-q.readyChan
}

func (q *queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()

	q.closeOnce.Do(func() {
		defer q.l.Log("close", "msg", "queue is closed", "remaining", q.Size())

		close(q.collection)
		q.isClosed = true
	})
}
