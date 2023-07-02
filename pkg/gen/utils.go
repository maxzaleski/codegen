package gen

import (
	"github.com/codegen/internal/core"
	"github.com/codegen/internal/utils/slice"
)

func mapToDomainScope(ts []*core.DomainScope, dt core.DomainType) []*domainScope {
	return slice.Map(ts, func(ds *core.DomainScope) *domainScope {
		return &domainScope{
			DomainScope: ds,
			DomainType:  dt,
		}
	})
}

const (
	fileOutcomeSuccess fileOutcome = "created"
	fileOutcomeIgnored fileOutcome = "already-exists"
)

type fileOutcome string

func logFileOutcome(l ILogger, o fileOutcome, j *genJob) {
	l.Ack("file", j, "status", string(o), "file", j.FileAbsolutePath)
}

func logCreatingDirectory(l ILogger, path string) {
	l.Log("dirwalk", "msg", "creating directory if not exist", "path", path)
}
