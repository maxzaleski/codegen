package gen

import "sync"

type (
	Metrics interface {
		Keys() []string
		Get(key string) []*measurement
	}

	metrics struct {
		mu *sync.Mutex

		seen map[string][]*measurement
	}
)

type measurement struct {
	File    string
	Created bool
}

var _ Metrics = (*metrics)(nil)

// Measure reports the metrics for the specified package.
func (m *metrics) Measure(pkg string, mr *measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seen[pkg] = append(m.seen[pkg], mr)
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
