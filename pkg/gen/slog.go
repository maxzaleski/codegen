package gen

import (
	"fmt"
	"github.com/codegen/internal/core/slog"
	"strings"
)

const (
	fileOutcomeSuccess fileOutcome = "created"
	fileOutcomeIgnored fileOutcome = "ignored"
)

type (
	fileOutcome string

	ILogger interface {
		Log(event string, fields ...any)
		Ack(event string, j *genJob, fields ...any)
	}

	logger struct {
		instance  slog.ILogger
		namespace string
		scopeKey  string
	}
)

// New creates a new logger.
func newLogger(parent slog.ILogger, ns string) ILogger {
	return &logger{
		instance:  parent,
		namespace: ns,
	}
}

func (l *logger) Log(event string, fields ...any) {
	if event == "" {
		panic("logger: event cannot be empty")
	}

	fieldsS := ""
	if fields != nil {
		for i := 0; i < len(fields); i += 2 {
			f, val := fields[i], fields[i+1]
			if _, ok := f.(string); !ok {
				panic("logger: field key must be a string")
			}
			fs := f.(string)
			if s, ok := val.(string); ok && !strings.Contains(fs, "file") {
				val = fmt.Sprintf("'%s'", s)
			}
			fieldsS += fs + "=" + slog.Atom(slog.Cyan, fmt.Sprintf("%v ", val))
		}
	}

	l.instance.Log(
		slog.Domain(slog.Purple, "origin", l.namespace),
		slog.Domain(slog.Blue, "event", event),
		fieldsS,
	)
}

func (l *logger) Ack(event string, j *genJob, fields ...any) {
	fieldsCopy := make([]any, 4, 4+len(fields))
	fieldsCopy[0], fieldsCopy[1] = "scope", j.Metadata.ScopeKey
	fieldsCopy[2], fieldsCopy[3] = "job", j.Key
	fieldsCopy = append(fieldsCopy, fields...)
	l.Log(event, fieldsCopy...)
}
