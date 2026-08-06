package fltk_bridge

import (
	"strings"
	"testing"
)

func TestTerminalParsesIncrementalANSIAndUTF8(t *testing.T) {
	terminal := NewTerminal(0, 0, 640, 320)
	defer terminal.Destroy()
	terminal.SetANSI(true)
	terminal.SetHistoryRows(128)
	terminal.AppendBytes([]byte("plain "))
	terminal.AppendBytes([]byte{0xe4, 0xb8})
	terminal.AppendBytes([]byte{0xad})
	terminal.AppendBytes([]byte(" · Русский\n\x1b[31mred\x1b[0m"))

	got := terminal.Text(true)
	for _, want := range []string{"plain 中 · Русский", "red"} {
		if !strings.Contains(got, want) {
			t.Fatalf("terminal text %q does not contain %q", got, want)
		}
	}
	if strings.Contains(got, "\x1b[31m") {
		t.Fatalf("terminal text leaked parsed ANSI bytes: %q", got)
	}
	if !terminal.ANSI() || terminal.HistoryRows() != 128 {
		t.Fatalf("terminal settings not retained: ansi=%t history=%d", terminal.ANSI(), terminal.HistoryRows())
	}
}

func TestTerminalResetAndClear(t *testing.T) {
	terminal := NewTerminal(0, 0, 320, 160)
	defer terminal.Destroy()
	terminal.Append("before")
	terminal.Clear()
	if got := strings.TrimSpace(terminal.Text(true)); got != "" {
		t.Fatalf("terminal text after clear = %q", got)
	}
	terminal.Append("after")
	terminal.Reset()
	if terminal.DisplayColumns() <= 0 || terminal.DisplayRows() <= 0 {
		t.Fatalf("invalid terminal grid after reset: %dx%d", terminal.DisplayColumns(), terminal.DisplayRows())
	}
}

func TestTerminalFitsColumnsAfterNativeResize(t *testing.T) {
	terminal := NewTerminal(0, 0, 900, 500)
	defer terminal.Destroy()
	before := terminal.DisplayColumns()
	terminal.Resize(0, 0, 400, 300)
	after := terminal.FitDisplayColumns()
	if before <= 0 || after <= 0 || after >= before {
		t.Fatalf("terminal columns did not fit resized width: before=%d after=%d", before, after)
	}
	if got := terminal.DisplayColumns(); got != after {
		t.Fatalf("terminal display columns = %d, want fitted %d", got, after)
	}
}
