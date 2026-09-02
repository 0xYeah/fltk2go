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
	win := uikit.NewUIWindow(720, 420, "Window Shortcut Example")
	root := win.RootView()

	title := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 34, Width: 648, Height: 34}, "Window-scoped native shortcut")
	title.SetFontSize(22)
	root.AddSubview(title)

	hint := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 78, Width: 648, Height: 28}, "Ctrl+K focuses search; Up/Down, Enter and Escape stay native and deterministic")
	root.AddSubview(hint)

	search := uikit.NewInput(36, 124, 648, 38, "")
	search.View().SetAutomationID("shortcut.search").SetAutomationName("Search saved connections")
	root.AddSubview(search)

	status := uikit.NewUILabel(&foundation.Rect{X: 36, Y: 178, Width: 648, Height: 28}, "Waiting for Ctrl+K")
	status.SetFrame(fltk_bridge.FLAT_BOX)
	status.SetBackgroundColor(uint(fltk_bridge.BACKGROUND_COLOR))
	status.View().SetAutomationID("shortcut.status")
	root.AddSubview(status)
	results := []string{"Production SSH", "Staging SSH", "Local Shell"}
	selected := 0
	search.OnNavigation(func(action uikit.InputNavigationAction) bool {
		switch action {
		case uikit.InputNavigationNext:
			if selected < len(results)-1 {
				selected++
			}
			status.SetText("Selected: " + results[selected])
		case uikit.InputNavigationPrevious:
			if selected > 0 {
				selected--
			}
			status.SetText("Selected: " + results[selected])
		case uikit.InputNavigationSubmit:
			status.SetText("Launched: " + results[selected])
		case uikit.InputNavigationCancel:
			if search.Text() == "" {
				return false
			}
			search.SetText("")
			status.SetText("Search cleared")
		}
		return true
	})

	terminal := uikit.NewUITerminalView(&foundation.Rect{X: 36, Y: 220, Width: 648, Height: 164})
	terminal.SetAutomationID("shortcut.terminal")
	terminal.Append("Terminal focused. Ctrl+K must not appear in this stream.\n")
	root.AddSubview(terminal)

	focusSearch := func() {
		if raw := search.View().Raw(); raw != nil {
			if focusable, ok := raw.(interface{ TakeFocus() int }); ok {
				focusable.TakeFocus()
			}
		}
		status.SetText("Search focused — type to continue")
	}
	win.OnShortcut(fltk_bridge.CTRL+int('k'), focusSearch)
	terminal.OnShortcut(fltk_bridge.CTRL+int('k'), focusSearch)
	terminal.OnInput(func(data []byte) {
		status.SetText("Unexpected terminal input: " + string(data))
	})

	win.Show()
	if raw := terminal.Raw(); raw != nil {
		raw.TakeFocus()
	}
	fltk2go.Run()
}
