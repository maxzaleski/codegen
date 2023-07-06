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
		// LogEnv logs an environment variable.
		//
		// format: "{descriptor} flag is set (-{flag}); {msg}"
		LogEnv(descriptor, flag string, msg string)
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
		l.LogEnv("Debug flag", "debug=1", "debug mode enabled, printing subsequent logs.")
		l.Log(
			Atom(Pink, "\n\n\tBe advised, this logger is called across goroutines, and as such logs may be in non-sequential order.\n"))
	}
	return l
}

func (l *logger) Log(lines ...any) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.debugFlag {
		linesCopy := make([]interface{}, 1, len(lines)+1)
		linesCopy[0] = domain(LightYellow, "start", "+"+time.Since(l.began).String())
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

func (l *logger) LogEnv(descriptor, flag string, msg string) {
	l.Log(Atom(Red, fmt.Sprintf("[env] %s flag is set (-%s); %s", descriptor, flag, msg)))
}
