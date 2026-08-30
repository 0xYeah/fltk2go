package terminalview

import (
	"fmt"

	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit"
	"github.com/0xdevelop/fltk2go/uikit/button"
	"github.com/0xdevelop/fltk2go/uikit/label"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

// BuildView demonstrates native scrollback navigation without owning a PTY.
// Shift+PageUp/PageDown moves by a viewport and Shift+Home/End jumps between
// the oldest history and live output. Applications can use the same public
// methods from buttons or command palettes.
func BuildView(parent *view.UIView) *uikit.UITerminalView {
	title := label.NewUILabel(&foundation.Rect{X: 24, Y: 18, Width: 852, Height: 30}, "Terminal Scrollback")
	title.SetFontSize(22)
	parent.AddSubview(title)

	hint := label.NewUILabel(&foundation.Rect{X: 24, Y: 48, Width: 852, Height: 24}, "Shift+PageUp/PageDown scrolls · Shift+Home/End jumps to history/live output")
	parent.AddSubview(hint)

	terminal := uikit.NewUITerminalView(&foundation.Rect{X: 24, Y: 80, Width: 852, Height: 450})
	terminal.SetHistoryRows(500)
	terminal.SetAutomationID("terminal.scrollback")
	parent.AddSubview(terminal)
	for row := 1; row <= 120; row++ {
		terminal.Append(fmt.Sprintf("\x1b[36m%03d\x1b[0m  deterministic scrollback sample · UTF-8 · Русский\r\n", row))
	}
	terminal.Append("\x1b[32mLIVE OUTPUT — Shift+Home shows row 001, Shift+End returns here\x1b[0m")

	status := label.NewUILabel(&foundation.Rect{X: 24, Y: 548, Width: 360, Height: 32}, "Live output")
	parent.AddSubview(status)
	terminal.OnInput(func(data []byte) { status.SetText(fmt.Sprintf("PTY input: %q", data)) })
	updateStatus := func() { status.SetText(fmt.Sprintf("History offset: %d rows", terminal.ScrollOffset())) }

	top := button.NewUIButton(&foundation.Rect{X: 430, Y: 544, Width: 130, Height: 36}, "Oldest")
	top.OnTouchUpInside(func() { terminal.ScrollToTop(); updateStatus() })
	parent.AddSubview(top)
	pageDown := button.NewUIButton(&foundation.Rect{X: 570, Y: 544, Width: 130, Height: 36}, "Page Down")
	pageDown.OnTouchUpInside(func() { terminal.ScrollByRows(-max(1, terminal.Size().Rows-1)); updateStatus() })
	parent.AddSubview(pageDown)
	bottom := button.NewUIButton(&foundation.Rect{X: 710, Y: 544, Width: 166, Height: 36}, "Live Output")
	bottom.OnTouchUpInside(func() { terminal.ScrollToBottom(); updateStatus() })
	parent.AddSubview(bottom)

	return terminal
}
