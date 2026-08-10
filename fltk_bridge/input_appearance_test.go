package fltk_bridge

import "testing"

func TestInputTextAndCaretAppearanceRoundTrip(t *testing.T) {
	in := NewInput(0, 0, 180, 34)
	in.SetTextFont(COURIER)
	in.SetTextSize(17)
	in.SetTextColor(ColorFromRgb(12, 34, 56))
	in.SetCursorColor(ColorFromRgb(78, 90, 123))

	if in.TextFont() != COURIER || in.TextSize() != 17 {
		t.Fatalf("input typography = font:%v size:%d", in.TextFont(), in.TextSize())
	}
	if got, want := in.TextColor(), ColorFromRgb(12, 34, 56); got != want {
		t.Fatalf("input text color = %v, want %v", got, want)
	}
	if got, want := in.CursorColor(), ColorFromRgb(78, 90, 123); got != want {
		t.Fatalf("input cursor color = %v, want %v", got, want)
	}
}
