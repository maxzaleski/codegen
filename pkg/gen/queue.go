package gen

import (
	"github.com/codegen/internal/core"
	"sync"
)

type (
	IQueue interface {
		Enqueue(j *job)
		Stream() <-chan *job
		Close()
	}

	queue struct {
		mu *sync.Mutex
		_q chan *job
	}

	job struct {
		*core.ScopeJob

		ScopeKey         string
		Metadata         Metadata
		Package          *core.Package
		DisableTemplates bool
	}

	Metadata struct {
		core.Metadata

		ScopeKey     string
		AbsolutePath string
		Inline       bool
	}
)

var _ IQueue = (*queue)(nil)

func newPool(capacity int) IQueue {
	return &queue{
		mu: &sync.Mutex{},
		_q: make(chan *job, capacity),
	}
}

func (p *queue) Enqueue(j *job) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p._q <- j
}

func (p *queue) Stream() <-chan *job {
	return p._q
}

func (p *queue) Close() {
	close(p._q)
}
