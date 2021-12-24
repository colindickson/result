package result

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

var NoneOptionError = errors.New("value is none")
var UnsetOptionError = errors.New("option not yet set")

type OptionKind int

const (
	Some   = OptionKind(0)
	None   = OptionKind(1)
	Error  = OptionKind(2)
	NotSet = OptionKind(9999)

	neverSet = OptionKind(-1) //used for testing
)

type Option[T any] struct {
	some T
	err  error

	isSet  bool
	isNone bool

	setterMux sync.RWMutex
	setOnce   sync.Once

	setChan chan struct{}
}

func NewOption[T any]() *Option[T] {
	setChan := make(chan struct{})
	return &Option[T]{
		setChan: setChan,
	}
}

func (o *Option[T]) SetSome(value T) *Option[T] {
	o.setterMux.RLock()
	if o.isSet {
		o.setterMux.RUnlock()
		return o
	}
	o.setterMux.RUnlock()

	o.setOnce.Do(func() {
		o.setterMux.Lock()
		defer o.setterMux.Unlock()

		o.isSet = true
		o.some = value

		close(o.setChan)
	})

	return o
}

func (o *Option[T]) SetNone() *Option[T] {
	o.setterMux.RLock()
	if o.isSet {
		o.setterMux.RUnlock()
		return o
	}
	o.setterMux.RUnlock()

	o.setOnce.Do(func() {
		o.setterMux.Lock()
		defer o.setterMux.Unlock()

		var emptyT T
		o.some = emptyT

		o.isSet = true
		o.isNone = true

		close(o.setChan)
	})

	return o
}

func (o *Option[T]) SetError(e error) *Option[T] {
	o.setterMux.RLock()
	if o.isSet {
		o.setterMux.RUnlock()
		return o
	}
	o.setterMux.RUnlock()

	o.setOnce.Do(func() {
		o.setterMux.Lock()
		defer o.setterMux.Unlock()

		var emptyT T
		o.some = emptyT

		o.isSet = true
		o.isNone = true
		o.err = e

		close(o.setChan)
	})

	return o
}

func (o *Option[T]) Unwrap() (T, error) {
	o.setterMux.RLock()
	defer o.setterMux.RUnlock()

	if !o.isSet {
		var emptyT T
		return emptyT, UnsetOptionError
	}

	if o.err != nil {
		var emptyT T
		return emptyT, o.err
	}

	if o.isNone {
		var emptyT T
		return emptyT, NoneOptionError
	}

	return o.some, nil
}

func (o *Option[T]) Wait(ctx context.Context) (T, error) {
	select {
	case <-ctx.Done():
		var emptyT T
		return emptyT, fmt.Errorf("wait context error: %w", ctx.Err())
	case <-o.setChan:
		if o.err != nil {
			var emptyT T
			return emptyT, o.err
		}

		if o.isNone {
			var emptyT T
			return emptyT, NoneOptionError
		}

		return o.some, nil
	}
}

func (o *Option[T]) Kind() OptionKind {
	o.setterMux.RLock()
	defer o.setterMux.RUnlock()

	if !o.isSet {
		return NotSet
	}

	if o.err != nil {
		return Error
	}

	if o.isNone {
		return None
	}

	return Some
}
