package modules

import (
	"sync"
)

type (
	IMetrics interface {
		CaptureJob(sk, pkg string, m MetricJob)
		GetJobsMetrics() map[string]interface{}
		CaptureWorkUnit(m MetricWorkUnit)
		GetWorkMetrics() map[int]int
	}

	MetricJob struct {
		FileAbsolutePath string
		FileCreated      bool
	}

	MetricWorkUnit struct {
		WorkerID int
	}

	metrics struct {
		mu *sync.Mutex

		jobsMap map[string]interface{}
		workMap map[int]int
	}
)

// NewMetrics returns a new instance of `IMetrics`.
func NewMetrics() IMetrics {
	return &metrics{
		mu:      &sync.Mutex{},
		jobsMap: map[string]interface{}{},
		workMap: map[int]int{},
	}
}

func (ms *metrics) CaptureJob(sk, pkg string, m MetricJob) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	type metricsByPackage = map[string][]MetricJob

	// Verbose code to avoid panics.
	// Go's laws of reflection: https://go.dev/blog/laws-of-reflection
	jm := ms.jobsMap
	if jm[sk] == nil {
		jm[sk] = make(metricsByPackage)
	}
	if jm[sk].(metricsByPackage)[pkg] == nil {
		jm[sk].(metricsByPackage)[pkg] = make([]MetricJob, 0)
	}
	jm[sk].(metricsByPackage)[pkg] = append(jm[sk].(metricsByPackage)[pkg], m)
}

func (ms *metrics) GetJobsMetrics() map[string]interface{} {
	return ms.jobsMap
}

func (ms *metrics) CaptureWorkUnit(m MetricWorkUnit) {
	ms.mu.Lock()
	defer ms.mu.Unlock()

	wID, wm := m.WorkerID, ms.workMap
	if v, ok := wm[wID]; ok {
		wm[wID] = v + 1
	} else {
		wm[wID] = 1
	}
}

func (ms *metrics) GetWorkMetrics() map[int]int {
	return ms.workMap
}
