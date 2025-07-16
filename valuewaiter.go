// Package valuewaiter provides a synchronization primitive that allows
// goroutines to wait for a specific value to be set.
package valuewaiter

import (
	"context"
	"sync"
)

// ValueWaiter is a synchronization primitive that allows goroutines to wait for
// a specific value to be set. It is useful for cases where you want to wait for
// a value to change before proceeding, without busy-waiting.
type ValueWaiter[T comparable] struct {
	c *sync.Cond
	v T
}

// NewValueWaiter creates a new ValueWaiter with an initial value.
func New[T comparable](initial T) *ValueWaiter[T] {
	return &ValueWaiter[T]{
		c: &sync.Cond{L: &sync.Mutex{}},
		v: initial,
	}
}

// WaitValue blocks until the ValueWaiter is set to the specified value.
func (vw *ValueWaiter[T]) WaitValue(v T) {
	vw.c.L.Lock()
	defer vw.c.L.Unlock()
	for {
		if v == vw.v {
			return
		}
		vw.c.Wait()
	}
}

// WaitValueContext blocks until the ValueWaiter is set to the specified value
// or the context is cancelled. If the context is cancelled, it returns the
// context error, otherwise nil.
func (vw *ValueWaiter[T]) WaitValueContext(ctx context.Context, v T) error {
	vw.c.L.Lock()
	defer vw.c.L.Unlock()
	stop := context.AfterFunc(ctx, func() {
		vw.c.L.Lock()
		defer vw.c.L.Unlock()
		vw.c.Broadcast()
	})
	defer stop()
	for {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if v == vw.v {
			return nil
		}
		vw.c.Wait()
	}
}

// SetValue sets the value of the ValueWaiter and unblocks all
// calls to WaitValue or WaitValueContext that are waiting for the
// specified value.
func (vw *ValueWaiter[T]) SetValue(v T) {
	vw.c.L.Lock()
	defer vw.c.L.Unlock()
	if v == vw.v {
		return
	}
	vw.v = v
	vw.c.Broadcast()
}

// GetValue returns the current value of the ValueWaiter.
func (vw *ValueWaiter[T]) GetValue() T {
	vw.c.L.Lock()
	defer vw.c.L.Unlock()
	return vw.v
}
