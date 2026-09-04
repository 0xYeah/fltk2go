package terminalview

import (
	"bytes"
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func TestTerminalStreamFilterStripsShellMetadataAndPreservesDisplayCSI(t *testing.T) {
	filter := terminalStreamFilter{}
	chunks := [][]byte{
		[]byte("before\x1b]0;ubuntu@host: ~"),
		[]byte("\x07\x1b[?2004h\x1b[32m简体·繁體·Русский"),
		[]byte("\x1b[0m\x1b[?2004l after"),
	}
	var got []byte
	for _, chunk := range chunks {
		got = append(got, filter.Filter(chunk)...)
	}
	want := []byte("before\x1b[32m简体·繁體·Русский\x1b[0m after")
	if !bytes.Equal(got, want) {
		t.Fatalf("filtered stream = %q, want %q", got, want)
	}
}

func TestTerminalStreamFilterHandlesSplitEscapeTerminators(t *testing.T) {
	filter := terminalStreamFilter{}
	var got []byte
	for _, chunk := range [][]byte{[]byte("a\x1b"), []byte("]2;title\x1b"), []byte("\\b\x1b["), []byte("31mred")} {
		got = append(got, filter.Filter(chunk)...)
	}
	if want := []byte("ab\x1b[31mred"); !bytes.Equal(got, want) {
		t.Fatalf("filtered split stream = %q, want %q", got, want)
	}
}

func TestEncodeKeyPreservesTerminalControlSemantics(t *testing.T) {
	tests := []struct {
		name    string
		event   KeyEvent
		want    []byte
		handled bool
	}{
		{name: "enter", event: KeyEvent{Key: fltk_bridge.ENTER_KEY}, want: []byte{'\r'}, handled: true},
		{name: "tab", event: KeyEvent{Key: fltk_bridge.TAB}, want: []byte{'\t'}, handled: true},
		{name: "shift tab", event: KeyEvent{Key: fltk_bridge.TAB, State: fltk_bridge.SHIFT}, want: []byte("\x1b[Z"), handled: true},
		{name: "ctrl c", event: KeyEvent{Key: 'c', State: fltk_bridge.CTRL}, want: []byte{0x03}, handled: true},
		{name: "ctrl d", event: KeyEvent{Key: 'd', State: fltk_bridge.CTRL}, want: []byte{0x04}, handled: true},
		{name: "ctrl z", event: KeyEvent{Key: 'z', State: fltk_bridge.CTRL}, want: []byte{0x1a}, handled: true},
		{name: "native copy", event: KeyEvent{Key: 'c', State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, handled: false},
		{name: "native paste", event: KeyEvent{Key: 'v', State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, handled: false},
		{name: "shift insert paste", event: KeyEvent{Key: fltk_bridge.INSERT, State: fltk_bridge.SHIFT}, handled: false},
		{name: "unicode commit", event: KeyEvent{Key: '中', Text: "中"}, want: []byte("中"), handled: true},
		{name: "alt unicode", event: KeyEvent{Key: 'ж', Text: "ж", State: fltk_bridge.ALT}, want: append([]byte{0x1b}, []byte("ж")...), handled: true},
		{name: "up", event: KeyEvent{Key: fltk_bridge.UP}, want: []byte("\x1b[A"), handled: true},
		{name: "delete", event: KeyEvent{Key: fltk_bridge.DELETE}, want: []byte("\x1b[3~"), handled: true},
		{name: "f12", event: KeyEvent{Key: fltk_bridge.F12}, want: []byte("\x1b[24~"), handled: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, handled := EncodeKey(test.event)
			if handled != test.handled || !bytes.Equal(got, test.want) {
				t.Fatalf("EncodeKey(%+v) = %v, %t; want %v, %t", test.event, got, handled, test.want, test.handled)
			}
		})
	}
}

func TestEncodeKeyRejectsNULTextAndUnknownKeys(t *testing.T) {
	for _, event := range []KeyEvent{{Key: 0}, {Key: 'x', Text: "a\x00b"}} {
		if got, handled := EncodeKey(event); handled || got != nil {
			t.Fatalf("EncodeKey(%+v) = %v, %t; want nil, false", event, got, handled)
		}
	}
}

func TestTerminalClipboardShortcutRouting(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  terminalClipboardAction
	}{
		{name: "copy", event: KeyEvent{Key: 'c', State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, want: terminalClipboardCopy},
		{name: "uppercase copy", event: KeyEvent{Key: 'C', State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, want: terminalClipboardCopy},
		{name: "paste", event: KeyEvent{Key: 'v', State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, want: terminalClipboardPaste},
		{name: "shift insert", event: KeyEvent{Key: fltk_bridge.INSERT, State: fltk_bridge.SHIFT}, want: terminalClipboardPaste},
		{name: "interrupt", event: KeyEvent{Key: 'c', State: fltk_bridge.CTRL}, want: terminalClipboardNone},
		{name: "quoted insert", event: KeyEvent{Key: fltk_bridge.INSERT, State: fltk_bridge.CTRL}, want: terminalClipboardNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := clipboardActionForKey(test.event); got != test.want {
				t.Fatalf("clipboardActionForKey(%+v) = %v, want %v", test.event, got, test.want)
			}
		})
	}
}

func TestTerminalScrollbackShortcutRouting(t *testing.T) {
	tests := []struct {
		name  string
		event KeyEvent
		want  terminalScrollAction
	}{
		{name: "page up", event: KeyEvent{Key: fltk_bridge.PAGE_UP, State: fltk_bridge.SHIFT}, want: terminalScrollPageUp},
		{name: "page down", event: KeyEvent{Key: fltk_bridge.PAGE_DOWN, State: fltk_bridge.SHIFT}, want: terminalScrollPageDown},
		{name: "history start", event: KeyEvent{Key: fltk_bridge.HOME, State: fltk_bridge.SHIFT}, want: terminalScrollTop},
		{name: "live bottom", event: KeyEvent{Key: fltk_bridge.END, State: fltk_bridge.SHIFT}, want: terminalScrollBottom},
		{name: "plain page up remains PTY input", event: KeyEvent{Key: fltk_bridge.PAGE_UP}, want: terminalScrollNone},
		{name: "ctrl shift page up remains available", event: KeyEvent{Key: fltk_bridge.PAGE_UP, State: fltk_bridge.CTRL | fltk_bridge.SHIFT}, want: terminalScrollNone},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := scrollActionForKey(test.event); got != test.want {
				t.Fatalf("scrollActionForKey(%+v) = %v, want %v", test.event, got, test.want)
			}
			_, handled := EncodeKey(test.event)
			if test.want != terminalScrollNone && handled {
				t.Fatalf("local scroll shortcut reached PTY encoding: %+v", test.event)
			}
			if test.want == terminalScrollNone && !handled {
				t.Fatalf("non-scroll shortcut was lost: %+v", test.event)
			}
		})
	}
}

func TestTerminalShortcutCommandsRunBeforePTYInput(t *testing.T) {
	terminal := NewUITerminalView(nil)
	shortcut := fltk_bridge.CTRL + int('k')
	commands := 0
	input := make([]byte, 0)
	terminal.OnShortcut(shortcut, func() { commands++ })
	terminal.OnInput(func(data []byte) { input = append(input, data...) })

	if !terminal.dispatchShortcut(func(candidate int) bool { return candidate == shortcut }) {
		t.Fatal("registered terminal shortcut was not dispatched")
	}
	if commands != 1 || len(input) != 0 {
		t.Fatalf("terminal shortcut leaked to PTY input: commands=%d input=%q", commands, input)
	}
	if terminal.dispatchShortcut(func(int) bool { return false }) {
		t.Fatal("unmatched key must remain available to terminal input")
	}
}

func TestTerminalContextMenuOnlyConsumesRightMousePush(t *testing.T) {
	terminal := NewUITerminalView(nil)
	invocations := 0
	var state ContextMenuState
	terminal.OnContextMenu(func(got ContextMenuState) {
		invocations++
		state = got
	})

	if terminal.dispatchContextMenu(fltk_bridge.LeftMouse) {
		t.Fatal("left mouse push must remain available to native terminal selection")
	}
	if !terminal.dispatchContextMenu(fltk_bridge.RightMouse) {
		t.Fatal("right mouse push did not open the terminal context menu")
	}
	if invocations != 1 || state.HasSelection {
		t.Fatalf("context menu state = invocations:%d selection:%t, want 1/false", invocations, state.HasSelection)
	}
}

func TestTerminalClipboardCommandsAreSafeWithoutSelection(t *testing.T) {
	terminal := NewUITerminalView(nil)
	if terminal.HasSelection() {
		t.Fatal("new terminal unexpectedly has a selection")
	}
	if terminal.CopySelection() {
		t.Fatal("copy without a selection must report no action")
	}
	if terminal.CopyAllText() {
		t.Fatal("copy-all without terminal text must report no action")
	}
}

func TestTerminalCopyAllTextCopiesCompleteRenderedOutput(t *testing.T) {
	terminal := NewUITerminalView(nil)
	terminal.Append("first line\r\nsecond line")
	if !terminal.CopyAllText() {
		t.Fatal("copy-all with terminal text must report success")
	}
}

func TestSearchTextFindsEveryCaseInsensitiveUnicodeOccurrence(t *testing.T) {
	matches := SearchText("Alpha needle\n简体 NEEDLE needle\nlast", "NeEdLe")
	if len(matches) != 3 {
		t.Fatalf("match count = %d, want 3: %#v", len(matches), matches)
	}
	want := []TextMatch{
		{Line: 0, Column: 6, Length: 6},
		{Line: 1, Column: 3, Length: 6},
		{Line: 1, Column: 10, Length: 6},
	}
	for i := range want {
		if matches[i] != want[i] {
			t.Fatalf("match %d = %#v, want %#v", i, matches[i], want[i])
		}
	}
}

func TestSearchTextTreatsQueryLiterallyAndRejectsBlankInput(t *testing.T) {
	if got := SearchText("a.b a-b", "."); len(got) != 1 || got[0].Column != 1 {
		t.Fatalf("literal dot matches = %#v, want one match at column 1", got)
	}
	if got := SearchText("text", "  "); got != nil {
		t.Fatalf("blank query matches = %#v, want nil", got)
	}
}

func TestTerminalSearchTextIncludesRetainedOutput(t *testing.T) {
	terminal := NewUITerminalView(nil)
	terminal.Append("older NEEDLE\r\nnewer needle")
	if got := terminal.SearchText("needle"); len(got) != 2 {
		t.Fatalf("terminal match count = %d, want 2: %#v", len(got), got)
	}
}

func TestTerminalRevealTextMatchCentersHistoryRow(t *testing.T) {
	terminal := NewUITerminalView(nil)
	terminal.SetHistoryRows(200)
	for i := 0; i < 80; i++ {
		terminal.Append("line\r\n")
	}
	matches := SearchText(terminal.Text(), "line")
	if len(matches) < 40 {
		t.Fatalf("match count = %d, want retained history", len(matches))
	}
	terminal.RevealTextMatch(matches[0])
	if terminal.ScrollOffset() == 0 {
		t.Fatal("revealing an old match did not move away from live output")
	}
	terminal.RevealTextMatch(matches[len(matches)-1])
	if terminal.ScrollOffset() != 0 {
		t.Fatalf("revealing newest match left scroll offset %d, want live bottom", terminal.ScrollOffset())
	}
}
