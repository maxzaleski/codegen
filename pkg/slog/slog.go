package slog

type (
	Logger interface {
	}

	logger struct {
		debugFlag bool
	}
)

// New creates a new logger.
func New(debugFlag bool) Logger {
	return &logger{
		debugFlag: debugFlag,
	}
}
