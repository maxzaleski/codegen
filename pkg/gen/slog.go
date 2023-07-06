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

	logger struct {
		slog.INamedLogger
	}
)

// New creates a new logger specific to the `gen` package.
func newLogger(p slog.ILogger, ns string, co slog.Colour) ILogger {
	return &logger{slog.NewNamed(p, ns, co)}
}

func (l *logger) Ack(event string, j *genJob, fields ...any) {
	fieldsCopy := make([]any, 4, 4+len(fields))
	fieldsCopy[0], fieldsCopy[1] = "scope", j.Metadata.ScopeKey
	fieldsCopy[2], fieldsCopy[3] = "job", j.Key
	fieldsCopy = append(fieldsCopy, fields...)
	l.Log(event, fieldsCopy...)
}

func (l *logger) NamedParent() slog.INamedLogger {
	return l.INamedLogger
}

func (l *logger) Parent() slog.ILogger {
	return l.INamedLogger.Parent()
}
