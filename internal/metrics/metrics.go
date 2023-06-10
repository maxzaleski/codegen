package metrics

import "sync"

type (
	IMetrics interface {
		// Keys returns the list of package names for which metrics have been recorded.
		Keys() []string
		// Get returns the list of measurements for the specified package.
		Get(key string) []*Measurement
	}

	Metrics struct {
		mu *sync.Mutex

		seen map[string][]*Measurement
	}

	Measurement struct {
		FileAbsolutePath string
		ScopeKey         string
		Created          bool
	}
)

var _ IMetrics = (*Metrics)(nil)

func New(seen map[string][]*Measurement) *Metrics {
	if seen == nil {
		seen = make(map[string][]*Measurement)
	}
	return &Metrics{
		mu:   &sync.Mutex{},
		seen: seen,
	}
}

func (m *Metrics) Measure(key string, mrt *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seen[key] = append(m.seen[key], mrt)
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

func (m *Metrics) Get(key string) []*Measurement {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.seen[key]
}
