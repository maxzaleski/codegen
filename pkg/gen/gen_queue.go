package gen

import (
	"github.com/maxzaleski/codegen/internal/lib/datastructure"
	"github.com/maxzaleski/codegen/internal/slog"
)

type (
	// IQueue is the interface that wraps the generation queue.
	IQueue = datastructure.IQueue[genJob]

	genQueue struct {
		IQueue

		logger ILogger
	}
)

var _ IQueue = (*genQueue)(nil)

func newQueue(l slog.ILogger, c Config) IQueue {
	logger := newLogger(l, "queue", slog.None)
	return &genQueue{
		IQueue: datastructure.NewQueue[genJob](logger, c.WorkerCount),
		logger: logger,
	}
}

func (q *genQueue) Enqueue(j *genJob) {
	defer q.logger.Ack("enqueue<-", j)
	q.IQueue.Enqueue(j)
}

func (q *genQueue) Dequeue() (j *genJob) {
	if j = q.IQueue.Dequeue(); j != nil {
		defer q.logger.Ack("dequeue->", j)
	}
	return
}
