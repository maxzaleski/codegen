package gen

import (
	"github.com/maxzaleski/codegen/internal/lib/datastructure"
	"github.com/maxzaleski/codegen/internal/slog"
)

func newQueue(logger slog.ILogger, c Config) datastructure.IQueue[genJob] {
	qLogger := newLogger(logger, "queue", slog.None)
	qHooks := datastructure.QueueHooks[genJob]{
		OnEnqueue: func(j *genJob) {
			qLogger.Ack("enqueue<-", j)
		},
		OnDequeue: func(j *genJob) {
			qLogger.Ack("dequeue->", j)
		},
	}
	return datastructure.NewQueue[genJob](qLogger, c.WorkerCount, &qHooks)
}
