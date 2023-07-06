package gen

import (
	"context"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/metrics"
	"github.com/codegen/internal/utils"
	"github.com/pkg/errors"
	"time"

	"github.com/codegen/internal/core"
)

type (
	// Config represents the tool's configuration .
	Config struct {
		DebugMode          bool
		DebugWorkerMetrics bool
		DeleteTmp          bool
		DisableTemplates   bool
		Location           string
		WorkerCount        int

		TemplateFuncMap map[string]interface{}
	}

	domainScope struct {
		*core.DomainScope
		DomainType core.DomainType
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
	aggrScopes := make([]*domainScope, 0, len(hScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, mapToDomainScope(pScopes, core.DomainTypePkg)...)
	aggrScopes = append(aggrScopes, mapToDomainScope(hScopes, core.DomainTypeHttp)...)

	ms := metrics.New()

	ctx := context.Background()
	ctx = context.WithValue(ctx, contextKeyBegan, began)
	ctx = context.WithValue(ctx, contextKeyLogger, sl)
	ctx = context.WithValue(ctx, contextKeyMetrics, ms)
	ctx = context.WithValue(ctx, contextKeyPackages, spec.Pkgs)

	concierge := newConcierge(ctx, sl, c, aggrScopes)

	// -> Guarantees output directories exist during execution.
	concierge.WalkDirectoryStructure(spec.Metadata, spec.Pkgs)

	// -> Extract jobs and feed the queue.
	concierge.WgAdd(1)
	go concierge.ExtractJobs(spec)

	// [blocking] wait for queue to be ready.
	concierge.WaitQueueReadiness()

	// -> Start workers.
	concierge.WgAdd(1)
	go concierge.StartWorkers()

	// [blocking] wait for all goroutines to finish.
	concierge.WaitAndCleanup()

	// Act upon the (dev) flag; print worker metrics.
	if c.DebugMode && c.DebugWorkerMetrics {
		printWorkerMetrics(ms, sl, c.WorkerCount)
	}

	return md, ms, <-concierge.GetErrChannel()
}
