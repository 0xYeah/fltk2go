package input

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

func TestInputAppearanceTargetsEditableTextAndCaret(t *testing.T) {
	in := New(0, 0, 180, 34, "Name")
	in.SetFont(fltk_bridge.COURIER)
	in.SetFontSize(16)
	in.SetTextColor(uint(fltk_bridge.ColorFromRgb(20, 30, 40)))
	in.SetCursorColor(uint(fltk_bridge.ColorFromRgb(220, 230, 240)))

	raw, ok := in.Raw().(*fltk_bridge.Input)
	if !ok {
		t.Fatalf("raw input type = %T", in.Raw())
	}
	if raw.TextFont() != fltk_bridge.COURIER || raw.TextSize() != 16 {
		t.Fatalf("editable typography = font:%v size:%d", raw.TextFont(), raw.TextSize())
	}
	if raw.TextColor() != fltk_bridge.ColorFromRgb(20, 30, 40) || raw.CursorColor() != fltk_bridge.ColorFromRgb(220, 230, 240) {
		t.Fatalf("editable colors = text:%v cursor:%v", raw.TextColor(), raw.CursorColor())
	}
}

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

func TestInputNavigationMapsNativeSearchKeys(t *testing.T) {
	in := New(0, 0, 120, 24, "Search")
	var actions []NavigationAction
	in.OnNavigation(func(action NavigationAction) bool {
		actions = append(actions, action)
		return action != NavigationCancel
	})

	tests := []struct {
		key     int
		want    NavigationAction
		handled bool
	}{
		{key: fltk_bridge.ENTER_KEY, want: NavigationSubmit, handled: true},
		{key: fltk_bridge.DOWN, want: NavigationNext, handled: true},
		{key: fltk_bridge.UP, want: NavigationPrevious, handled: true},
		{key: fltk_bridge.ESCAPE, want: NavigationCancel, handled: false},
	}
	for _, test := range tests {
		if got := in.dispatchNavigation(test.key, 0); got != test.handled {
			t.Fatalf("dispatchNavigation(%d) = %v, want %v", test.key, got, test.handled)
		}
	}
	if len(actions) != len(tests) {
		t.Fatalf("navigation actions = %v, want %d actions", actions, len(tests))
	}
	for index, test := range tests {
		if actions[index] != test.want {
			t.Fatalf("action %d = %v, want %v", index, actions[index], test.want)
		}
	}
}

func TestInputNavigationMapsFindAgainKeys(t *testing.T) {
	in := New(0, 0, 120, 24, "Search")
	var actions []NavigationAction
	in.OnNavigation(func(action NavigationAction) bool {
		actions = append(actions, action)
		return true
	})

	if !in.dispatchNavigation(fltk_bridge.F3, 0) {
		t.Fatal("F3 must dispatch next navigation")
	}
	if !in.dispatchNavigation(fltk_bridge.F3, fltk_bridge.SHIFT) {
		t.Fatal("Shift+F3 must dispatch previous navigation")
	}
	want := []NavigationAction{NavigationNext, NavigationPrevious}
	if len(actions) != len(want) || actions[0] != want[0] || actions[1] != want[1] {
		t.Fatalf("find-again actions = %#v, want %#v", actions, want)
	}
}

func TestInputNavigationLeavesTextEditingKeysNative(t *testing.T) {
	in := New(0, 0, 120, 24, "Search")
	calls := 0
	in.OnNavigation(func(NavigationAction) bool {
		calls++
		return true
	})

	if in.dispatchNavigation('x', 0) {
		t.Fatal("ordinary text key must remain available to native input")
	}
	if calls != 0 {
		t.Fatalf("ordinary key invoked navigation callback %d times", calls)
	}
}
