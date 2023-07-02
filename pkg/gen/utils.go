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

func logFileOutcome(l ILogger, o fileOutcome, j *genJob) {
	l.Ack("file", j, "outcome", string(o), "file", j.FileAbsolutePath)
}
