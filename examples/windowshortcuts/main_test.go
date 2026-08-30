package main

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func TestSearchShortcutUsesNativeControlK(t *testing.T) {
	if got := fltk_bridge.CTRL + int('k'); got == 0 {
		t.Fatal("native Ctrl+K shortcut must be non-zero")
	}
}
