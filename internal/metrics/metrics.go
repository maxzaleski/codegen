package metrics

import "sync"

type (
	IMetrics interface {
		// Keys returns the list of scopes for which metrics have been recorded.
		Keys() []string
		// Get returns the list of measurements for the specified scope.
		Get(key string) map[string][]*Measurement
	}

	Metrics struct {
		mu *sync.Mutex

		seen map[string]map[string][]*Measurement
	}

	Measurement struct {
		FileAbsolutePath string
		Created          bool
	}
)

var _ IMetrics = (*Metrics)(nil)

func New(seen map[string]map[string][]*Measurement) *Metrics {
	if seen == nil {
		seen = make(map[string]map[string][]*Measurement)
	}
	return &Metrics{
		mu:   &sync.Mutex{},
		seen: seen,
	}
}

func (m *Metrics) Measure(scope, pkg string, mrt *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seen[scope][pkg] = append(m.seen[scope][pkg], mrt)
}

func (m *Metrics) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]string, 0, len(m.seen))
	for k := range m.seen {
		keys = append(keys, k)
	}
	return keys
}

func (m *Metrics) Get(key string) map[string][]*Measurement {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.seen[key]
}
