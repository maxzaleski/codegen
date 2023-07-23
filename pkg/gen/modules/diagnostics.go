package modules

import (
	"github.com/maxzaleski/codegen/internal/core"
	"github.com/maxzaleski/codegen/internal/db"
	"strings"
)

type (
	IDiagnostics interface {
		// Prepare optimises the diagnostics phase by determining which packages should be checked per a job's override
		// rules.
		Prepare(ds []*core.DomainScope)
	}

	diagnostics struct {
		db db.IDatabase

		requiredAll      bool
		requiredMap      map[string]*core.ScopeJobOverride
		localSnapshotMap map[string]*snapshot
	}

	snapshot struct {
	}
)

func NewDiagnostics(db db.IDatabase) IDiagnostics {
	return &diagnostics{
		db:          db,
		requiredMap: map[string]*core.ScopeJobOverride{},
	}
}

func (c *diagnostics) Prepare(ds []*core.DomainScope) {
	for _, scope := range ds {
		for _, job := range scope.Jobs {
			for pkg, newO := range job.OverrideOn {
				// Check for wildcard '*' token = all packages.
				if pkg = strings.TrimPrefix(pkg, "\\"); pkg == "*" {
					c.requiredAll = true
					c.requiredMap = nil
					return
				}
				// Otherwise, merge the two datasets.
				if o, ok := c.requiredMap[pkg]; ok {
					o.Merge(newO)
				}
			}
		}
	}
}
