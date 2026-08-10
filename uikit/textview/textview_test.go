package textview

import (
	"strings"
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func TestTextViewAppearanceTargetsTextCaretAndSelection(t *testing.T) {
	view := NewUITextView(nil)
	text := uint(fltk_bridge.ColorFromRgb(220, 230, 240))
	caret := uint(fltk_bridge.ColorFromRgb(69, 178, 157))
	selection := uint(fltk_bridge.ColorFromRgb(45, 67, 89))
	view.SetTextColor(text)
	view.SetCursorColor(caret)
	view.SetSelectionColor(selection)

	if view.Raw().TextColor() != fltk_bridge.Color(text) || view.Raw().CursorColor() != fltk_bridge.Color(caret) {
		t.Fatalf("editable colors = text:%v cursor:%v", view.Raw().TextColor(), view.Raw().CursorColor())
	}
	if view.Raw().SelectionColor() != fltk_bridge.Color(selection) {
		t.Fatalf("selection color = %v, want %v", view.Raw().SelectionColor(), fltk_bridge.Color(selection))
	}
}

func TestBuildFallbackStylesMatchesUTF8BytePositions(t *testing.T) {
	text := "A中🚀e\u0301"
	styles := buildFallbackStyles(text, func(r rune) bool { return r == '🚀' })
	if len(styles) != len(text) {
		t.Fatalf("style bytes = %d, text bytes = %d", len(styles), len(text))
	}
	if got := strings.Count(styles, "B"); got != len("🚀") {
		t.Fatalf("fallback style byte count = %d, want %d: %q", got, len("🚀"), styles)
	}
	if strings.Count(styles, "A") != len(text)-len("🚀") {
		t.Fatalf("base style byte count is wrong: %q", styles)
	}
}

func TestBuildFallbackStylesWithoutMatcherUsesBaseStyle(t *testing.T) {
	const text = "中文 · emoji 🚀"
	if got := buildFallbackStyles(text, nil); got != strings.Repeat("A", len(text)) {
		t.Fatalf("styles without matcher = %q", got)
	}
}
