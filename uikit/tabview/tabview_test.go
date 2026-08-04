package tabview_test

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/tabview"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

// mockViewable 实现一个简单的 Viewable
type mockViewable struct {
	v view.UIView
}

func newMockViewable() *mockViewable {
	m := &mockViewable{}
	raw := fltk_bridge.NewGroup(0, 0, 100, 100, "")
	m.v.BindRaw(raw)
	return m
}

func (m *mockViewable) View() *view.UIView {
	return &m.v
}

func TestUITabView_Creation(t *testing.T) {
	tv := tabview.NewUITabView(&foundation.Rect{X: 10, Y: 10, Width: 400, Height: 300})
	if tv == nil {
		t.Fatal("Expected UITabView to be created")
	}
	if tv.Raw() == nil {
		t.Fatal("Expected Raw() to return underlying Group widget")
	}
	if tv.View() == nil {
		t.Fatal("Expected View() to return UIView")
	}
}

func TestUITabView_AddTab(t *testing.T) {
	tv := tabview.NewUITabView(nil)

	child1 := newMockViewable()
	child2 := newMockViewable()

	tv.AddTab("Tab 1", child1)
	tv.AddTab("Tab 2", child2)

	// Since we added two tabs, activeIndex should be 0 by default
	// and there should be multiple children in the tab bar and content area
	if tv.Raw().ChildCount() < 2 {
		t.Fatalf("Expected UITabView to have children")
	}
}

func TestUITabView_SelectTab(t *testing.T) {
	tv := tabview.NewUITabView(nil)

	child1 := newMockViewable()
	child2 := newMockViewable()

	tv.AddTab("Tab 1", child1)
	tv.AddTab("Tab 2", child2)

	called := false
	tv.OnTabChanged(func(idx int) {
		if idx == 1 {
			called = true
		}
	})

	tv.SelectTab(1)
	if !called {
		t.Fatal("Expected OnTabChanged callback to be triggered for index 1")
	}
	if tv.ActiveIndex() != 1 || tv.Count() != 2 {
		t.Fatalf("unexpected tab state: active=%d count=%d", tv.ActiveIndex(), tv.Count())
	}
}

func TestUITabViewStableIDsAndDynamicRemoval(t *testing.T) {
	tv := tabview.NewUITabView(nil)
	tv.SetAutomationID("sessions")
	tv.AddTabWithID("local", "Local", newMockViewable())
	tv.AddTabWithID("server", "Server", newMockViewable())
	tv.AddTabWithID("logs", "Logs", newMockViewable())

	if got := tv.IndexOfID("server"); got != 1 {
		t.Fatalf("IndexOfID(server) = %d, want 1", got)
	}
	if _, ok := view.AutomationLookup("sessions.tab.server"); !ok {
		t.Fatal("stable semantic tab id was not registered")
	}

	tv.SelectTab(1)
	if !tv.RemoveTab(0) {
		t.Fatal("RemoveTab(0) failed")
	}
	if tv.ActiveIndex() != 0 || tv.TabID(0) != "server" {
		t.Fatalf("active tab identity changed after removing preceding tab: active=%d id=%q", tv.ActiveIndex(), tv.TabID(0))
	}
	if _, ok := view.AutomationLookup("sessions.tab.local"); ok {
		t.Fatal("removed tab remained in automation registry")
	}

	if !tv.SetTabTitle(0, "Server · 运行中") {
		t.Fatal("SetTabTitle failed")
	}
	if node, ok := view.AutomationLookup("sessions.tab.server"); !ok || node.AutomationName() != "Server · 运行中" {
		t.Fatal("updated title was not published semantically")
	}
}

func TestUITabViewRemovalSelectsNearestTabAndNotifies(t *testing.T) {
	tv := tabview.NewUITabView(nil)
	tv.AddTabWithID("one", "One", nil)
	tv.AddTabWithID("two", "Two", nil)
	tv.AddTabWithID("three", "Three", nil)
	tv.SelectTab(2)

	var changed []int
	tv.OnTabChanged(func(index int) { changed = append(changed, index) })
	if !tv.RemoveTab(2) {
		t.Fatal("RemoveTab(active) failed")
	}
	if tv.ActiveIndex() != 1 || tv.TabID(1) != "two" {
		t.Fatalf("expected nearest remaining tab, active=%d id=%q", tv.ActiveIndex(), tv.TabID(tv.ActiveIndex()))
	}
	if len(changed) != 1 || changed[0] != 1 {
		t.Fatalf("unexpected selection callbacks: %#v", changed)
	}

	tv.RemoveTab(1)
	tv.RemoveTab(0)
	if tv.ActiveIndex() != -1 || tv.Count() != 0 {
		t.Fatalf("empty tab view state is invalid: active=%d count=%d", tv.ActiveIndex(), tv.Count())
	}
}
