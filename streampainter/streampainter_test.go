package streampainter

import (
	"testing"

	"github.com/faith/color"
)

func TestStreamPainter(t *testing.T) {
	writer := NewStreamPainter(color.FgRed)
	_, err := writer.Write([]byte("should be in red"))
	if err != nil {
		t.Error("ERRORED")
	}
	if writer.paintColor != color.FgRed {
		t.Error("ERRORED")
	}
}
