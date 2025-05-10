package gotaskflow

import (
	"io"
)

var dot = dotVizer{}

type Visualizer interface {
	// Visualize generate raw dag text in dot format and write to writer
	Visualize(tf *TaskFlow, writer io.Writer) error
}
