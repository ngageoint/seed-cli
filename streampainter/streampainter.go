package streampainter

import (
	"os"

	"github.com/fatih/color"
)

type StreamPainter struct {
	paintColor color.Attribute
}

/*
 * NewStreamPainter returns new streampainter
 */
func NewStreamPainter(textColor color.Attribute) *StreamPainter {
	return &StreamPainter{textColor}
}

func (w *StreamPainter) Write(p []byte) (int, error) {
	n := len(p)
	color.New(w.paintColor).Fprintf(os.Stderr, string(p[:n]))
	return n, nil
}
