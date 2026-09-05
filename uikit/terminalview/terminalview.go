package terminalview

import (
	"strings"
	"unicode"

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

// ContextMenuState describes terminal state at the exact right-click that
// requests an application-owned context menu. Applications can use it to
// disable Copy without reaching through to the native FLTK widget.
type ContextMenuState struct {
	HasSelection bool
}

// TextMatch identifies a literal case-insensitive match in the terminal's
// rendered text. Line, Column, and Length are measured in Unicode code points,
// so callers can present stable positions without exposing byte offsets.
type TextMatch struct {
	Line   int
	Column int
	Length int
}

// SearchText returns every non-overlapping literal match in display order.
// Matching is Unicode-aware and case-insensitive; a whitespace-only query is
// treated as empty. Keeping this pure makes search state and navigation an
// application concern while the reusable terminal owns text/grid semantics.
func SearchText(text, query string) []TextMatch {
	if strings.TrimSpace(query) == "" {
		return nil
	}
	needle := foldRunes([]rune(query))
	if len(needle) == 0 {
		return nil
	}
	var matches []TextMatch
	for lineIndex, line := range strings.Split(text, "\n") {
		haystack := foldRunes([]rune(strings.TrimSuffix(line, "\r")))
		for column := 0; column+len(needle) <= len(haystack); {
			if equalRunes(haystack[column:column+len(needle)], needle) {
				matches = append(matches, TextMatch{Line: lineIndex, Column: column, Length: len(needle)})
				column += len(needle)
				continue
			}
			column++
		}
	}
	return matches
}

func foldRunes(value []rune) []rune {
	folded := make([]rune, len(value))
	for i, r := range value {
		folded[i] = unicode.ToLower(r)
	}
	return folded
}

func equalRunes(left, right []rune) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

type terminalClipboardAction uint8

const (
	terminalClipboardNone terminalClipboardAction = iota
	terminalClipboardCopy
	terminalClipboardPaste
)

type terminalScrollAction uint8

const (
	terminalScrollNone terminalScrollAction = iota
	terminalScrollPageUp
	terminalScrollPageDown
	terminalScrollTop
	terminalScrollBottom
)

func clipboardActionForKey(event KeyEvent) terminalClipboardAction {
	ctrl := event.State&fltk_bridge.CTRL != 0
	shift := event.State&fltk_bridge.SHIFT != 0
	if ctrl && shift && (event.Key == int('c') || event.Key == int('C')) {
		return terminalClipboardCopy
	}
	if (ctrl && shift && (event.Key == int('v') || event.Key == int('V'))) || (shift && !ctrl && event.Key == fltk_bridge.INSERT) {
		return terminalClipboardPaste
	}
	return terminalClipboardNone
}

func scrollActionForKey(event KeyEvent) terminalScrollAction {
	if event.State&fltk_bridge.SHIFT == 0 || event.State&fltk_bridge.CTRL != 0 || event.State&fltk_bridge.ALT != 0 {
		return terminalScrollNone
	}
	switch event.Key {
	case fltk_bridge.PAGE_UP:
		return terminalScrollPageUp
	case fltk_bridge.PAGE_DOWN:
		return terminalScrollPageDown
	case fltk_bridge.HOME:
		return terminalScrollTop
	case fltk_bridge.END:
		return terminalScrollBottom
	default:
		return terminalScrollNone
	}
}

// UITerminalView is a native ANSI/VT presentation and input surface. It owns
// terminal parsing, scrollback, selection, focus, and key-sequence translation;
// it intentionally does not own a process or network connection.
type UITerminalView struct {
	v             view.UIView
	raw           *fltk_bridge.Terminal
	onInput       func([]byte)
	onResize      func(Size)
	lastSize      Size
	inputBound    bool
	shortcuts     []terminalShortcut
	onContextMenu func(ContextMenuState)
	textObservers map[uint64]func()
	nextObserver  uint64
	stream        terminalStreamFilter
}

type terminalShortcut struct {
	shortcut int
	handler  func()
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
	raw.SetHorizontalScrollbar(fltk_bridge.TerminalScrollbarOff)
	t := &UITerminalView{raw: raw}
	t.v.BindRaw(raw)
	t.v.SetAutomationRole("terminal").SetAutomationName("Terminal")
	t.v.SetAutomationValueHandler(func() (string, bool) { return t.Text(), true })
	t.v.On(fltk_bridge.PUSH, func(fltk_bridge.Event) bool {
		raw.TakeFocus()
		if t.dispatchContextMenu(fltk_bridge.EventButton()) {
			return true
		}
		return false // preserve Fl_Terminal mouse selection handling
	})
	raw.SetResizeHandler(func() {
		raw.FitDisplayColumns()
		t.publishSize()
	})
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
		filtered := t.stream.Filter(data)
		if len(filtered) == 0 {
			return
		}
		t.raw.AppendBytes(filtered)
		t.notifyTextChanged()
	}
}

func (t *UITerminalView) Append(text string) { t.Feed([]byte(text)) }

func (t *UITerminalView) Text() string {
	if t == nil || t.raw == nil {
		return ""
	}
	return t.raw.Text()
}

// SearchText returns literal case-insensitive matches in the terminal's current
// retained output, including scrollback.
func (t *UITerminalView) SearchText(query string) []TextMatch {
	if t == nil {
		return nil
	}
	return SearchText(t.Text(), query)
}

// ObserveTextChanged registers a lightweight notification for rendered text
// mutations. It is useful for search/count surfaces that should follow live
// terminal output without polling. The returned function is idempotent and
// removes only this observer.
func (t *UITerminalView) ObserveTextChanged(handler func()) func() {
	if t == nil || handler == nil {
		return func() {}
	}
	if t.textObservers == nil {
		t.textObservers = make(map[uint64]func())
	}
	t.nextObserver++
	id := t.nextObserver
	t.textObservers[id] = handler
	return func() { delete(t.textObservers, id) }
}

func (t *UITerminalView) notifyTextChanged() {
	if t == nil || len(t.textObservers) == 0 {
		return
	}
	observers := make([]func(), 0, len(t.textObservers))
	for _, observer := range t.textObservers {
		observers = append(observers, observer)
	}
	for _, observer := range observers {
		observer()
	}
}

func (t *UITerminalView) Clear() {
	if t != nil && t.raw != nil {
		t.raw.Clear()
		t.notifyTextChanged()
	}
}

func (t *UITerminalView) ClearHistory() {
	if t != nil && t.raw != nil {
		t.raw.ClearHistory()
		t.notifyTextChanged()
	}
}

// HasSelection reports whether the native terminal currently has selected
// text. It does not expose the complete scrollback buffer.
func (t *UITerminalView) HasSelection() bool {
	return t != nil && t.raw != nil && t.raw.SelectedText() != ""
}

// CopySelection copies only the current native terminal selection. It returns
// false when there is no selection, allowing menu actions to remain no-ops.
func (t *UITerminalView) CopySelection() bool {
	if t == nil || t.raw == nil {
		return false
	}
	selected := t.raw.SelectedText()
	if selected == "" {
		return false
	}
	fltk_bridge.CopyToClipboard(selected)
	return true
}

// CopyAllText copies the terminal's complete rendered text, including
// scrollback. It returns false when the terminal has no text to copy.
func (t *UITerminalView) CopyAllText() bool {
	if t == nil || t.raw == nil {
		return false
	}
	text := t.raw.Text()
	if strings.TrimSpace(text) == "" {
		return false
	}
	fltk_bridge.CopyToClipboard(text)
	return true
}

// PasteClipboard requests the system clipboard and delivers it through the
// same OnInput path used by the native Ctrl+Shift+V and Shift+Insert actions.
func (t *UITerminalView) PasteClipboard() {
	if t != nil && t.raw != nil {
		t.raw.PasteClipboard()
	}
}

func (t *UITerminalView) Reset() {
	if t != nil && t.raw != nil {
		t.stream.Reset()
		t.raw.Reset()
		t.notifyTextChanged()
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
		t.notifyTextChanged()
	}
}

// ScrollOffset reports history rows above the live bottom. ScrollToOffset and
// ScrollByRows provide transport-neutral scrollback navigation for application
// controls in addition to the built-in Shift+PageUp/PageDown/Home/End keys.
func (t *UITerminalView) ScrollOffset() int {
	if t == nil || t.raw == nil {
		return 0
	}
	return t.raw.ScrollOffset()
}

func (t *UITerminalView) ScrollToOffset(rowsFromBottom int) {
	if t != nil && t.raw != nil {
		t.raw.ScrollTo(rowsFromBottom)
	}
}

func (t *UITerminalView) ScrollByRows(rows int) {
	if t != nil && t.raw != nil {
		t.raw.ScrollTo(t.raw.ScrollOffset() + rows)
	}
}

func (t *UITerminalView) ScrollToTop() {
	if t != nil && t.raw != nil {
		t.raw.ScrollTo(t.raw.ScrollMaximum())
	}
}

func (t *UITerminalView) ScrollToBottom() { t.ScrollToOffset(0) }

// RevealTextMatch scrolls a retained output match near the vertical center of
// the native viewport. It deliberately does not mutate the user's mouse
// selection; applications can cycle matches without destroying copied text.
func (t *UITerminalView) RevealTextMatch(match TextMatch) {
	if t == nil || t.raw == nil || match.Line < 0 {
		return
	}
	totalLines := len(strings.Split(t.Text(), "\n"))
	if match.Line >= totalLines {
		return
	}
	rowsBelow := totalLines - 1 - match.Line
	target := rowsBelow - t.Size().Rows/2
	if target < 0 {
		target = 0
	}
	t.ScrollToOffset(target)
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

// OnShortcut registers an application command that takes priority over PTY/SSH
// input while this terminal owns focus. This is the terminal counterpart to a
// window shortcut: consuming controls can handle global search/settings commands
// without leaking their control bytes to the attached session.
func (t *UITerminalView) OnShortcut(shortcut int, handler func()) {
	if t == nil || shortcut == 0 || handler == nil {
		return
	}
	for i := range t.shortcuts {
		if t.shortcuts[i].shortcut == shortcut {
			t.shortcuts[i].handler = handler
			return
		}
	}
	t.shortcuts = append(t.shortcuts, terminalShortcut{shortcut: shortcut, handler: handler})
}

// OnContextMenu registers an application-owned menu request. Right-click is
// consumed only when a handler exists; ordinary mouse selection remains native.
func (t *UITerminalView) OnContextMenu(handler func(ContextMenuState)) {
	if t != nil {
		t.onContextMenu = handler
	}
}

func (t *UITerminalView) dispatchContextMenu(button fltk_bridge.MouseButton) bool {
	if t == nil || t.onContextMenu == nil || button != fltk_bridge.RightMouse {
		return false
	}
	t.onContextMenu(ContextMenuState{HasSelection: t.HasSelection()})
	return true
}

func (t *UITerminalView) dispatchShortcut(matches func(int) bool) bool {
	if t == nil || matches == nil {
		return false
	}
	for _, shortcut := range t.shortcuts {
		if matches(shortcut.shortcut) {
			shortcut.handler()
			return true
		}
	}
	return false
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
		event := KeyEvent{Key: fltk_bridge.EventKey(), Text: fltk_bridge.EventText(), State: fltk_bridge.EventState()}
		if t.dispatchShortcut(fltk_bridge.TestShortcut) {
			return true
		}
		switch clipboardActionForKey(event) {
		case terminalClipboardCopy:
			t.CopySelection()
			return true
		case terminalClipboardPaste:
			t.PasteClipboard()
			return true
		}
		switch scrollActionForKey(event) {
		case terminalScrollPageUp:
			t.ScrollByRows(max(1, t.Size().Rows-1))
			return true
		case terminalScrollPageDown:
			t.ScrollByRows(-max(1, t.Size().Rows-1))
			return true
		case terminalScrollTop:
			t.ScrollToTop()
			return true
		case terminalScrollBottom:
			t.ScrollToBottom()
			return true
		}
		data, handled := EncodeKey(event)
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
// Ctrl+C/D/Z remain distinct ETX/EOT/SUB bytes. Clipboard shortcuts are routed
// by UITerminalView and must never leak control bytes into the PTY.
func EncodeKey(event KeyEvent) ([]byte, bool) {
	if clipboardActionForKey(event) != terminalClipboardNone || scrollActionForKey(event) != terminalScrollNone {
		return nil, false
	}
	ctrl := event.State&fltk_bridge.CTRL != 0
	shift := event.State&fltk_bridge.SHIFT != 0
	alt := event.State&fltk_bridge.ALT != 0

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
