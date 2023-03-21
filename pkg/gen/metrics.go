package gen

import "sync"

type (
	// Metrics represents the generation metrics.
	Metrics interface {
		// Keys returns the list of package names for which metrics have been recorded.
		Keys() []string
		// Get returns the list of measurements for the specified package.
		Get(key string) []*Measurement
	}

	metrics struct {
		mu *sync.Mutex

		seen map[string][]*Measurement
	}

	Measurement struct {
		FileName string
		Path     string
		Created  bool
	}
)

var _ Metrics = (*metrics)(nil)

// newMetrics returns a new metrics instance.
func newMetrics(seen map[string][]*Measurement) *metrics {
	if seen == nil {
		seen = make(map[string][]*Measurement)
	}
	return &metrics{
		mu:   &sync.Mutex{},
		seen: seen,
	}
}

func (m *metrics) Measure(key string, mrt *Measurement) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.seen[key] = append(m.seen[key], mrt)
}

func (m *metrics) NewIntent(key string, mrt *Measurement) *MeasurementIntent {
	return &MeasurementIntent{
		mrt:    mrt,
		key:    key,
		parent: m,
	}
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

func (m *metrics) Get(key string) []*Measurement {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.seen[key]
}

type MeasurementIntent struct {
	mrt    *Measurement
	key    string
	parent *metrics
}

func (mi *MeasurementIntent) Measure(mrt *Measurement) {
	if fn := mi.mrt.FileName; fn != "" {
		mrt.FileName = fn
	}
	if path := mi.mrt.Path; path != "" {
		mrt.Path = path
	}
	mi.parent.Measure(mi.key, mrt)
}
