package window

import (
	"github.com/0xYeah/fltk2go/core"
	"github.com/0xYeah/fltk2go/fltk_bridge"
)

type Window struct {
	fltk_bridge.Window
	frame *core.Frame
}

func NewWindow(frame core.Frame, title string) *Window {
	aWindow := &Window{
		frame: &frame,
	}
	return aWindow
}
