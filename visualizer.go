//go:build !without_visualizer

package gotaskflow

import "io"

type Visualizer interface {
	Visualize(tf *TaskFlow, writer io.Writer) error
}

var vizer = Visualizer(mockVizer{})
