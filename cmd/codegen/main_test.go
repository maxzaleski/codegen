package main

import (
	"testing"

	"github.com/codegen/pkg/gen"
)

func TestOutputMetrics(t *testing.T) {
	t.Skip("Visual inspection only")

	t.Run("empty", func(t *testing.T) {
		m := &mockMetrics{
			seen: map[string][]*gen.Measurement{
				"todo": {
					{Key: "service.go", Created: false},
					{Key: "service_logger.go", Created: false},
					{Key: "repository.go", Created: false},
					{Key: "repository_logger.go", Created: false},
				},
				"user": {
					{Key: "service.go", Created: false},
					{Key: "service_logger.go", Created: false},
					{Key: "repository.go", Created: false},
					{Key: "repository_logger.go", Created: false},
				},
			},
		}
		outputMetrics(m)
	})

	t.Run("populated", func(t *testing.T) {
		m := &mockMetrics{
			seen: map[string][]*gen.Measurement{
				"todo": {
					{Key: "service.go", Created: false},
					{Key: "service_logger.go", Created: true},
					{Key: "repository.go", Created: false},
					{Key: "repository_logger.go", Created: true},
				},
				"user": {
					{Key: "service.go", Created: false},
					{Key: "service_logger.go", Created: true},
					{Key: "repository.go", Created: false},
					{Key: "repository_logger.go", Created: true},
				},
			},
		}
		outputMetrics(m)
	})
}

type mockMetrics struct {
	seen map[string][]*gen.Measurement
}

var _ gen.Metrics = (*mockMetrics)(nil)

func (m *mockMetrics) Get(key string) []*gen.Measurement {
	return m.seen[key]
}

func (m *mockMetrics) Keys() []string {
	keys := make([]string, 0, len(m.seen))
	for k := range m.seen {
		keys = append(keys, k)
	}
	return keys
}
