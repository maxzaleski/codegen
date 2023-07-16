package metrics

import (
	"fmt"
	"github.com/maxzaleski/codegen/internal/lib/slice"
	"sync"
)

type (
	OutcomesByPkg = map[string][]FileOutcome

	IMetrics interface {
		// Keys returns the keys for the metricsMap.
		Keys() []string
		// Get returns the value for the specified key.
		Get(string string) interface{}
		// CaptureScope captures the generation outcome for the specified scope and package.
		CaptureScope(scope, pkg string, o FileOutcome)
		// CaptureWorker increments the jobs (performed) count for the specified worker ID.
		CaptureWorker(wID int)
	}

	Metrics struct {
		mu *sync.Mutex

		metricsMap map[string]interface{}
	}

	FileOutcome struct {
		AbsolutePath string
		Created      bool
	}
)

var _ IMetrics = (*Metrics)(nil)

func New() *Metrics {
	return &Metrics{
		mu:         &sync.Mutex{},
		metricsMap: map[string]interface{}{},
	}
}

func (m *Metrics) CaptureScope(scope, pkg string, o FileOutcome) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verbose code to avoid panics.
	// Go's laws of reflection: https://go.dev/blog/laws-of-reflection
	if m.metricsMap[scope] == nil {
		m.metricsMap[scope] = make(OutcomesByPkg)
	}
	if m.metricsMap[scope].(OutcomesByPkg)[pkg] == nil {
		m.metricsMap[scope].(OutcomesByPkg)[pkg] = make([]FileOutcome, 0)
	}
	m.metricsMap[scope].(OutcomesByPkg)[pkg] = append(m.metricsMap[scope].(OutcomesByPkg)[pkg], o)
}

func (m *Metrics) CaptureWorker(wID int) {
	m.mu.Lock()
	defer m.mu.Unlock()

	wIDs := fmt.Sprintf("worker_%d", wID)
	if m.metricsMap[wIDs] == nil {
		m.metricsMap[wIDs] = 0
	}
	m.metricsMap[wIDs] = m.metricsMap[wIDs].(int) + 1
}

func (m *Metrics) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	return slice.MapKeys(m.metricsMap)
}

func (m *Metrics) Get(key string) interface{} {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.metricsMap[key]
}
