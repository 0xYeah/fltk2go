package view

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

func TestUIViewNilSafety(t *testing.T) {
	var v *UIView
	if v.Raw() != nil {
		t.Fatal("nil view Raw() should return nil")
	}
	if v.Superview() != nil {
		t.Fatal("nil view Superview() should return nil")
	}
	v.AddSubview(nil)
	v.RemoveFromSuperview()
}

func TestUIViewLifecycleWithoutHostIsNoop(t *testing.T) {
	v := &UIView{}
	if v.Raw() != nil {
		t.Fatal("zero UIView Raw() should return nil")
	}
	if v.Superview() != nil {
		t.Fatal("zero UIView Superview() should return nil")
	}
	v.AddSubview(nil)
	v.RemoveFromSuperview()
}

func TestUIViewTooltipUsesNativeWidgetAndAutomationMetadata(t *testing.T) {
	raw := fltk_bridge.NewButton(0, 0, 120, 36, "Open")
	defer raw.Destroy()

	v := &UIView{}
	v.BindRaw(raw)
	v.SetTooltip("Open Connection Manager (Ctrl+O)")

	if got := raw.Tooltip(); got != "Open Connection Manager (Ctrl+O)" {
		t.Fatalf("native tooltip = %q", got)
	}
	if got := v.AutomationSnapshot().Properties["tooltip"]; got != "Open Connection Manager (Ctrl+O)" {
		t.Fatalf("automation tooltip = %#v", got)
	}

	v.SetTooltip("")
	if got := raw.Tooltip(); got != "" {
		t.Fatalf("cleared native tooltip = %q", got)
	}
	if _, ok := v.AutomationSnapshot().Properties["tooltip"]; ok {
		t.Fatal("cleared tooltip must not remain in automation metadata")
	}
}

func TestUIViewTooltipNilSafety(t *testing.T) {
	var nilView *UIView
	nilView.SetTooltip("ignored")
	(&UIView{}).SetTooltip("ignored")
}
