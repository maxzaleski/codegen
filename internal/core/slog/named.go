package slog

import (
	"fmt"
	"strings"
)

type (
	// INamedLogger represents a named logger.
	//
	// It differs from the default `ILogger` by providing a namespace to subsequent logs.
	INamedLogger interface {
		// Log logs an event.
		Log(event string, lines ...any)
	}

	namedLogger struct {
		*logger

		namespace string
	}
)

func NewNamed(p ILogger, ns string) INamedLogger {
	return &namedLogger{
		logger:    p.(*logger),
		namespace: ns,
	}
}

func (l *namedLogger) Log(event string, fields ...interface{}) {
	if event == "" {
		l.panic("event cannot be empty")
	}

	// Transform log fields into a printable string,
	// e.g. fields = [k1, v1, k2, v2,...kn, vn] => "k1=v1 k2=v2 ... kn=vn"
	//
	// Specificities: values are quoted if a string and output as cyan.
	fieldsS := ""
	if fields != nil {
		for i := 0; i < len(fields); i += 2 {
			f, val := fields[i], fields[i+1]
			if _, ok := f.(string); !ok {
				l.panic("field key must be a string")
			}

			fs := f.(string)
			if s, ok := val.(string); ok && !strings.Contains(fs, "file") {
				val = fmt.Sprintf("'%s'", s)
			}

			fieldsS += fs + "=" + Atom(Cyan, fmt.Sprintf("%v ", val))
		}
	}

	l.logger.Log(
		domain(Purple, "origin", l.namespace),
		domain(Blue, "event", event),
		fieldsS,
	)
}

func (l *namedLogger) panic(msg string) {
	panic(fmt.Sprintf("namedLogger.Log: %s", msg))
}
