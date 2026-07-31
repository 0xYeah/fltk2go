package dropdown

import (
	"testing"

	"github.com/0xdevelop/fltk2go/foundation"
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
