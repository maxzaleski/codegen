package gen

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/lib/datastructure"
	"github.com/maxzaleski/codegen/internal/metrics"
)

type (
	// IGenQueue is the interface that wraps the generation queue.
	IGenQueue = datastructure.IQueue[genJob]

	genQueue struct {
		datastructure.IQueue[genJob]

		l ILogger
	}
)

var _ IGenQueue = (*genQueue)(nil)

func newGenQueue(l slog.ILogger, m *metrics.Metrics, c Config) IGenQueue {
	nl := newLogger(l, "gen-queue", slog.None)
	return &genQueue{
		IQueue: datastructure.NewQueue[genJob](nl, m, c.WorkerCount),
		l:      nl,
	}
}

func (q *genQueue) Enqueue(j *genJob) {
	defer q.l.Ack("enqueue<-", j)
	q.IQueue.Enqueue(j)
}

func (q *genQueue) Dequeue(wID int) (j *genJob) {
	if j = q.IQueue.Dequeue(wID); j != nil {
		defer q.l.Ack(
			fmt.Sprintf("dequeue%s", slog.Atom(slog.Purple, fmt.Sprintf("->worker_%d", wID))),
			j,
		)
	}
	return
}
