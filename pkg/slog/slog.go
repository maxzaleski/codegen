package slog

import "fmt"

type (
	ILogger interface {
		// Log wraps `fmt.Println`.
		Log(msg string)
		// Logf wraps `fmt.Println` with the given format.
		Logf(format string, a ...any)
	}

	logger struct {
		debugFlag bool
	}
)

// New creates a new logger.
func New(debugFlag bool) ILogger {
	return &logger{
		debugFlag: debugFlag,
	}
}

func (l *logger) Log(msg string) {
	if l.debugFlag {
		fmt.Println(msg)
	}
}

func (l *logger) Logf(format string, a ...any) {
	if l.debugFlag {
		s := fmt.Sprintf(format, a)
		fmt.Println(s)
	}
}
