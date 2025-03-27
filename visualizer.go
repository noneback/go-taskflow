package gotaskflow

import (
	"errors"
	"io"
)

type Visualizer interface {
	// Visualize generate raw dag text in dot format and write to writer
	Visualize(tf *TaskFlow, writer io.Writer) error
}

var vizer = Visualizer(mockVizer{})

var ErrVisualizerNotSupport = errors.New("visualization not support")

type mockVizer struct {
}

// Visualize implements Visualizer.
func (v mockVizer) Visualize(tf *TaskFlow, writer io.Writer) error {
	return ErrVisualizerNotSupport
}
