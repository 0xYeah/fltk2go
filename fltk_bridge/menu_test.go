package fltk_bridge

import "testing"

func TestMenuClearReleasesItemCallbacks(t *testing.T) {
	menu := NewMenuButton(0, 0, 0, 0)
	t.Cleanup(menu.Destroy)
	menu.Add("Copy", func() {})
	menu.Add("Paste", func() {})
	if got := len(menu.itemCallbacks); got != 2 {
		t.Fatalf("registered callbacks = %d, want 2", got)
	}
	menu.Clear()
	if got := len(menu.itemCallbacks); got != 0 {
		t.Fatalf("callbacks after Clear = %d, want 0", got)
	}
	if got := menu.Size(); got != 0 {
		t.Fatalf("menu size after Clear = %d, want 0", got)
	}
}
