package gen

import (
	"context"
	"github.com/maxzaleski/codegen/internal"
	"github.com/maxzaleski/codegen/internal/core/slog"
	"github.com/maxzaleski/codegen/internal/metrics"
	"github.com/maxzaleski/codegen/pkg/gen/diagnostics"
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
		// Enable verbose debug mode; print out verbose debug messages to stdout.
		DebugVerbose bool
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
		// Number of workers available in the runtime concierge.
		WorkerCount int
		// TemplateFuncMap is a map of functions that can be called from templates.
		TemplateFuncMap template.FuncMap
	}
)

const (
	contextKeyBegan    internal.ContextKey = "began"
	contextKeyLogger   internal.ContextKey = "logger"
	contextKeyMetrics  internal.ContextKey = "metrics"
	contextKeyPackages internal.ContextKey = "packages"
)

func Execute(c Config, began time.Time) (md *core.Metadata, _ metrics.IMetrics, err error) {
	sl := slog.New(c.DebugMode, began)

	var spec *core.Spec

	// Parse configuration via `.codegen` directory.
	spec, err = core.NewSpec(sl, c.Location)
	md = spec.Metadata // Always returned.
	if err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	// Run diagnostics.
	ctx1 := context.Background()
	if err = diagnostics.Run(ctx1, sl, md.CodegenDir); err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	// -> [dev] Act upon the flag; delete tmp directory.
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

	errg, ctx2 := errgroup.WithContext(ctx1)
	ctx2 = context.WithValue(ctx2, contextKeyBegan, began)
	ctx2 = context.WithValue(ctx2, contextKeyLogger, sl)
	ctx2 = context.WithValue(ctx2, contextKeyMetrics, ms)
	ctx2 = context.WithValue(ctx2, contextKeyPackages, spec.Pkgs)

	// Start the runtime concierge.
	concierge := newConcierge(ctx2, sl, errg, c)

	// -> Execute preflight functions.
	concierge.Start(spec, c)

	// [blocking] wait for all goroutines to terminate.
	if err = concierge.Wait(); err == nil {
		// Act upon the (dev) flag; print worker metrics.
		if c.DebugMode && c.DebugWorkerMetrics {
			defer printWorkerMetrics(ms, sl, c.WorkerCount)
		}
	}

	return md, ms, err
}
