// Package promise mimics the Promise class from javascript
package promise

type Await[T any] struct {
	Done   chan Promise[T]
	Loaded bool
}

type Promise[T any] struct {
	Data *T
	Err  error
}

func NewAwait[T any]() *Await[T] {
	return &Await[T]{
		Done: make(chan Promise[T]),
	}
}

func Ok[T any](data *T) Promise[T] {
	return Promise[T]{
		Data: data,
		Err:  nil,
	}
}

func Fail[T any](err error) Promise[T] {
	return Promise[T]{
		Data: nil,
		Err:  nil,
	}
}
