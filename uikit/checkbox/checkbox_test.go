package checkbox

import (
	"testing"

	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

func TestCheckboxExposesSemanticAutomationToggle(t *testing.T) {
	checkbox := NewUICheckbox(&foundation.Rect{Width: 160, Height: 30}, "Remember")
	checkbox.View().SetAutomationID("test.checkbox")
	defer checkbox.View().SetAutomationID("")

	called := 0
	checkbox.OnValueChanged(func(value bool) { called++ })
	if err := view.AutomationClick("test.checkbox"); err != nil {
		t.Fatalf("semantic toggle failed: %v", err)
	}
	if !checkbox.Value() || called != 1 {
		t.Fatalf("semantic toggle state = value:%v callbacks:%d", checkbox.Value(), called)
	}
	node := checkbox.View().AutomationSnapshot()
	if node.Role != "checkbox" || node.Value != "true" {
		t.Fatalf("automation snapshot = %#v", node)
	}
}
