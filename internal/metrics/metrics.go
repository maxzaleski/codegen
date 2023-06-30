package metrics

import "sync"

type (
	IMetrics interface {
		// Keys returns the list of scopes for which metrics have been recorded.
		Keys() []string
		// Get returns the collection of measurements for the specified scope.
		Get(scope string) map[string][]*Measurement
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

	if m.seen[scope] == nil {
		m.seen[scope] = map[string][]*Measurement{}
	}
	if m.seen[scope].(map[string][]*Measurement)[pkg] == nil {
		m.seen[scope].(map[string][]*Measurement)[pkg] = make([]*Measurement, 0)
	}
	m.seen[scope].(map[string][]*Measurement)[pkg] = append(m.seen[scope].(map[string][]*Measurement)[pkg], mrt)
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

	return m.seen[key].(map[string][]*Measurement)
}
