package gen

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/pkg/errors"
	"os"
)

const (
	fileOutcomeSuccess jobOutcome = "created"
	fileOutcomeIgnored jobOutcome = "already-exists"
)

type jobOutcome string

func logFileOutcome(l ILogger, o jobOutcome, j *genJob) {
	l.Ack("file", j, "status", string(o), "file", j.OutputFile.AbsolutePath)
}

func removeTmpDir(md *core.Metadata, l slog.ILogger) error {
	path := md.Cwd + "/tmp"

	l.LogEnv("Delete tmp directory", "deleteTmp=1", "deleting "+path)

	if err := os.RemoveAll(path); err != nil {
		return errors.Wrapf(err, "failed to delete tmp directory at '%s'", path)
	}
	return nil
}

func printWorkerMetrics(logger slog.ILogger, m map[int]int, wc int) {
	keys, total, highest := 0, 0, 0
	for _, w := range m {
		total += w
		if w > highest {
			highest = w
		}
		keys++
	}

	msg := fmt.Sprintf("workers=%d avg_throughput=%d highest=%d utilised=%d%%",
		wc,
		total/keys,
		highest,
		(keys*100)/wc)
	logger.LogEnv("Worker metrics", "workerMetrics=1", msg)
}
