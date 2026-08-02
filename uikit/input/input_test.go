package input

import "testing"

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
