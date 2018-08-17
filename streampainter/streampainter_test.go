package streampainter

import (
	"testing"

<<<<<<< HEAD
	"github.com/fatih/color"
=======
	"github.com/faith/color"
>>>>>>> 9060b75... added red text for stderr output
)

func TestStreamPainter(t *testing.T) {
	writer := NewStreamPainter(color.FgRed)
	_, err := writer.Write([]byte("should be in red"))
	if err != nil {
<<<<<<< HEAD
		t.Errorf("Error should be 'nil', but was %v", err.Error())
	}
	if writer.paintColor != color.FgRed {
		t.Errorf("Assigned color expected FgRed, but was %v", writer.paintColor)
=======
		t.Error("ERRORED")
	}
	if writer.paintColor != color.FgRed {
		t.Error("ERRORED")
>>>>>>> 9060b75... added red text for stderr output
	}
}
