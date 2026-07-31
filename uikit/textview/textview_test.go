package textview

import (
	"strings"
	"testing"
)

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
