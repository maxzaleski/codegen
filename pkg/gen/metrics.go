package gen

import "sync"

type (
	// Metrics represents the generation metrics.
	Metrics interface {
		// Keys returns the list of package names for which metrics have been recorded.
		Keys() []string
		// Get returns the list of measurements for the specified package.
		Get(key string) []*measurement
	}

	metrics struct {
		mu *sync.Mutex

		seen map[string][]*measurement
	}

	measurement struct {
		Key     string
		Created bool
	}
)

var _ Metrics = (*metrics)(nil)

func (m *metrics) Measure(pkg string, mrt *measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seen[pkg] = append(m.seen[pkg], mrt)
}

func (m *metrics) Keys() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	keys := make([]string, 0, len(m.seen))
	for k := range m.seen {
		keys = append(keys, k)
	}
	return keys
}

func (m *metrics) Get(key string) []*measurement {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.seen[key]
}
