package terminalview

import (
	"strings"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

// Size is the terminal grid size in character cells. It is transport-neutral
// and can be passed directly to PTY, SSH window-change, or ConPTY adapters.
type Size struct {
	Columns int
	Rows    int
}

// KeyEvent is a stable, testable projection of FLTK keyboard state.
type KeyEvent struct {
	Key   int
	Text  string
	State int
}

// UITerminalView is a native ANSI/VT presentation and input surface. It owns
// terminal parsing, scrollback, selection, focus, and key-sequence translation;
// it intentionally does not own a process or network connection.
type UITerminalView struct {
	v          view.UIView
	raw        *fltk_bridge.Terminal
	onInput    func([]byte)
	onResize   func(Size)
	lastSize   Size
	inputBound bool
}

func NewUITerminalView(r *foundation.Rect) *UITerminalView {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 640, Height: 400}
	}
	raw := fltk_bridge.NewTerminal(r.X, r.Y, r.Width, r.Height)
	// Fl_Terminal is an Fl_Group because it owns native scrollbars. Close its
	// implicit construction group immediately so subsequently created application
	// widgets cannot become accidental terminal children.
	raw.End()
	t := &UITerminalView{raw: raw}
	t.v.BindRaw(raw)
	t.v.SetAutomationRole("terminal").SetAutomationName("Terminal")
	t.v.SetAutomationValueHandler(func() (string, bool) { return t.Text(), true })
	t.v.On(fltk_bridge.PUSH, func(fltk_bridge.Event) bool {
		raw.TakeFocus()
		return false // preserve Fl_Terminal mouse selection handling
	})
	raw.SetResizeHandler(t.publishSize)
	return t
}

func (t *UITerminalView) View() *view.UIView {
	if t == nil {
		return nil
	}
	return &t.v
}

func (t *UITerminalView) Raw() *fltk_bridge.Terminal {
	if t == nil {
		return nil
	}
	return t.raw
}

func (t *UITerminalView) Feed(data []byte) {
	if t != nil && t.raw != nil {
		t.raw.AppendBytes(data)
	}
}

func (t *UITerminalView) Append(text string) { t.Feed([]byte(text)) }

func (t *UITerminalView) Text() string {
	if t == nil || t.raw == nil {
		return ""
	}
	return t.raw.Text()
}

func (t *UITerminalView) Clear() {
	if t != nil && t.raw != nil {
		t.raw.Clear()
	}
}

func (t *UITerminalView) ClearHistory() {
	if t != nil && t.raw != nil {
		t.raw.ClearHistory()
	}
}

func (t *UITerminalView) Reset() {
	if t != nil && t.raw != nil {
		t.raw.Reset()
	}
}

func (t *UITerminalView) SetFont(font fltk_bridge.Font) {
	if t != nil && t.raw != nil {
		t.raw.SetTextFont(font)
		t.publishSize()
	}
}

func (t *UITerminalView) SetFontSize(size int) {
	if t != nil && t.raw != nil {
		t.raw.SetTextSize(size)
		t.publishSize()
	}
}

func (t *UITerminalView) SetTextColor(color uint) {
	if t != nil && t.raw != nil {
		t.raw.SetTextColor(fltk_bridge.Color(color))
	}
}

func (t *UITerminalView) SetBackgroundColor(color uint) {
	if t != nil && t.raw != nil {
		t.raw.SetBackgroundColor(fltk_bridge.Color(color))
	}
}

func (t *UITerminalView) SetSelectionColors(foreground, background uint) {
	if t != nil && t.raw != nil {
		t.raw.SetSelectionColors(fltk_bridge.Color(foreground), fltk_bridge.Color(background))
	}
}

func (t *UITerminalView) SetHistoryRows(rows int) {
	if t != nil && t.raw != nil && rows >= 0 {
		t.raw.SetHistoryRows(rows)
	}
}

func (t *UITerminalView) SetMargins(left, top, right, bottom int) {
	if t != nil && t.raw != nil {
		t.raw.SetMargins(left, top, right, bottom)
		t.publishSize()
	}
}

func (t *UITerminalView) SetRedrawRate(seconds float32) {
	if t != nil && t.raw != nil && seconds > 0 {
		t.raw.SetRedrawRate(seconds)
	}
}

func (t *UITerminalView) Size() Size {
	if t == nil || t.raw == nil {
		return Size{}
	}
	return Size{Columns: t.raw.DisplayColumns(), Rows: t.raw.DisplayRows()}
}

// OnInput receives exact terminal bytes for printable/IME text, control keys,
// navigation keys, and paste commits. Returning ownership to the caller keeps
// PTY and SSH lifecycle outside the reusable native widget.
func (t *UITerminalView) OnInput(handler func([]byte)) {
	if t == nil {
		return
	}
	t.onInput = handler
	if t.inputBound {
		return
	}
	t.inputBound = true
	t.v.On(fltk_bridge.KEYDOWN, func(fltk_bridge.Event) bool {
		data, handled := EncodeKey(KeyEvent{Key: fltk_bridge.EventKey(), Text: fltk_bridge.EventText(), State: fltk_bridge.EventState()})
		if !handled || len(data) == 0 || t.onInput == nil {
			return false
		}
		t.onInput(append([]byte(nil), data...))
		return true
	})
	t.v.On(fltk_bridge.PASTE, func(fltk_bridge.Event) bool {
		if t.onInput == nil || fltk_bridge.EventText() == "" {
			return false
		}
		t.onInput([]byte(fltk_bridge.EventText()))
		return true
	})
}

func (t *UITerminalView) OnResize(handler func(Size)) {
	if t == nil {
		return
	}
	t.onResize = handler
	t.publishSize()
}

func (t *UITerminalView) publishSize() {
	if t == nil {
		return
	}
	size := t.Size()
	if size.Columns <= 0 || size.Rows <= 0 || size == t.lastSize {
		return
	}
	t.lastSize = size
	if t.onResize != nil {
		t.onResize(size)
	}
}

func (t *UITerminalView) SetAutomationID(id string) *UITerminalView {
	if t != nil {
		t.v.SetAutomationID(id)
	}
	return t
}

func (t *UITerminalView) SetAutomationName(name string) *UITerminalView {
	if t != nil {
		t.v.SetAutomationName(name)
	}
	return t
}

// EncodeKey translates native key events into conventional xterm input bytes.
// Ctrl+C/D/Z remain distinct ETX/EOT/SUB bytes. Ctrl+Shift+C is intentionally
// left to the native widget's selection-copy behavior.
func EncodeKey(event KeyEvent) ([]byte, bool) {
	ctrl := event.State&fltk_bridge.CTRL != 0
	shift := event.State&fltk_bridge.SHIFT != 0
	alt := event.State&fltk_bridge.ALT != 0
	if ctrl && shift && (event.Key == int('c') || event.Key == int('C')) {
		return nil, false
	}

	var data []byte
	switch event.Key {
	case fltk_bridge.ENTER_KEY:
		data = []byte{'\r'}
	case fltk_bridge.TAB:
		if shift {
			data = []byte("\x1b[Z")
		} else {
			data = []byte{'\t'}
		}
	case fltk_bridge.BACKSPACE:
		data = []byte{0x7f}
	case fltk_bridge.ESCAPE:
		data = []byte{0x1b}
	case fltk_bridge.UP:
		data = []byte("\x1b[A")
	case fltk_bridge.DOWN:
		data = []byte("\x1b[B")
	case fltk_bridge.RIGHT:
		data = []byte("\x1b[C")
	case fltk_bridge.LEFT:
		data = []byte("\x1b[D")
	case fltk_bridge.HOME:
		data = []byte("\x1b[H")
	case fltk_bridge.END:
		data = []byte("\x1b[F")
	case fltk_bridge.INSERT:
		data = []byte("\x1b[2~")
	case fltk_bridge.DELETE:
		data = []byte("\x1b[3~")
	case fltk_bridge.PAGE_UP:
		data = []byte("\x1b[5~")
	case fltk_bridge.PAGE_DOWN:
		data = []byte("\x1b[6~")
	default:
		if event.Key >= fltk_bridge.F1 && event.Key <= fltk_bridge.F4 {
			data = []byte{0x1b, 'O', byte('P' + event.Key - fltk_bridge.F1)}
		} else if event.Key >= fltk_bridge.F5 && event.Key <= fltk_bridge.F12 {
			codes := [...]string{"15", "17", "18", "19", "20", "21", "23", "24"}
			data = []byte("\x1b[" + codes[event.Key-fltk_bridge.F5] + "~")
		} else if ctrl {
			key := event.Key
			if key >= 'a' && key <= 'z' {
				key -= 'a' - 'A'
			}
			if key >= '@' && key <= '_' {
				data = []byte{byte(key & 0x1f)}
			}
		} else if event.Text != "" && !strings.ContainsRune(event.Text, '\x00') {
			data = []byte(event.Text)
		}
	}
	if len(data) == 0 {
		return nil, false
	}
	if alt && data[0] != 0x1b {
		data = append([]byte{0x1b}, data...)
	}
	return data, true
}
