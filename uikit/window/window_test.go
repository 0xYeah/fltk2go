package window

import (
	"testing"

	"github.com/0xdevelop/fltk2go/foundation"
)

func TestUIWindowCloseLifecycleIsIdempotent(t *testing.T) {
	win := NewWindowWithRect(&foundation.Rect{X: 0, Y: 0, Width: 320, Height: 180}, "lifecycle")
	closed := 0
	win.OnClose(func() { closed++ })

	if win.IsClosed() {
		t.Fatal("new window must be open")
	}
	win.Close()
	win.Close()

	if !win.IsClosed() {
		t.Fatal("closed window must report closed")
	}
	if closed != 1 {
		t.Fatalf("close handler called %d times, want 1", closed)
	}
	if win.Raw() != nil {
		t.Fatal("closed window must release its native handle")
	}
}
