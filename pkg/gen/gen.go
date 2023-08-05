package gen

import (
	"context"
	"database/sql"
	"encoding/json"
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
	Config struct {
		// Enable debug mode; print out debug messages to stdout.
		DebugMode bool `json:"debug_mode"`
		// Enable verbose debug mode; print out verbose debug messages to stdout.
		DebugVerbose bool `json:"debug_verbose"`
		// Enable debug worker metrics.go; print out worker metrics.go to stdout.
		DebugWorkerMetrics bool `json:"debug_worker_metrics"`
		// Delete '{cwd}/.codegen/tmp' directory.
		DeleteTmp bool `json:"delete_tmp"`
		// Ignore specified templates; an empty template will be used instead.
		IgnoreTemplates bool `json:"ignore_templates"`
		// Disable log file; log file will not be created/populated.
		DisableLogFile bool `json:"disable_log_file"`
		// Location of the tool's folder; default: '{cwd}/.codegen'.
		Location string `json:"location"`
		// Number of workers available in the runtime concierge.
		WorkerCount int `json:"worker_count"`
		// TemplateFuncMap is a map of functions that can be called from templates.
		TemplateFuncMap template.FuncMap
	}

	Result struct {
		Metadata *core.Metadata
		Metrics  modules.IMetrics
	}
)

func (c Config) Marshal() ([]byte, error) {
	return json.Marshal(c)
}

func Execute(c Config, began time.Time) (res *Result, err error) {
	logger := slog.New(c.DebugMode, began)

	// [1] Parse configuration via `.codegen` directory.
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

	// [2] Start local sqlite database.
	dbc, err2 := db.New(logger, spec.Metadata.CodegenDir)
	if err = err2; err != nil {
		return
	}
	defer func(conn *sql.DB) { _ = conn.Close() }(dbc.Conn())

	// [3] Aggregate scopes from both domains.
	hScopes, pScopes := spec.Config.HttpDomain.Scopes, spec.Config.PkgDomain.Scopes
	ds := make([]*core.DomainScope, 0, len(hScopes)+len(pScopes))
	ds = append(ds, hScopes...)
	ds = append(ds, pScopes...)

	// [4] Start the runtime concierge.
	rc := newConcierge(errg, gctx, c, logger, dbc, ds)

	// -> Begin generation.
	rc.Start(c, spec, ds)

	// -> [blocking] wait for all goroutines to terminate.
	if err = rc.Wait(); err == nil {
		// Act upon the (dev) flag; print worker metrics.
		if c.DebugMode && c.DebugWorkerMetrics {
			defer printWorkerMetrics(logger, res.Metrics.GetWorkMetrics(), c.WorkerCount)
		}
	}

	return
}
