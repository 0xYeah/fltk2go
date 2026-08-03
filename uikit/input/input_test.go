package input

import (
	"testing"

	"github.com/0xdevelop/fltk2go/uikit/view"
)

func TestSecretInputAutomationSnapshotDoesNotExposeText(t *testing.T) {
	in := NewWithType(0, 0, 120, 24, "Password", SecretInput)
	in.SetText("sensitive-value")
	in.View().SetAutomationID("test.secret")
	defer in.View().SetAutomationID("")

	node := in.View().AutomationSnapshot()
	if node.Text != "" || node.Value != "" {
		t.Fatalf("secret automation snapshot exposed text: %#v", node)
	}
	if node.Properties["secure"] != "true" {
		t.Fatalf("secure property = %q, want true", node.Properties["secure"])
	}
	if got := in.Text(); got != "sensitive-value" {
		t.Fatalf("widget text = %q, want owner-visible value", got)
	}
}

func TestTextInputAutomationSnapshotRetainsText(t *testing.T) {
	in := New(0, 0, 120, 24, "Name")
	in.SetText("visible-value")

	if got := in.View().AutomationSnapshot().Text; got != "visible-value" {
		t.Fatalf("snapshot text = %q, want visible-value", got)
	}
}

func TestAutomationSetTextPublishesNativeChangeCallback(t *testing.T) {
	in := New(0, 0, 120, 24, "Search")
	in.View().SetAutomationID("test.search")
	defer in.View().SetAutomationID("")
	changes := 0
	in.OnChange(func() { changes++ })

	if err := view.AutomationSetText("test.search", "日本語 SSH"); err != nil {
		t.Fatalf("AutomationSetText failed: %v", err)
	}
	if in.Text() != "日本語 SSH" || changes != 1 {
		t.Fatalf("text/change count = %q/%d, want 日本語 SSH/1", in.Text(), changes)
	}
}
