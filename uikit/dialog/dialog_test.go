package dialog

import "testing"

func TestChoiceRejectsUnsupportedOptionCountsWithoutPanicking(t *testing.T) {
	if got := Choice("message"); got != -1 {
		t.Fatalf("Choice with no options = %d, want -1", got)
	}
	if got := Choice("message", "A", "B", "C"); got != -1 {
		t.Fatalf("Choice with three options = %d, want -1", got)
	}
}

func TestTitledChoiceRejectsUnsupportedOptionCountsWithoutOpeningDialog(t *testing.T) {
	if got := TitledChoice("Security warning", "message"); got != -1 {
		t.Fatalf("TitledChoice with no options = %d, want -1", got)
	}
	if got := TitledChoice("Security warning", "message", "A", "B", "C"); got != -1 {
		t.Fatalf("TitledChoice with three options = %d, want -1", got)
	}
}
