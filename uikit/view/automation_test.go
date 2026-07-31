package view

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func TestAutomationRegistryActionsAndSnapshot(t *testing.T) {
	var clicked bool
	v := (&UIView{}).
		SetAutomationID("demo.button").
		SetAutomationRole("button").
		SetAutomationName("Demo Button").
		SetAutomationProperty("semantic", "primary").
		OnAutomationClick(func() error {
			clicked = true
			return nil
		})
	defer v.SetAutomationID("")

	if err := AutomationClick("demo.button"); err != nil {
		t.Fatalf("AutomationClick error: %v", err)
	}
	if !clicked {
		t.Fatal("automation click handler was not invoked")
	}

	nodes := AutomationSnapshot()
	var found *AutomationNode
	for i := range nodes {
		if nodes[i].ID == "demo.button" {
			found = &nodes[i]
			break
		}
	}
	if found == nil {
		t.Fatal("registered automation node not found in snapshot")
	}
	if found.Role != "button" || found.Name != "Demo Button" || found.Properties["semantic"] != "primary" {
		t.Fatalf("unexpected snapshot: %#v", *found)
	}
	if len(found.Actions) != 1 || found.Actions[0] != "click" {
		t.Fatalf("snapshot actions = %#v, want click", found.Actions)
	}
}

func TestAutomationSetText(t *testing.T) {
	var text string
	v := (&UIView{}).SetAutomationID("demo.input").SetAutomationTextHandlers(func(s string) error {
		text = s
		return nil
	}, func() (string, bool) { return text, true })
	defer v.SetAutomationID("")

	if err := AutomationSetText("demo.input", "hello"); err != nil {
		t.Fatalf("AutomationSetText error: %v", err)
	}
	if text != "hello" {
		t.Fatalf("text = %q, want hello", text)
	}
	if got := v.AutomationSnapshot().Text; got != "hello" {
		t.Fatalf("snapshot text = %q, want hello", got)
	}
	if actions := v.AutomationSnapshot().Actions; len(actions) != 1 || actions[0] != "set_text" {
		t.Fatalf("snapshot actions = %#v, want set_text", actions)
	}
}

func TestAutomationIDLifecycleAndChildren(t *testing.T) {
	parent := (&UIView{}).SetAutomationID("demo.parent")
	child := (&UIView{}).SetAutomationID("demo.child")
	defer parent.SetAutomationID("")
	defer child.SetAutomationID("")

	parent.AddAutomationChild(child).AddAutomationChild(child)
	if node := parent.AutomationSnapshot(); len(node.Children) != 1 || node.Children[0].ID != "demo.child" {
		t.Fatalf("children snapshot = %#v", node.Children)
	}

	child.SetAutomationID("demo.child.renamed")
	if _, ok := AutomationLookup("demo.child"); ok {
		t.Fatal("old automation id still registered after rename")
	}
	if _, ok := AutomationLookup("demo.child.renamed"); !ok {
		t.Fatal("new automation id not registered after rename")
	}

	child.SetAutomationID("")
	if _, ok := AutomationLookup("demo.child.renamed"); ok {
		t.Fatal("automation id still registered after clear")
	}
	parent.ClearAutomationChildren()
	if node := parent.AutomationSnapshot(); len(node.Children) != 0 {
		t.Fatalf("children after clear = %#v", node.Children)
	}
}

func TestAutomationUnregistersWhenNativeWidgetIsDestroyed(t *testing.T) {
	raw := fltk_bridge.NewGroup(0, 0, 100, 100)
	raw.End()
	v := &UIView{}
	v.BindRaw(raw)
	v.SetAutomationID("destroyed.native.view")

	raw.Destroy()
	fltk_bridge.Check()
	if _, ok := AutomationLookup("destroyed.native.view"); ok {
		t.Fatal("destroyed native view remains in automation registry")
	}
	// A stale native node must never make global snapshot panic.
	_ = AutomationSnapshot()
	if v.Raw() != nil || v.Superview() != nil {
		t.Fatal("destroyed view retained raw widget or host")
	}
}

func TestAutomationDeletionRemovesChildFromParentsAndHandlers(t *testing.T) {
	parent := &UIView{}
	raw := fltk_bridge.NewGroup(0, 0, 100, 100)
	raw.End()
	child := &UIView{}
	child.BindRaw(raw)
	child.SetAutomationID("destroyed.automation.child")
	child.SetAutomationTextHandlers(nil, func() (string, bool) {
		panic("stale automation handler called")
	})
	parent.AddAutomationChild(child)

	raw.Destroy()
	fltk_bridge.Check()
	if children := parent.AutomationSnapshot().Children; len(children) != 0 {
		t.Fatalf("destroyed automation child remained: %#v", children)
	}
}

func TestAutomationValueHandlerAndUnregisterPrefix(t *testing.T) {
	v1 := (&UIView{}).SetAutomationID("prefix.one").SetAutomationValueHandler(func() (string, bool) {
		return "42", true
	})
	v2 := (&UIView{}).SetAutomationID("prefix.two")
	v3 := (&UIView{}).SetAutomationID("other.three")
	defer v1.SetAutomationID("")
	defer v2.SetAutomationID("")
	defer v3.SetAutomationID("")

	if got := v1.AutomationSnapshot().Value; got != "42" {
		t.Fatalf("value = %q, want 42", got)
	}
	AutomationUnregisterPrefix("prefix.")
	if _, ok := AutomationLookup("prefix.one"); ok {
		t.Fatal("prefix.one still registered")
	}
	if _, ok := AutomationLookup("prefix.two"); ok {
		t.Fatal("prefix.two still registered")
	}
	if _, ok := AutomationLookup("other.three"); !ok {
		t.Fatal("other.three should remain registered")
	}
}
