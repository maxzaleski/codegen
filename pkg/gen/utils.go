package gen

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/pkg/errors"
	"os"
	"strings"
)

const (
	fileOutcomeSuccess fileOutcome = "created"
	fileOutcomeIgnored fileOutcome = "already-exists"
)

type fileOutcome string

func logFileOutcome(l ILogger, o fileOutcome, j *genJob) {
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

func printWorkerMetrics(m metrics.IMetrics, l slog.ILogger, wc int) {
	// Filter out the scope keys.
	keys := slice.Filter(m.Keys(), func(s string) bool {
		return strings.HasPrefix(s, "worker_")
	})

	// Calculate the total throughput.
	total := 0
	for _, k := range keys {
		total += m.Get(k).(int)
	}

	msg := fmt.Sprintf("workers=%d avg_throughput=%d utilised=%d%%", wc, total/len(keys), (len(keys)*100)/wc)
	l.LogEnv("Worker metrics", "workerMetrics=1", msg)
}
