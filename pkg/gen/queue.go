package gen

import (
	"github.com/codegen/internal/core"
	"sync"
)

type (
	IQueue interface {
		Enqueue(j *genJob)
		Dequeue() (*genJob, bool)
		Size() int
		Close()
		// WaitReadiness blocks until the queue is ready.
		WaitReadiness()
		// Ready marks the queue as ready. This is a one-time operation and is safe to call multiple times.
		Ready()
	}

	queue struct {
		mu *sync.Mutex

		_q          chan *genJob
		_qOnce      sync.Once
		_qClosed    bool
		_qReady     chan bool
		_qReadyOnce sync.Once
	}

	genJob struct {
		*core.ScopeJob

		Metadata         Metadata
		Package          *core.Package
		DisableTemplates bool
	}

	Metadata struct {
		core.Metadata

		ScopeKey     string
		DomainType   core.DomainType
		AbsolutePath string
		Inline       bool
	}
)

var _ IQueue = (*queue)(nil)

func newPool(capacity int) IQueue {
	mu := &sync.Mutex{}
	return &queue{
		mu: mu,

		//_q:          make(chan *genJob, capacity),
		_q:          make(chan *genJob, 10),
		_qOnce:      sync.Once{},
		_qReady:     make(chan bool, 1),
		_qReadyOnce: sync.Once{},
	}
}

func (p *queue) Enqueue(j *genJob) {
	if p._qClosed {
		panic("queue is closed")
	}
	p._q <- j
}

func (p *queue) Dequeue() (*genJob, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	j := <-p._q
	if j == nil {
		if p._qClosed {
			return nil, false
		}
		panic("queue: race condition")
	}
	return j, true
}

func (p *queue) Size() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p._q)
}

func (p *queue) Ready() {
	p._qReadyOnce.Do(func() { p._qReady <- true })
}

func (p *queue) WaitReadiness() {
	<-p._qReady
}

func (p *queue) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p._qOnce.Do(func() {
		close(p._q)
		p._qClosed = true
	})
}
