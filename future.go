package pctk

import (
	"errors"
	"fmt"
	"time"
)

var (
	// PromiseBroken is an error that indicates that the promise is broken.
	PromiseBroken = errors.New("broken promise")
)

// Future is a value that will be available in the future.
type Future interface {
	// Wait waits for the future to be completed.
	Wait() (any, error)

	// IsCompleted returns true if the future is completed.
	IsCompleted() bool
}

// AlreadySucceeded creates a future that is already succeeded.
func AlreadySucceeded(v any) Future {
	return alreadySucceeded{v}
}

type alreadySucceeded struct{ any }

func (f alreadySucceeded) Wait() (any, error) {
	return f.any, nil
}

func (f alreadySucceeded) IsCompleted() bool {
	return true
}

// AlreadyFailed creates a future that is already failed.
func AlreadyFailed(err error) Future {
	return alreadyFailed{err}
}

type alreadyFailed struct{ error }

func (f alreadyFailed) Wait() (any, error) {
	return nil, error(error(f))
}

func (f alreadyFailed) IsCompleted() bool {
	return true
}

// Continue continues a future with another future. This will wait for future f and call g once f is
// completed, returning the result of g. If f is nil, g will be called with nil and its future will
// be returned. If f fails with an error, the resulting future will fail without calling g.
func Continue(f Future, g func(any) Future) Future {
	if f == nil {
		return g(nil)
	}

	prom := NewPromise()
	go func() {
		v, err := f.Wait()
		if err != nil {
			prom.CompleteWithError(err)
			return
		}
		v, err = g(v).Wait()
		prom.CompleteWith(v, err)
	}()
	return prom
}

// FutureMap maps the value of a future to another value. If the future fails, the resulting future
// will fail with the same error.
func FutureMap(f Future, g func(any) any) Future {
	return Continue(f, func(v any) Future {
		return AlreadySucceeded(g(v))
	})
}

// IgnoreError returns a future that will ignore the error of the given future f, replacing it with
// val if fails.
func IgnoreError(f Future, val any) Future {
	return RecoverWithValue(f, func(error) any { return val })
}

// Recover recovers from an error in a future. If the given future fails, the given function will be
// called with the error. If the function returns a future, it will be waited for and its value or
// its error will be returned.
func Recover(f Future, g func(error) Future) Future {
	prom := NewPromise()
	go func() {
		v, err := f.Wait()
		if err != nil {
			v, err = g(err).Wait()
		}
		prom.CompleteWith(v, err)
	}()
	return prom
}

// RecoverWithValue recovers from an error in a future.
func RecoverWithValue(f Future, g func(error) any) Future {
	return Recover(f, func(err error) Future {
		prom := NewPromise()
		prom.CompleteWithValue(g(err))
		return prom
	})
}

// Promise is an instant when some event will be produced.
type Promise struct {
	done   chan struct{}
	result any
	err    error
}

// NewPromise creates a new future.
func NewPromise() *Promise {
	done := make(chan struct{})
	return &Promise{done: done}
}

// Wait implements the Future interface.
func (f *Promise) Wait() (any, error) {
	<-f.done
	return f.result, f.err
}

// IsCompleted implements the Future interface.
func (f *Promise) IsCompleted() bool {
	select {
	case <-f.done:
		return true
	default:
		return false
	}
}

// Complete completes the future. This sets no value. The Wait function will return a zero value and
// no error.
func (f *Promise) Complete() {
	close(f.done)
}

// CompleteWith completes the future with the given value and error.
func (f *Promise) CompleteWith(v any, err error) {
	f.result = v
	f.err = err
	close(f.done)
}

// CompleteWithValue completes the future with a value.
func (f *Promise) CompleteWithValue(v any) {
	f.result = v
	close(f.done)
}

// CompleteWithError completes the future with an error.
func (f *Promise) CompleteWithError(err error) {
	f.err = err
	close(f.done)
}

// CompleteWithErrorf completes the future with an error formatted with the given format and args.
func (f *Promise) CompleteWithErrorf(format string, args ...any) {
	f.CompleteWithError(fmt.Errorf(format, args...))
}

// Bind binds the future to another future. The future will be completed with the value of the given
// future when it is completed.
func (f *Promise) Bind(other Future) {
	go func() {
		v, err := other.Wait()
		f.CompleteWith(v, err)
	}()
}

// Break breaks the promise. This will complete the future with a PromiseBroken error as value.
func (f *Promise) Break() {
	f.CompleteWithError(PromiseBroken)
}

// CompleteAfter completes the future after the given duration.
func (f *Promise) CompleteAfter(v any, d time.Duration) {
	if d == 0 {
		f.CompleteWithValue(v)
		return
	}
	time.AfterFunc(d, func() {
		f.CompleteWithValue(v)
	})
}
