package slog

type (
	ILogger interface {
		Log(msg string)
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
