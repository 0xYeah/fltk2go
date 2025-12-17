package window

import (
	"github.com/0xYeah/fltk2go/fltk_bridge"
	"github.com/0xYeah/fltk2go/ui/core"
)

type Window struct {
	fltk_bridge.Window
	frame *core.Frame
}

func NewWindow(frame *core.Frame, title string) *Window {
	win := &Window{
		frame: frame,
	}

	win := fltk_bridge.NewWindowWithPosition(frame.X, frame.Y, frame.Y, frame.Y, "flutter like example")
	return aWindow
}
