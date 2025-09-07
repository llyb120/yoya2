package y

import "sync"

type Pool[T any] struct {
	pool sync.Pool
	// new       func() T
	finalizer func(T)
}

func NewPool[T any](new func() T, finalizer func(T)) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return new()
			},
		},
		// new:       new,
		finalizer: finalizer,
	}
}

func (p *Pool[T]) Get() (T, func()) {
	x := p.pool.Get().(T)
	return x, func() {
		if p.finalizer != nil {
			p.finalizer(x)
		}
		p.pool.Put(x)
	}
}
