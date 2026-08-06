package dropdown

import (
	"testing"

	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

func TestMenuItemSelectionNotifiesDropdownHandler(t *testing.T) {
	dd := NewUIDropdown(&foundation.Rect{Width: 120, Height: 30})
	dd.SetOptions([]string{"English", "简体中文"})

	called := 0
	gotIndex := -1
	gotOption := ""
	dd.OnSelectionChanged(func(index int, option string) {
		called++
		gotIndex = index
		gotOption = option
	})

	dd.selectOption(1)
	if called != 1 || gotIndex != 1 || gotOption != "简体中文" {
		t.Fatalf("selection callback = called:%d index:%d option:%q", called, gotIndex, gotOption)
	}
	if dd.SelectedIndex() != 1 || dd.SelectedOption() != "简体中文" {
		t.Fatalf("selected state = index:%d option:%q", dd.SelectedIndex(), dd.SelectedOption())
	}
}

func TestDropdownExposesSemanticAutomationSelection(t *testing.T) {
	dd := NewUIDropdown(&foundation.Rect{Width: 120, Height: 30})
	dd.SetOptions([]string{"English", "简体中文"})
	dd.View().SetAutomationID("test.dropdown")
	defer dd.View().SetAutomationID("")

	called := 0
	redrawn := 0
	dd.redrawSelection = func() { redrawn++ }
	dd.OnSelectionChanged(func(index int, option string) { called++ })
	if err := view.AutomationSetText("test.dropdown", "简体中文"); err != nil {
		t.Fatalf("semantic selection failed: %v", err)
	}
	if dd.SelectedIndex() != 1 || dd.SelectedOption() != "简体中文" || called != 1 {
		t.Fatalf("semantic selection state = index:%d option:%q callbacks:%d", dd.SelectedIndex(), dd.SelectedOption(), called)
	}
	if redrawn != 1 {
		t.Fatalf("semantic selection redraws = %d, want 1", redrawn)
	}
	node := dd.View().AutomationSnapshot()
	if node.Role != "combobox" || node.Value != "简体中文" {
		t.Fatalf("automation snapshot = %#v", node)
	}
}
