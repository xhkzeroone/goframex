package async

import (
	"fmt"
	"time"
)

type Result[T any] struct {
	Value T
	Err   error
}

type Future[T any] struct {
	Done chan Result[T]
}

func Async[T any](f func() (T, error)) *Future[T] {
	fut := &Future[T]{Done: make(chan Result[T], 1)}

	go func() {
		defer func() {
			if r := recover(); r != nil {
				fut.Done <- Result[T]{Err: fmt.Errorf("panic: %v", r)}
			}
		}()

		res, err := f()
		fut.Done <- Result[T]{Value: res, Err: err}
	}()

	return fut
}

func (f *Future[T]) Wait(timeout time.Duration) (T, error) {
	var zero T
	select {
	case res := <-f.Done:
		return res.Value, res.Err
	case <-time.After(timeout):
		return zero, fmt.Errorf("timeout")
	}
}
