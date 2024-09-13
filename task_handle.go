package gotaskflow

import (
	"context"
	"errors"
)

type TaskHandle func(ctx *context.Context)

var (
	ErrFutureClosed = errors.New("future already closed")
)
