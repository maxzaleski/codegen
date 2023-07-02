package slog

import (
	"fmt"
	"sync"
	"time"
)

type (
	ILogger interface {
		// Log wraps `fmt.Println`.
		Log(lines ...any)
		// Logf wraps `fmt.Println` with the given format.
		Logf(format string, a ...any)
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
			Atom(Red, "Debug flag is set (--debug=1); debug mode enabled, printing subsequent logs.\n"),
			Atom(Pink, "\n\tBe advised, this logger is called across goroutines, and as such logs may be in non-sequential order.\n"))
	}
	return l
}

func (l *logger) Log(lines ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugFlag {
		linesCopy := make([]interface{}, 1, len(lines)+1)
		linesCopy[0] = domain(Yellow, "start", "+"+time.Since(l.began).String())
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
