package textview

import (
	"strings"
	"unicode/utf8"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

type UITextView struct {
	v             view.UIView
	raw           *fltk_bridge.TextEditor
	buffer        *fltk_bridge.TextBuffer
	styleBuffer   *fltk_bridge.TextBuffer
	font          fltk_bridge.Font
	fontSize      int
	textColor     fltk_bridge.Color
	fallbackFont  fltk_bridge.Font
	fallbackMatch func(rune) bool
	fallbackWatch bool
}

type KeyEvent struct {
	Key   int
	Text  string
	State int
}

func NewUITextView(r *foundation.Rect) *UITextView {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 240, Height: 120}
	}

	raw := fltk_bridge.NewTextEditor(r.X, r.Y, r.Width, r.Height)
	buffer := fltk_bridge.NewTextBuffer()
	raw.SetBuffer(buffer)
	raw.SetWrapMode(fltk_bridge.WRAP_AT_BOUNDS)

	t := &UITextView{
		raw: raw, buffer: buffer,
		font: fltk_bridge.HELVETICA, fontSize: 14, textColor: fltk_bridge.FOREGROUND_COLOR,
	}
	t.v.BindRaw(raw)
	t.v.SetAutomationRole("textbox").SetAutomationName("Text view")
	t.v.SetAutomationTextHandlers(func(text string) error {
		t.SetText(text)
		return nil
	}, func() (string, bool) {
		return t.Text(), true
	})
	t.v.SetAutomationValueHandler(func() (string, bool) {
		return t.Text(), true
	})
	return t
}

func (t *UITextView) View() *view.UIView {
	if t == nil {
		return nil
	}
	return &t.v
}

func (t *UITextView) Raw() *fltk_bridge.TextEditor {
	if t == nil {
		return nil
	}
	return t.raw
}

func (t *UITextView) TextBuffer() *fltk_bridge.TextBuffer {
	if t == nil {
		return nil
	}
	return t.buffer
}

func (t *UITextView) SetText(text string) {
	if t != nil && t.buffer != nil {
		t.buffer.SetText(text)
	}
}

func (t *UITextView) Text() string {
	if t == nil || t.buffer == nil {
		return ""
	}
	return t.buffer.Text()
}

func (t *UITextView) Append(text string) {
	if t != nil && t.buffer != nil {
		t.buffer.Append(text)
	}
}

func (t *UITextView) AppendText(text string) {
	t.Append(text)
}

func (t *UITextView) AppendAndScroll(text string) {
	t.Append(text)
	t.ScrollToEnd()
}

func (t *UITextView) SetWrapAtBounds() {
	if t != nil && t.raw != nil {
		t.raw.SetWrapMode(fltk_bridge.WRAP_AT_BOUNDS)
	}
}

func (t *UITextView) SetWrapNone() {
	if t != nil && t.raw != nil {
		t.raw.SetWrapMode(fltk_bridge.WRAP_NONE)
	}
}

func (t *UITextView) SetFont(font fltk_bridge.Font) {
	if t != nil && t.raw != nil {
		t.font = font
		t.raw.SetTextFont(font)
	}
}

func (t *UITextView) SetFontSize(size int) {
	if t != nil && t.raw != nil {
		t.fontSize = size
		t.raw.SetTextSize(size)
	}
}

func (t *UITextView) SetTextColor(rgb uint) {
	if t != nil && t.raw != nil {
		t.textColor = fltk_bridge.Color(rgb)
		t.raw.SetTextColor(fltk_bridge.Color(rgb))
	}
}

// SetCursorColor styles the native insertion caret independently from text so
// editable views remain visible on custom light and dark surfaces.
func (t *UITextView) SetCursorColor(rgb uint) {
	if t != nil && t.raw != nil {
		t.raw.SetCursorColor(fltk_bridge.Color(rgb))
		t.raw.Redraw()
	}
}

// SetSelectionColor styles the native selection highlight.
func (t *UITextView) SetSelectionColor(rgb uint) {
	if t != nil && t.raw != nil {
		t.raw.SetSelectionColor(fltk_bridge.Color(rgb))
		t.raw.Redraw()
	}
}

// SetFallbackFont styles matching Unicode runes with a second FLTK font while
// preserving the primary font for all other text. It is intended for native
// font stacks where one platform font cannot cover every required script (for
// example, a CJK font plus an emoji font). Configure the primary font, size and
// color before calling this method.
func (t *UITextView) SetFallbackFont(font fltk_bridge.Font, match func(rune) bool) {
	if t == nil || t.raw == nil || t.buffer == nil || match == nil {
		return
	}
	if t.styleBuffer == nil {
		t.styleBuffer = fltk_bridge.NewTextBuffer()
	}
	t.fallbackFont = font
	t.fallbackMatch = match
	t.styleBuffer.SetText(buildFallbackStyles(t.Text(), match))
	if !t.fallbackWatch {
		t.fallbackWatch = true
		t.buffer.AddModifyCallback(func(pos, inserted, deleted, _ int, _ string) {
			if deleted > 0 {
				t.styleBuffer.Remove(pos, pos+deleted)
			}
			if inserted > 0 {
				text := t.buffer.GetTextRange(pos, pos+inserted)
				t.styleBuffer.Insert(pos, buildFallbackStyles(text, t.fallbackMatch))
			}
		})
	}
	t.raw.SetHighlightData(t.styleBuffer, []fltk_bridge.StyleTableEntry{
		{Color: t.textColor, Font: t.font, Size: t.fontSize},
		{Color: t.textColor, Font: t.fallbackFont, Size: t.fontSize},
	})
}

func buildFallbackStyles(text string, match func(rune) bool) string {
	if match == nil {
		return strings.Repeat("A", len(text))
	}
	var styles strings.Builder
	styles.Grow(len(text))
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		style := byte('A')
		if match(r) {
			style = 'B'
		}
		for range size {
			styles.WriteByte(style)
		}
		text = text[size:]
	}
	return styles.String()
}

func (t *UITextView) SetBackgroundColor(rgb uint) {
	if t != nil && t.raw != nil {
		t.raw.SetColor(fltk_bridge.Color(rgb))
		t.raw.Redraw()
	}
}

func (t *UITextView) SetTabWidth(width int) {
	if t != nil && t.buffer != nil {
		t.buffer.SetTabWidth(width)
	}
}

func (t *UITextView) ScrollToEnd() {
	if t != nil && t.raw != nil && t.buffer != nil {
		t.raw.SetInsertPosition(t.buffer.Length())
		t.raw.ShowInsertPosition()
	}
}

func (t *UITextView) SetAutomationID(id string) *UITextView {
	if t != nil {
		t.v.SetAutomationID(id)
	}
	return t
}

func (t *UITextView) SetAutomationName(name string) *UITextView {
	if t != nil {
		t.v.SetAutomationName(name)
	}
	return t
}

func (t *UITextView) OnTextChanged(cb func()) {
	if t == nil || t.buffer == nil {
		return
	}
	t.buffer.AddModifyCallback(func(int, int, int, int, string) {
		if cb != nil {
			cb()
		}
	})
}

func (t *UITextView) On(event fltk_bridge.Event, handler func(fltk_bridge.Event) bool) {
	if t != nil {
		t.v.On(event, handler)
	}
}

func (t *UITextView) OnKey(cb func(KeyEvent) bool) {
	if t == nil || cb == nil {
		return
	}
	t.On(fltk_bridge.KEYDOWN, func(e fltk_bridge.Event) bool {
		return cb(KeyEvent{
			Key:   fltk_bridge.EventKey(),
			Text:  fltk_bridge.EventText(),
			State: fltk_bridge.EventState(),
		})
	})
}
