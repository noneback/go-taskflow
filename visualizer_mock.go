package gotaskflow

import (
	"errors"
	"io"
)

var ErrVisualizerNotCompile = errors.New("feature not support, try to remove !noviz build tag instead")

type mockVizer struct {
}

// Visualize implements Visualizer.
func (v mockVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	return ErrVisualizerNotCompile
}
