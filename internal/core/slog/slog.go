package slog

import (
	"fmt"
	"sync"
)

type (
	ILogger interface {
		// Log wraps `fmt.Println`.
		Log(lines ...interface{})
		// Logf wraps `fmt.Println` with the given format.
		Logf(format string, a ...any)
	}

	logger struct {
		mu        *sync.Mutex
		debugFlag bool
		workerId  int
	}
)

// New creates a new logger.
func New(debugFlag bool) ILogger {
	return &logger{
		mu:        &sync.Mutex{},
		debugFlag: debugFlag,
	}
}

func (l *logger) Log(lines ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugFlag {
		fmt.Println(lines...)
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
