package gen

import (
	"context"
	"fmt"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/pkg/errors"
	"os"
	"strings"
)

var errWorkerAllJobsProcessed = errors.New("all jobs processed")

func worker(ctx context.Context, id int, q IGenQueue) error {
	pkgs, m, logger :=
		ctx.Value(contextKeyPackages).([]*core.Package),
		ctx.Value(contextKeyMetrics).(*metrics.Metrics),
		newLogger(
			ctx.Value(contextKeyLogger).(slog.ILogger),
			fmt.Sprintf("worker_%d", id),
			slog.None,
		)

	logger.Log("start", "msg", "starting worker")
	defer logger.Log("exit", "msg", "worker exiting")

	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		j := q.Dequeue(id)
		if j == nil {
			return errWorkerAllJobsProcessed // queue is empty; all jobs processed.
		}

		mo, sk :=
			&metrics.FileOutcome{AbsolutePath: j.OutputFile.AbsolutePath},
			j.Metadata.ScopeKey
		if s := strings.Split(sk, "/"); len(s) == 2 {
			sk = s[0]
		}

		tag := core.UniquePkgAlias
		if p := j.Package; p != nil {
			tag = p.Name
		}
		m.CaptureScope(sk, tag, *mo)

		if err := generateFile(pkgs, logger, j); err != nil {
			if errors.Is(errFileAlreadyPresent, err) {
				return nil
			}
			return err
		}

		logFileOutcome(logger, fileOutcomeSuccess, j)
		mo.Created = true

		return nil
	}
}

var errFileAlreadyPresent = errors.New("file already exists")

func generateFile(pkgs []*core.Package, l ILogger, j *genJob) error {
	// TODO: evaluate whether to run job.
	if _, err := os.Stat(j.OutputFile.AbsolutePath); err != nil {
		if !os.IsNotExist(err) {
			return errors.WithMessagef(err, "failed presence check at '%s'", j.OutputFile.AbsolutePath)
		}
	} else {
		defer logFileOutcome(l, fileOutcomeIgnored, j)
		return errFileAlreadyPresent
	}

	tf := templateFactory{
		j:       j,
		pkgs:    pkgs,
		funcMap: nil,
	}
	return tf.ExecuteTemplate()
}
