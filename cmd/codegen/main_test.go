package main

import (
	"fmt"
	"testing"
	"time"

	"github.com/codegen/pkg/gen"
)

func TestOutputMetrics(t *testing.T) {
	// t.Skip("Visual inspection only")

	t.Run("empty", func(t *testing.T) {
		defer fmt.Println()
		m := &mockMetrics{
			seen: map[string][]*gen.Measurement{
				"todo": {
					{
						FileName: "service.go",
						Created:  false,
						Path:     "pkg/todo/service.go",
					},
					{
						FileName: "service_logger.go",
						Created:  false,
						Path:     "pkg/todo/service_logger.go",
					},
					{
						FileName: "repository.go",
						Created:  false,
						Path:     "pkg/todo/repository.go",
					},
					{
						FileName: "repository_logger.go",
						Created:  false,
						Path:     "pkg/todo/repository_logger.go",
					},
				},
				"user": {
					{
						FileName: "service.go",
						Created:  false,
						Path:     "pkg/user/service.go",
					},
					{
						FileName: "service_logger.go",
						Created:  false,
						Path:     "pkg/user/service_logger.go",
					},
					{
						FileName: "repository.go",
						Created:  false,
						Path:     "pkg/user/repository.go",
					},
					{
						FileName: "repository_logger.go",
						Created:  false,
						Path:     "pkg/user/repository_logger.go",
					},
				},
			},
		}
		outputMetrics(time.Now(), m)
	})

	t.Run("populated", func(t *testing.T) {
		defer fmt.Println()
		m := &mockMetrics{
			seen: map[string][]*gen.Measurement{
				"todo": {
					{FileName: "service.go", Created: false},
					{FileName: "service_logger.go", Created: true},
					{FileName: "repository.go", Created: false},
					{FileName: "repository_logger.go", Created: true},
				},
				"pkg/user": {
					{FileName: "service.go", Created: false},
					{FileName: "service_logger.go", Created: true},
					{FileName: "repository.go", Created: false},
					{FileName: "repository_logger.go", Created: true},
				},
			},
		}
		outputMetrics(time.Now(), m)
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
