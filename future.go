package gotaskflow

type Future[T any] struct {
	c chan T
}

func newFuture[T any]() *Future[T] {
	return &Future[T]{
		c: make(chan T, 0),
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
