package terminalview

import (
	"fmt"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
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

	hint := label.NewUILabel(&foundation.Rect{X: 24, Y: 48, Width: 852, Height: 24}, "Right-click for Copy/Find/Paste · Unicode-folded search follows live output · Shift+Home/End jumps")
	parent.AddSubview(hint)

	terminal := uikit.NewUITerminalView(&foundation.Rect{X: 24, Y: 80, Width: 852, Height: 450})
	terminal.SetHistoryRows(500)
	terminal.SetAutomationID("terminal.scrollback")
	parent.AddSubview(terminal)
	for row := 1; row <= 120; row++ {
		terminal.Append(fmt.Sprintf("\x1b[36m%03d\x1b[0m  deterministic scrollback sample · UTF-8 · Русский\r\n", row))
	}
	terminal.Append("\x1b[32mLIVE OUTPUT — Shift+Home shows row 001, Shift+End returns here\x1b[0m")

	status := label.NewUILabel(&foundation.Rect{X: 24, Y: 548, Width: 240, Height: 32}, "Live output")
	status.SetFrame(fltk_bridge.FLAT_BOX)
	status.SetBackgroundColor(uint(fltk_bridge.BACKGROUND_COLOR))
	status.View().SetAutomationID("terminal.search-status")
	parent.AddSubview(status)
	terminal.OnInput(func(data []byte) { status.SetText(fmt.Sprintf("PTY input: %q", data)) })
	updateStatus := func() { status.SetText(fmt.Sprintf("History offset: %d rows", terminal.ScrollOffset())) }
	terminal.ObserveTextChanged(func() {
		if count := len(terminal.SearchText("οσ")); count > 0 {
			status.SetText(fmt.Sprintf("Unicode matches: %d", count))
		}
	})
	findOldest := func() {
		matches := terminal.SearchText("001")
		if len(matches) == 0 {
			status.SetText("Row 001 not found")
			return
		}
		terminal.RevealTextMatch(matches[0])
		status.SetText("Found row 001 in scrollback")
	}
	contextMenu := uikit.NewUIContextMenu(&foundation.Rect{})
	parent.AddSubview(contextMenu)
	terminal.OnContextMenu(func(state uikit.ContextMenuState) {
		copyFlags := 0
		if !state.HasSelection {
			copyFlags = fltk_bridge.MENU_INACTIVE
		}
		contextMenu.SetMenu([]uikit.MenuItem{
			{Title: "Copy	Ctrl+Shift+C", Flags: copyFlags, Callback: func() { terminal.CopySelection() }},
			{Title: "Copy All Output", Callback: func() { terminal.CopyAllText() }},
			{Title: "Paste	Ctrl+Shift+V", Callback: terminal.PasteClipboard},
			{Title: "Find Row 001", Callback: findOldest},
		})
		contextMenu.Popup()
	})

	find := button.NewUIButton(&foundation.Rect{X: 274, Y: 544, Width: 146, Height: 36}, "Find Row 001")
	find.View().SetAutomationID("terminal.find-oldest")
	find.OnTouchUpInside(findOldest)
	parent.AddSubview(find)
	top := button.NewUIButton(&foundation.Rect{X: 430, Y: 544, Width: 130, Height: 36}, "Oldest")
	top.OnTouchUpInside(func() { terminal.ScrollToTop(); updateStatus() })
	parent.AddSubview(top)
	liveMatchNumber := 0
	appendMatch := button.NewUIButton(&foundation.Rect{X: 570, Y: 544, Width: 130, Height: 36}, "Append Match")
	appendMatch.View().SetAutomationID("terminal.append-match")
	appendMatch.OnTouchUpInside(func() {
		liveMatchNumber++
		terminal.Append(fmt.Sprintf("\r\n\x1b[33mSIGMA %d · ΟΣ / ος / οσ\x1b[0m", liveMatchNumber))
	})
	parent.AddSubview(appendMatch)
	bottom := button.NewUIButton(&foundation.Rect{X: 710, Y: 544, Width: 166, Height: 36}, "Live Output")
	bottom.OnTouchUpInside(func() { terminal.ScrollToBottom(); updateStatus() })
	parent.AddSubview(bottom)

	return terminal
}
