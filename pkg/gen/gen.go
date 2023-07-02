package gen

import (
	"context"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/metrics"
	"github.com/pkg/errors"
	"time"

	"github.com/codegen/internal/core"
)

var aggrScopes []*domainScope

type (
	// Config represents the setup configuration .
	Config struct {
		Location         string
		WorkerCount      int
		DisableTemplates bool
		DebugMode        bool
	}

	domainScope struct {
		*core.DomainScope
		DomainType core.DomainType
	}
)

func Execute(c Config, began time.Time) (md *core.Metadata, _ metrics.IMetrics, _ error) {
	sl := slog.New(c.DebugMode, began)

	// Parse configuration via `.codegen` directory.
	spec, err := core.NewSpec(sl, c.Location)
	md = spec.Metadata
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	sc := spec.Config
	hScopes, pScopes := sc.HttpDomain.Scopes, sc.PkgDomain.Scopes
	aggrScopes = make([]*domainScope, 0, len(hScopes)+len(pScopes))
	aggrScopes = append(aggrScopes, mapToDomainScope(pScopes, core.DomainTypePkg)...)
	aggrScopes = append(aggrScopes, mapToDomainScope(hScopes, core.DomainTypeHttp)...)

	ms := metrics.New(nil)

	ctx := context.Background()
	ctx = context.WithValue(ctx, "began", began)
	ctx = context.WithValue(ctx, "logger", sl)
	ctx = context.WithValue(ctx, "metrics", ms)
	ctx = context.WithValue(ctx, "packages", spec.Pkgs)

	gs := newSetup(ctx, sl, c)

	// -> Guarantees output directories exist during execution.
	gs.WalkDirectoryStructure(spec.Metadata, spec.Pkgs)

	// -> Extract jobs and feed the queue.
	gs.WgAdd(1)
	go gs.ExtractJobs(spec)

	// [blocking] wait for queue to be ready.
	gs.WaitQueueReadiness()

	// -> Start workers.
	gs.WgAdd(1)
	go gs.StartWorkers()

	// [blocking] wait for all goroutines to finish.
	gs.WaitAndCleanup()

	return md, ms, <-gs.GetErrChannel()
}
