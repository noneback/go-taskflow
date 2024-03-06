package gotaskflow

import (
	"context"
	"errors"
)

type TaskHandle func(ctx *context.Context)

type StatefulTaskHandle[T any] func(ctx *context.Context) *Future[T] // TODO: Not Now

var (
	ErrFutureClosed = errors.New("future already closed")
)
