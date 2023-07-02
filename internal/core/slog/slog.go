package slog

import (
	"fmt"
	"sync"
	"time"
)

type (
	ILogger interface {
		// Log wraps `fmt.Println`.
		Log(lines ...interface{})
		// Logf wraps `fmt.Println` with the given format.
		Logf(format string, a ...interface{})
	}

	logger struct {
		mu *sync.Mutex

		debugFlag bool
		began     time.Time
	}
)

// New creates a new logger.
func New(debugFlag bool, began time.Time) ILogger {
	if began.IsZero() {
		panic("logger: began cannot be zero")
	}

	l := &logger{
		mu:        &sync.Mutex{},
		debugFlag: debugFlag,
		began:     began,
	}
	if debugFlag {
		l.Log(
			Atom(Red, "Debug flag is set (--debug=1); Debug mode enabled, printing subsequent logs"))
	}
	return l
}

func (l *logger) Log(lines ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugFlag {
		linesCopy := make([]interface{}, 1, len(lines)+1)
		linesCopy[0] = Domain(Yellow, "start", "+"+time.Since(l.began).String())
		linesCopy = append(linesCopy, lines...)
		fmt.Println(linesCopy...)
	}
}

func (l *logger) Logf(format string, a ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugFlag {
		s := fmt.Sprintf(format, a...)
		fmt.Println(s)
	}
}

// Domain returns a string in the format of `[domain:value]`.
func Domain(tc Colour, domain, value string) string {
	return fmt.Sprintf("[%s:%s]", domain, Atom(tc, value))
}

// OriginDomain wraps `Domain` with the domain set to `origin`.
func OriginDomain(value string) string {
	return Domain(Purple, "origin", value)
}
