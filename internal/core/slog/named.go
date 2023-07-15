package slog

import (
	"fmt"
	"github.com/iancoleman/strcase"
	"strings"
)

type (
	// INamedLogger represents a named logger.
	//
	// It differs from the default `ILogger` by providing a namespace to subsequent logs.
	INamedLogger interface {
		// Log logs an event.
		Log(event string, fields ...any)
		// Parent returns the parent logger.
		Parent() ILogger
	}

	namedLogger struct {
		*logger

		namespace string
		colour    Colour
	}
)

func NewNamed(p ILogger, ns string, c Colour) INamedLogger {
	return &namedLogger{
		logger:    p.(*logger),
		namespace: ns,
		colour:    c,
	}
}

func (l *namedLogger) Log(event string, fields ...interface{}) {
	if event == "" {
		l.panic("event cannot be empty")
	}
	if len(fields)%2 != 0 {
		l.panic("fields must be key-value pairs")
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

			fs := strcase.ToSnake(f.(string))
			if s, ok := val.(string); ok && !strings.Contains(fs, "file") {
				val = fmt.Sprintf("'%s'", s)
			}

			fieldsS += fs + "=" + Atom(Cyan, fmt.Sprintf("%v ", val))
		}
	}

	oc := Purple
	if l.colour != "" {
		oc = l.colour
	}
	ec := Blue
	if strings.Contains(event, "error") {
		ec = Red
	}
	l.logger.Log(
		domain(oc, "origin", l.namespace),
		domain(ec, "event", event),
		fieldsS,
	)
}

func (l *namedLogger) Parent() ILogger {
	return l.logger
}

func (l *namedLogger) panic(msg string) {
	panic(fmt.Sprintf("namedLogger.Log: %s", msg))
}
