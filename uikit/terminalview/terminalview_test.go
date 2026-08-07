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
