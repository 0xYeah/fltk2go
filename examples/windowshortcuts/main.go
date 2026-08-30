package main

import (
	"runtime"

	"github.com/0xdevelop/fltk2go"
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit"
)

func main() {
	runtime.LockOSThread()
	win := uikit.NewUIWindow(720, 280, "Window Shortcut Example")
	root := win.RootView()

	title := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 34, Width: 648, Height: 34}, "Window-scoped native shortcut")
	title.SetFontSize(22)
	root.AddSubview(title)

	hint := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 78, Width: 648, Height: 28}, "Focus the terminal button, then press Ctrl+K to search")
	root.AddSubview(hint)

	search := uikit.NewInput(36, 124, 648, 38, "")
	search.View().SetAutomationID("shortcut.search").SetAutomationName("Search saved connections")
	root.AddSubview(search)

	status := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 178, Width: 648, Height: 28}, "Waiting for Ctrl+K")
	status.SetFrame(fltk_bridge.FLAT_BOX)
	status.SetBackgroundColor(uint(fltk_bridge.BACKGROUND_COLOR))
	status.View().SetAutomationID("shortcut.status")
	root.AddSubview(status)

	other := uikit.NewUIButton(&foundation.Rect{X: 36, Y: 220, Width: 180, Height: 36}, "Unrelated control")
	root.AddSubview(other)

	win.OnShortcut(fltk_bridge.CTRL+int('k'), func() {
		if raw := search.View().Raw(); raw != nil {
			if focusable, ok := raw.(interface{ TakeFocus() int }); ok {
				focusable.TakeFocus()
			}
		}
		status.SetText("Search focused — type to continue")
	})

	win.Show()
	if raw := other.Raw(); raw != nil {
		raw.TakeFocus()
	}
	fltk2go.Run()
}
