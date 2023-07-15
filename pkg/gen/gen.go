package gen

import (
	"context"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/maxzaleski/codegen/internal/utils"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
	"text/template"
	"time"

	"github.com/maxzaleski/codegen/internal/core"
)

type (
	// Config represents the tool's configuration .
	Config struct {
		// Enable debug mode; print out debug messages to stdout.
		DebugMode bool
		// Enable debug worker metrics; print out worker metrics to stdout.
		DebugWorkerMetrics bool
		// Delete '{cwd}/.codegen/tmp' directory.
		DeleteTmp bool
		// Ignore specified templates; an empty template will be used instead.
		IgnoreTemplates bool
		// Disable log file; log file will not be created/populated.
		DisableLogFile bool
		// Location of the tool's folder; default: '{cwd}/.codegen'.
		Location string
		// Number of workers available in the runtime pool.
		WorkerCount int
		// TemplateFuncMap is a map of functions that can be called from templates.
		TemplateFuncMap template.FuncMap
	}
)

const (
	contextKeyBegan    utils.ContextKey = "began"
	contextKeyLogger   utils.ContextKey = "logger"
	contextKeyMetrics  utils.ContextKey = "metrics"
	contextKeyPackages utils.ContextKey = "packages"
)

func Execute(c Config, began time.Time) (md *core.Metadata, _ metrics.IMetrics, _ error) {
	sl := slog.New(c.DebugMode, began)

	// Parse configuration via `.codegen` directory.
	spec, err := core.NewSpec(sl, c.Location)
	md = spec.Metadata // Always returned.
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	// Act upon the (dev) flag; delete tmp directory.
	if c.DeleteTmp {
		if err = removeTmpDir(md, sl); err != nil {
			return
		}
	}

	sc := spec.Config

	hScopes, pScopes := sc.HttpDomain.Scopes, sc.PkgDomain.Scopes
	aggrScopes := make([]*core.DomainScope, 0, len(hScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, hScopes...)
	aggrScopes = append(aggrScopes, pScopes...)

	ms := metrics.New()

	errg, ctx := errgroup.WithContext(context.Background())
	ctx = context.WithValue(ctx, contextKeyBegan, began)
	ctx = context.WithValue(ctx, contextKeyLogger, sl)
	ctx = context.WithValue(ctx, contextKeyMetrics, ms)
	ctx = context.WithValue(ctx, contextKeyPackages, spec.Pkgs)

	concierge := newConcierge(ctx, sl, errg, c, aggrScopes)

	// -> Guarantees output directories exist during execution.
	if err = concierge.WalkDirectoryStructure(spec.Metadata, spec.Pkgs); err != nil {
		err = errors.Wrapf(err, "failed to walk directory structure")
		return
	}

	// -> Extract jobs and feed the queue.
	errg.Go(func() error { return concierge.ExtractJobs(spec) })

	// [blocking] wait for queue to be ready.
	concierge.WaitQueueReadiness()

	// -> Start workers.
	errg.Go(func() error { return concierge.StartWorkers() })

	// [blocking] wait for all goroutines to terminate.
	if err = concierge.WaitAndCleanup(); err == nil {
		// Act upon the (dev) flag; print worker metrics.
		if c.DebugMode && c.DebugWorkerMetrics {
			defer printWorkerMetrics(ms, sl, c.WorkerCount)
		}
	}

	return md, ms, err
}
