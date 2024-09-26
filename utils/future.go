package utils

import "errors"

var ErrFutureClosed = errors.New("future has closed")

type Future[T any] struct {
	c chan T
}

func NewFuture[T any]() Future[T] {
	return Future[T]{
		c: make(chan T),
	}
}

func (f *Future[T]) Set(result T) {
	f.c <- result
	close(f.c)
}

func (f *Future[T]) Get() (T, error) {
	result, ok := <-f.c
	if !ok {
		var zero T
		return zero, ErrFutureClosed
	}
	return result, nil
}
