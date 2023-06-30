package gen

import (
	"fmt"
	"github.com/codegen/internal/core/slog"
	"github.com/codegen/internal/utils/slice"
	"github.com/codegen/internal/utils/terminal"
	"time"
)

type (
	ilogger interface {
		EventFile(outcome string, j *genJob)
		Start()
		Exit()
		Ack(j *genJob)
	}

	logger struct {
		began    time.Time
		parent   slog.ILogger
		workerId int
		scopeKey string
	}
)

// New creates a new logger.
func newLogger(parent slog.ILogger, t time.Time, workerId int) ilogger {
	return &logger{
		parent:   parent,
		workerId: workerId,
		began:    t,
	}
}

func (l *logger) log(event string, lines ...interface{}) {
	eventColour := terminal.Blue
	if event == "ack" {
		eventColour = terminal.Yellow
	}

	linesCopy := make([]interface{}, 1, len(lines)+1)
	linesCopy[0] = fmt.Sprintf("[start:%s] %s [event:%s]",
		terminal.Atom(terminal.Yellow, "+"+time.Since(l.began).String()),
		terminal.Atom(terminal.Purple, fmt.Sprintf("[worker_%d]", l.workerId)),
		terminal.Atom(eventColour, event),
	)
	linesCopy = append(linesCopy, lines...)
	l.parent.Log(linesCopy...)
}

func (l *logger) EventFile(outcome string, j *genJob) {
	tokens := slice.Map([]string{j.Metadata.ScopeKey, j.Key, j.FileAbsolutePath, outcome}, func(t string) any {
		return terminal.Atom(terminal.Cyan, t)
	})
	s := fmt.Sprintf("scopeKey=%s jobKey=%s file=%s status=%s", tokens...)
	l.log("file", s)
}

func (l *logger) Start() {
	l.log("system", "started")
}

func (l *logger) Exit() {
	l.log("system", "exited")
}

func (l *logger) Ack(j *genJob) {
	s := fmt.Sprintf("scopeKey=%s jobKey=%s unique=%v file=%s", j.Metadata.ScopeKey, j.Key, j.Unique, j.FileAbsolutePath)
	l.log("ack", s)
}
