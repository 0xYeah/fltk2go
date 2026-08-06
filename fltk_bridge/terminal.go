package fltk_bridge

/*
#include <stdlib.h>
#include "terminal.h"
*/
import "C"

import (
	"unsafe"
)

// Terminal is the native FLTK ANSI/VT output surface. It parses arbitrary
// incremental UTF-8 and escape-sequence chunks, but deliberately does not own a
// PTY or SSH connection; transport lifecycle remains an application concern.
type Terminal struct {
	Group
}

func NewTerminal(x, y, w, h int, label ...string) *Terminal {
	t := &Terminal{}
	initWidget(t, unsafe.Pointer(C.go_fltk_new_Terminal(C.int(x), C.int(y), C.int(w), C.int(h), cStringOpt(label))))
	return t
}

func (t *Terminal) ptrTerminal() *C.Fl_Terminal {
	return (*C.Fl_Terminal)(unsafe.Pointer(t.ptr()))
}

// AppendBytes feeds a raw terminal byte-stream chunk. Embedded NUL bytes are
// preserved at the bridge boundary and split UTF-8/ANSI sequences remain the
// native terminal parser's responsibility.
func (t *Terminal) AppendBytes(data []byte) {
	if len(data) == 0 {
		return
	}
	ptr := C.CBytes(data)
	defer C.free(ptr)
	C.go_fltk_Terminal_append(t.ptrTerminal(), (*C.char)(ptr), C.int(len(data)))
}

func (t *Terminal) Append(text string) { t.AppendBytes([]byte(text)) }

func (t *Terminal) Text(linesBelowCursor ...bool) string {
	includeBelow := false
	if len(linesBelowCursor) > 0 {
		includeBelow = linesBelowCursor[0]
	}
	value := C.go_fltk_Terminal_text(t.ptrTerminal(), boolToInt(includeBelow))
	if value == nil {
		return ""
	}
	defer C.free(unsafe.Pointer(value))
	return C.GoString(value)
}

func (t *Terminal) Clear()        { C.go_fltk_Terminal_clear(t.ptrTerminal()) }
func (t *Terminal) ClearHistory() { C.go_fltk_Terminal_clear_history(t.ptrTerminal()) }
func (t *Terminal) Reset()        { C.go_fltk_Terminal_reset(t.ptrTerminal()) }
func (t *Terminal) SetANSI(enabled bool) {
	C.go_fltk_Terminal_set_ansi(t.ptrTerminal(), boolToInt(enabled))
}
func (t *Terminal) ANSI() bool { return C.go_fltk_Terminal_ansi(t.ptrTerminal()) != 0 }
func (t *Terminal) SetHistoryRows(rows int) {
	C.go_fltk_Terminal_set_history_rows(t.ptrTerminal(), C.int(rows))
}
func (t *Terminal) HistoryRows() int {
	return int(C.go_fltk_Terminal_history_rows(t.ptrTerminal()))
}
func (t *Terminal) DisplayRows() int {
	return int(C.go_fltk_Terminal_display_rows(t.ptrTerminal()))
}
func (t *Terminal) DisplayColumns() int {
	return int(C.go_fltk_Terminal_display_columns(t.ptrTerminal()))
}
func (t *Terminal) FitDisplayColumns() int {
	return int(C.go_fltk_Terminal_fit_display_columns((*C.GTerminal)(unsafe.Pointer(t.ptr()))))
}

// TerminalScrollbarStyle controls the native terminal's horizontal scrollbar.
type TerminalScrollbarStyle int

const (
	TerminalScrollbarOff TerminalScrollbarStyle = iota
	TerminalScrollbarAuto
	TerminalScrollbarOn
)

func (t *Terminal) SetHorizontalScrollbar(style TerminalScrollbarStyle) {
	C.go_fltk_Terminal_set_horizontal_scrollbar(t.ptrTerminal(), C.int(style))
}
func (t *Terminal) SetTextFont(font Font) {
	C.go_fltk_Terminal_set_text_font(t.ptrTerminal(), C.int(font))
}
func (t *Terminal) SetTextSize(size int) {
	C.go_fltk_Terminal_set_text_size(t.ptrTerminal(), C.int(size))
}
func (t *Terminal) SetTextColor(color Color) {
	C.go_fltk_Terminal_set_text_color(t.ptrTerminal(), C.uint(color))
}
func (t *Terminal) SetBackgroundColor(color Color) {
	C.go_fltk_Terminal_set_background_color(t.ptrTerminal(), C.uint(color))
}
func (t *Terminal) SetSelectionColors(foreground, background Color) {
	C.go_fltk_Terminal_set_selection_colors(t.ptrTerminal(), C.uint(foreground), C.uint(background))
}
func (t *Terminal) SetMargins(left, top, right, bottom int) {
	C.go_fltk_Terminal_set_margins(t.ptrTerminal(), C.int(left), C.int(top), C.int(right), C.int(bottom))
}
func (t *Terminal) SetRedrawRate(seconds float32) {
	C.go_fltk_Terminal_set_redraw_rate(t.ptrTerminal(), C.float(seconds))
}

func boolToInt(value bool) C.int {
	if value {
		return 1
	}
	return 0
}
