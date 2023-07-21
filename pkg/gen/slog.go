package gen

import (
	"github.com/maxzaleski/codegen/internal/core/slog"
)

type (
	// ILogger is the logger interface for the `gen` package.
	ILogger interface {
		slog.INamedLogger

		// Ack logs an acknowledgement event.
		Ack(event string, j *genJob, fields ...any)
		// NamedParent returns the parent (named) logger.
		NamedParent() slog.INamedLogger
		// Parent returns the parent logger.
		Parent() slog.ILogger
	}

	logger_ struct { // '_' used to avoid name collision with this type.
		slog.INamedLogger
	}
)

// New creates a new logger specific to the `gen` package.
func newLogger(p slog.ILogger, ns string, co slog.Colour) ILogger {
	return &logger_{slog.NewNamed(p, ns, co)}
}

func (l *logger_) Ack(event string, j *genJob, fields ...any) {
	fieldsCopy := make([]any, 4, 4+len(fields))
	fieldsCopy[0], fieldsCopy[1] = "scope", j.Metadata.ScopeKey
	fieldsCopy[2], fieldsCopy[3] = "job", j.Key
	fieldsCopy = append(fieldsCopy, fields...)
	l.Log(event, fieldsCopy...)
}

func (l *logger_) NamedParent() slog.INamedLogger {
	return l.INamedLogger
}

func (l *logger_) Parent() slog.ILogger {
	return l.INamedLogger.Parent()
}
