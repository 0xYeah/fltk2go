//go:build ignore

package main

import (
	"runtime"

	"examples/comprehensive"
	"github.com/0xdevelop/fltk2go"
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit/window"
)

func main() {
	runtime.LockOSThread()
	fltk_bridge.EnableTooltips()
	fltk_bridge.SetTooltipDelay(0.4)

	win := window.NewUIWindow(1060, 740, "FLTK2Go Comprehensive Example")
	root := win.RootView()

	comprehensive.BuildView(root)

	win.Show()
	fltk2go.Run()
}
