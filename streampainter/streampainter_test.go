package streampainter

import (
	"testing"

	"github.com/faith/color"
)

func TestStreamPainter(t *testing.T) {
	writer := NewStreamPainter(color.FgRed)
	_, err := writer.Write([]byte("should be in red"))
	if err != nil {
		t.Errorf("Error should be 'nil', but was %v", err.Error())
	}
	if writer.paintColor != color.FgRed {
		t.Errorf("Assigned color expected FgRed, but was %v", writer.paintColor)
	}
}
