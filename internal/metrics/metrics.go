package metrics

import "sync"

type (
	PackageMeasurements = map[string][]*Measurement

	IMetrics interface {
		// ScopeKeys returns the list of scopes for which metrics have been recorded.
		ScopeKeys() []string
		// GetPackageMeasurements returns the collection of measurements for the specified scope.
		GetPackageMeasurements(scope string) PackageMeasurements
	}

	Metrics struct {
		mu *sync.Mutex

		seen map[string]interface{}
	}

	Measurement struct {
		FileAbsolutePath string
		Created          bool
	}
)

var _ IMetrics = (*Metrics)(nil)

func New(seen map[string]interface{}) *Metrics {
	if seen == nil {
		seen = map[string]interface{}{}
	}
	return &Metrics{
		mu:   &sync.Mutex{},
		seen: seen,
	}
}

func (m *Metrics) Measure(scope, pkg string, mrt *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Verbose code to avoid panics.
	// Go's laws of reflection: https://go.dev/blog/laws-of-reflection
	if m.seen[scope] == nil {
		m.seen[scope] = make(PackageMeasurements)
	}
	if m.seen[scope].(PackageMeasurements)[pkg] == nil {
		m.seen[scope].(PackageMeasurements)[pkg] = make([]*Measurement, 0)
	}
	m.seen[scope].(PackageMeasurements)[pkg] = append(m.seen[scope].(PackageMeasurements)[pkg], mrt)
}

func (m *Metrics) ScopeKeys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]string, 0, len(m.seen))
	for k := range m.seen {
		keys = append(keys, k)
	}
	return keys
}

func (m *Metrics) GetPackageMeasurements(key string) PackageMeasurements {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.seen[key].(PackageMeasurements)
}
