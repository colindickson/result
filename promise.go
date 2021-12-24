package result

import (
	"context"
	"fmt"
)

type Promise[T any] struct {
	done chan struct{}

	result T
	err    error
}

func NewPromise[T any](ctx context.Context, f func(ctx context.Context) (T, error)) *Promise[T] {
	done := make(chan struct{})

	p := Promise[T]{
		done: done,
	}

	go func() {
		defer close(done)
		p.result, p.err = f(ctx)
	}()

	return &p
}

func (p *Promise[T]) Wait(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var empty T
		return empty, fmt.Errorf("wait context error: %w", ctx.Err())
	case <-p.done:
		break
	}

	return p.result, p.err
}
