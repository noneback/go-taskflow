package gotaskflow

import (
	"context"
	"errors"
)

type TaskHandle func(ctx *context.Context)

type ConditionTaskHandle StatefulTaskHandle[int]

type StatefulTaskHandle[T any] func(ctx *context.Context) T // TODO: Not Now

var (
	ErrFutureClosed = errors.New("future already closed")
)
