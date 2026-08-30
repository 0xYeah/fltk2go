//go:build ignore

package main

import (
	"runtime"

	"examples/terminalview"
	"github.com/0xdevelop/fltk2go"
	"github.com/0xdevelop/fltk2go/uikit/window"
)

func main() {
	runtime.LockOSThread()
	win := window.NewUIWindow(900, 600, "Terminal Scrollback Example")
	terminal := terminalview.BuildView(win.RootView())
	win.Show()
	if terminal.Raw() != nil {
		terminal.Raw().TakeFocus()
	}
	fltk2go.Run()
}
