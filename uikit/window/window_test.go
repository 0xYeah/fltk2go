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

func TestUIWindowCloseRequestCanBeVetoedBeforeLifecycleCleanup(t *testing.T) {
	win := NewWindowWithRect(&foundation.Rect{X: 0, Y: 0, Width: 320, Height: 180}, "close request")
	requests := 0
	closed := 0
	allow := false
	win.OnCloseRequest(func() bool {
		requests++
		return allow
	})
	win.OnClose(func() { closed++ })

	if win.RequestClose() {
		t.Fatal("vetoed close request must report that the window remains open")
	}
	if win.IsClosed() || win.Raw() == nil || requests != 1 || closed != 0 {
		t.Fatalf("vetoed request mutated lifecycle: closed=%v raw=%v requests=%d callbacks=%d", win.IsClosed(), win.Raw(), requests, closed)
	}

	allow = true
	if !win.RequestClose() {
		t.Fatal("accepted close request must report completion")
	}
	if !win.IsClosed() || requests != 2 || closed != 1 {
		t.Fatalf("accepted request did not close exactly once: closed=%v requests=%d callbacks=%d", win.IsClosed(), requests, closed)
	}
	if !win.RequestClose() || requests != 2 || closed != 1 {
		t.Fatal("requesting close after completion must remain idempotent")
	}
}

func TestUIWindowCloseBypassesRequestPolicyForOwnerTeardown(t *testing.T) {
	win := NewWindowWithRect(&foundation.Rect{X: 0, Y: 0, Width: 320, Height: 180}, "forced close")
	requests := 0
	win.OnCloseRequest(func() bool {
		requests++
		return false
	})

	win.Close()
	if !win.IsClosed() || requests != 0 {
		t.Fatalf("owner Close must bypass user-request policy: closed=%v requests=%d", win.IsClosed(), requests)
	}
}
