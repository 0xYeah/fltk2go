package fltk_bridge

import (
	"errors"
	"testing"
)

func TestTextEditorCursorColorRoundTrips(t *testing.T) {
	editor := NewTextEditor(0, 0, 240, 80)
	want := ColorFromRgb(78, 90, 123)
	editor.SetCursorColor(want)
	if got := editor.CursorColor(); got != want {
		t.Fatalf("cursor color = %v, want %v", got, want)
	}
}

func TestPanicWhenTestBufferIsMissing(t *testing.T) {
	win := NewWindow(400, 400)
	textEditor := NewTextEditor(2, 2, 300, 300, "")
	win.End()
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Did not panic")
		} else if err, ok := r.(error); !ok {
			t.Errorf("Panicked with not an error: %v", r)
		} else if !errors.Is(err, ErrNoTextBufferAssociated) {
			t.Errorf("Unexpected error: %v", err)
		}
		textEditor.Destroy()
		Unlock()
		// clean up after outselves as other tests check if this map is empty
		globalCallbackMap.clear()
	}()
	textEditor.SetEventHandler(func(event Event) bool {
		if event != SHOW {
			return false
		}
		textEditor.SelectAll()
		panic("Should have panicked")
	})
	Lock()
	win.Show()
	Run()
}
