package gen

import "sync"

type (
	Pool interface {
		// Acquire blocks until a seat is available.
		Acquire()
		// Release releases the current seat.
		Release()
	}

	pool struct {
		mu    *sync.Mutex
		wg    *sync.WaitGroup
		seats chan struct{}
	}
)

var _ Pool = (*pool)(nil)

func newPool(capacity int, wg *sync.WaitGroup) Pool {
	if wg == nil {
		wg = &sync.WaitGroup{}
	}
	return &pool{
		mu:    &sync.Mutex{},
		wg:    wg,
		seats: make(chan struct{}, capacity),
	}
}

func (p *pool) Acquire() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.seats <- struct{}{}
}

func (p *pool) Release() {
	p.mu.Lock()
	defer p.mu.Unlock()

	<-p.seats
}
