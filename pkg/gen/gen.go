package gen

import (
	"context"
	"database/sql"
	"github.com/maxzaleski/codegen/internal"
	"github.com/maxzaleski/codegen/internal/db"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/maxzaleski/codegen/pkg/gen/modules"
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
		// Enable debug worker metrics.go; print out worker metrics.go to stdout.
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

	Result struct {
		Metadata *core.Metadata
		Metrics  modules.IMetrics
	}
)

const (
	contextKeyBegan    internal.ContextKey = "began"
	contextKeyLogger   internal.ContextKey = "logger"
	contextKeyMetrics  internal.ContextKey = "metrics.go"
	contextKeyPackages internal.ContextKey = "packages"
)

func Execute(c Config, began time.Time) (res *Result, err error) {
	logger := slog.New(c.DebugMode, began)

	// Parse configuration via `.codegen` directory.
	spec, err1 := core.NewSpec(logger, c.Location)
	res = &Result{
		Metadata: spec.Metadata, // Always returned; error handled second.
		Metrics:  modules.NewMetrics(),
	}
	if err = err1; err != nil {
		err = errors.Wrapf(err, "failed to produce a new specification")
		return
	}

	// -> [dev] Act upon the flag; delete tmp directory.
	if c.DeleteTmp {
		if err = removeTmpDir(res.Metadata, logger); err != nil {
			return
		}
	}

	errg, ctx := errgroup.WithContext(context.Background())
	gctx := newGenContext(ctx)
	gctx.SetAny(contextKeyBegan, began)
	gctx.SetAny(contextKeyLogger, logger)
	gctx.SetAny(contextKeyMetrics, res.Metrics)
	gctx.SetAny(contextKeyPackages, spec.Pkgs)

	// -> Start local sqlite database.
	dbc, err2 := db.New(logger, spec.Metadata.CodegenDir)
	if err = err2; err != nil {
		return
	}
	defer func(conn *sql.DB) { _ = conn.Close() }(dbc.Conn())

	hScopes, pScopes := spec.Config.HttpDomain.Scopes, spec.Config.PkgDomain.Scopes
	ds := make([]*core.DomainScope, 0, len(hScopes)+len(pScopes))
	ds = append(ds, hScopes...)
	ds = append(ds, pScopes...)

	// Start the runtime concierge.
	concierge := newConcierge(gctx, errg, c, logger, dbc, ds)

	// -> Execute preflight functions.
	concierge.Start(spec, c)

	// [blocking] wait for all goroutines to terminate.
	if err = concierge.Wait(); err == nil {
		// Act upon the (dev) flag; print worker metrics.
		if c.DebugMode && c.DebugWorkerMetrics {
			defer printWorkerMetrics(logger, res.Metrics.GetWorkMetrics(), c.WorkerCount)
		}
	}

	return
}
