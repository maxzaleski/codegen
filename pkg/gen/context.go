package gen

import (
	"context"
	"github.com/maxzaleski/codegen/internal"
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/slog"
	"github.com/maxzaleski/codegen/pkg/gen/modules"
	"time"
)

type (
	IContext interface {
		GetUnderlying() context.Context
		GetBegan() time.Time
		GetLogger() slog.ILogger
		GetMetrics() modules.IMetrics
		GetPackages() []*core.Package
		SetUnderlying(ctx context.Context)
		SetAny(key internal.ContextKey, val any)
	}

	genContext struct {
		ctx context.Context
	}
)

const (
	contextKeyBegan    internal.ContextKey = "began"
	contextKeyLogger   internal.ContextKey = "logger"
	contextKeyMetrics  internal.ContextKey = "metrics.go"
	contextKeyPackages internal.ContextKey = "packages"
)

var _ IContext = (*genContext)(nil)

func newGenContext(ctx context.Context) IContext {
	return &genContext{ctx: ctx}
}

func (c *genContext) GetUnderlying() context.Context {
	return c.ctx
}

func (c *genContext) GetBegan() time.Time {
	return c.ctx.Value(contextKeyBegan).(time.Time)
}

func (c *genContext) GetLogger() slog.ILogger {
	return c.ctx.Value(contextKeyLogger).(slog.ILogger)
}

func (c *genContext) GetMetrics() modules.IMetrics {
	return c.ctx.Value(contextKeyMetrics).(modules.IMetrics)
}

func (c *genContext) GetPackages() []*core.Package {
	return c.ctx.Value(contextKeyPackages).([]*core.Package)
}

func (c *genContext) SetUnderlying(ctx context.Context) {
	c.ctx = ctx
}

func (c *genContext) SetAny(key internal.ContextKey, val any) {
	c.ctx = context.WithValue(c.ctx, key, val)
}
