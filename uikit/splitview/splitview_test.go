package splitview

import (
	"math"
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/group"
	"github.com/0xdevelop/fltk2go/uikit/input"
	"github.com/0xdevelop/fltk2go/uikit/label"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

func TestSplitViewOwnsAndReflowsPaneContent(t *testing.T) {
	sv := New(10, 20, 900, 500, Horizontal)
	left := label.NewUILabel(&foundation.Rect{}, "left")
	right := label.NewUILabel(&foundation.Rect{}, "right")
	sv.SetLeftView(left)
	sv.SetRightView(right)
	sv.SetDividerSize(8)
	sv.SetMinimumSizes(250, 320)
	sv.SetPosition(0.4)

	got := sv.Geometry()
	if got.Leading != 360 || got.Divider != 8 || got.Trailing != 532 {
		t.Fatalf("geometry = %#v, want 360/8/532", got)
	}
	assertWidgetGeometry(t, left.View().Raw(), 10, 20, 360, 500)
	assertWidgetGeometry(t, right.View().Raw(), 378, 20, 532, 500)

	sv.raw.Resize(10, 20, 1200, 500)
	got = sv.Geometry()
	if got.Leading != 480 || got.Trailing != 712 || math.Abs(got.Ratio-0.4) > 0.002 {
		t.Fatalf("resized geometry = %#v", got)
	}
	assertWidgetGeometry(t, left.View().Raw(), 10, 20, 480, 500)
	assertWidgetGeometry(t, right.View().Raw(), 498, 20, 712, 500)
}

func TestVerticalSplitViewReflowsTopAndBottomContent(t *testing.T) {
	sv := New(10, 20, 400, 600, Vertical)
	top := label.NewUILabel(&foundation.Rect{}, "top")
	bottom := label.NewUILabel(&foundation.Rect{}, "bottom")
	sv.SetLeftView(top)
	sv.SetRightView(bottom)
	sv.SetDividerSize(10)
	sv.SetMinimumSizes(100, 120)
	sv.SetPosition(0.25)

	if got := sv.Geometry(); got.Leading != 150 || got.Divider != 10 || got.Trailing != 440 {
		t.Fatalf("vertical geometry = %#v", got)
	}
	assertWidgetGeometry(t, top.View().Raw(), 10, 20, 400, 150)
	assertWidgetGeometry(t, bottom.View().Raw(), 10, 180, 400, 440)
}

func TestSplitViewResizesPaneContentExactlyOnce(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal)
	left := group.NewUIGroup(&foundation.Rect{})
	right := group.NewUIGroup(&foundation.Rect{})
	leftCount, rightCount := 0, 0
	left.Raw().SetResizeHandler(func() { leftCount++ })
	right.Raw().SetResizeHandler(func() { rightCount++ })
	sv.SetLeftView(left)
	sv.SetRightView(right)
	leftCount, rightCount = 0, 0

	// Moving and resizing the outer group exercises FLTK's automatic child
	// translation path as well as SplitView's explicit final layout.
	sv.raw.Resize(25, 35, 1000, 520)
	if leftCount != 1 || rightCount != 1 {
		t.Fatalf("outer resize counts = left:%d right:%d, want exactly 1 each", leftCount, rightCount)
	}

	leftCount, rightCount = 0, 0
	sv.SetPositionPixels(600)
	if leftCount != 1 || rightCount != 1 {
		t.Fatalf("divider layout counts = left:%d right:%d, want exactly 1 each", leftCount, rightCount)
	}
}

func TestSplitViewResizableCanBeDisabled(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal)
	if !sv.Resizable() {
		t.Fatal("new split view should be resizable")
	}
	sv.SetResizable(false)
	if sv.Resizable() {
		t.Fatal("SetResizable(false) did not disable resizing")
	}
	before := sv.Geometry()
	if sv.onPush(0) {
		t.Fatal("disabled divider accepted push")
	}
	if got := sv.Geometry(); got != before {
		t.Fatalf("disabled divider changed geometry: before=%#v after=%#v", before, got)
	}
}

func TestDisablingDuringDragCommitsOneChange(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal)
	start := sv.Geometry()
	var changes []PositionChange
	sv.OnPositionChanged(func(change PositionChange) { changes = append(changes, change) })
	sv.isDragging = true
	sv.dragStart = start
	sv.model.setLeadingForCurrentPolicy(start.Leading + 80)
	sv.layout("", false)

	sv.SetResizable(false)
	if len(changes) != 1 || changes[0].Reason != ChangeDrag {
		t.Fatalf("disable-during-drag changes = %#v", changes)
	}
	if sv.isDragging || sv.Resizable() {
		t.Fatal("dragging or resizing remained enabled")
	}
	if sv.onRelease(0) {
		t.Fatal("release after disabling should not be handled twice")
	}
}

func TestDraggingKeepsResizeStateWhenPointerLeaves(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal)
	sv.isDragging = true
	sv.isHovering = true
	if !sv.onLeave(0) {
		t.Fatal("drag leave was not handled")
	}
	if sv.isHovering || !sv.isDragging {
		t.Fatalf("unexpected drag/hover state: dragging=%t hovering=%t", sv.isDragging, sv.isHovering)
	}
}

func TestSplitViewCallbacksDescribeProgrammaticAndResizeChanges(t *testing.T) {
	sv := New(0, 0, 1000, 500, Horizontal)
	sv.SetMinimumSizes(100, 100)
	var changes []PositionChange
	sv.OnPositionChanged(func(change PositionChange) { changes = append(changes, change) })

	sv.SetPosition(0.3)
	if len(changes) != 1 || changes[0].Reason != ChangeProgrammatic {
		t.Fatalf("programmatic changes = %#v", changes)
	}
	sv.raw.Resize(0, 0, 1200, 500)
	if len(changes) != 2 || changes[1].Reason != ChangeResize {
		t.Fatalf("resize changes = %#v", changes)
	}
	if math.Abs(changes[1].Geometry.Ratio-0.3) > 0.002 {
		t.Fatalf("resize callback ratio = %f", changes[1].Geometry.Ratio)
	}
}

func TestSplitViewDragCallbackFiresOnReleaseAndPreservesRatio(t *testing.T) {
	sv := New(0, 0, 1000, 500, Horizontal)
	sv.SetMinimumSizes(100, 100)
	start := sv.Geometry()
	var change PositionChange
	sv.OnPositionChanged(func(got PositionChange) { change = got })

	sv.isDragging = true
	sv.dragStart = start
	sv.model.setLeadingForCurrentPolicy(start.Leading + 120)
	sv.layout("", false)
	if !sv.onRelease(0) {
		t.Fatal("release was not handled")
	}
	if change.Reason != ChangeDrag || change.Geometry.Leading != start.Leading+120 {
		t.Fatalf("drag change = %#v", change)
	}
	beforeResize := sv.Geometry().Ratio
	sv.raw.Resize(0, 0, 1200, 500)
	if math.Abs(sv.Geometry().Ratio-beforeResize) > 0.002 {
		t.Fatalf("drag ratio was not preserved on resize: before=%f after=%f", beforeResize, sv.Geometry().Ratio)
	}
}

func TestSplitViewReparentsContentAndMaintainsAutomationTree(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal).SetAutomationID("workspace.tree")
	content := label.NewUILabel(&foundation.Rect{}, "content")
	content.View().SetAutomationID("workspace.content")
	replacement := label.NewUILabel(&foundation.Rect{}, "replacement")
	replacement.View().SetAutomationID("workspace.replacement")
	defer sv.SetAutomationID("")
	defer content.View().SetAutomationID("")
	defer replacement.View().SetAutomationID("")

	sv.SetLeftView(content)
	leading := sv.LeadingPane().AutomationSnapshot()
	if len(leading.Children) != 1 || leading.Children[0].ID != "workspace.content" {
		t.Fatalf("leading automation children = %#v", leading.Children)
	}

	sv.SetRightView(content)
	if sv.leftView != nil || content.View().Superview() != sv.trailingRaw {
		t.Fatal("content was not atomically reparented to trailing pane")
	}
	if children := sv.LeadingPane().AutomationSnapshot().Children; len(children) != 0 {
		t.Fatalf("leading children after reparent = %#v", children)
	}
	if children := sv.TrailingPane().AutomationSnapshot().Children; len(children) != 1 || children[0].ID != "workspace.content" {
		t.Fatalf("trailing children after reparent = %#v", children)
	}

	sv.SetRightView(replacement)
	if content.View().Superview() != nil {
		t.Fatal("replaced content retained stale superview")
	}
	if children := sv.TrailingPane().AutomationSnapshot().Children; len(children) != 1 || children[0].ID != "workspace.replacement" {
		t.Fatalf("trailing children after replacement = %#v", children)
	}
}

func TestSplitViewCrossContainerReparentReleasesOldOwner(t *testing.T) {
	first := New(0, 0, 800, 500, Horizontal).SetAutomationID("first.split")
	second := New(100, 80, 900, 600, Horizontal).SetAutomationID("second.split")
	content := label.NewUILabel(&foundation.Rect{}, "shared")
	content.View().SetAutomationID("shared.content")
	defer first.SetAutomationID("")
	defer second.SetAutomationID("")
	defer content.View().SetAutomationID("")

	first.SetLeftView(content)
	second.SetRightView(content)
	if first.leftView != nil || second.rightView != content {
		t.Fatal("cross-split reparent retained stale ownership")
	}
	if content.View().Superview() != second.trailingRaw {
		t.Fatal("cross-split reparent did not bind the new host")
	}
	if children := first.LeadingPane().AutomationSnapshot().Children; len(children) != 0 {
		t.Fatalf("old automation pane retained content: %#v", children)
	}
	before := widgetGeometry(t, content.View().Raw())
	first.raw.Resize(0, 0, 1200, 700)
	after := widgetGeometry(t, content.View().Raw())
	if after != before {
		t.Fatalf("old SplitView mutated reparented content: before=%v after=%v", before, after)
	}
}

func TestExternalRemoveFromSuperviewReleasesSplitViewOwnership(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal)
	content := label.NewUILabel(&foundation.Rect{}, "external removal")
	sv.SetLeftView(content)
	content.View().RemoveFromSuperview()

	if sv.leftView != nil || content.View().Superview() != nil {
		t.Fatal("RemoveFromSuperview retained SplitView ownership")
	}
	if children := sv.LeadingPane().AutomationSnapshot().Children; len(children) != 0 {
		t.Fatalf("RemoveFromSuperview retained automation child: %#v", children)
	}
	before := widgetGeometry(t, content.View().Raw())
	sv.raw.Resize(10, 20, 1000, 600)
	if after := widgetGeometry(t, content.View().Raw()); after != before {
		t.Fatalf("old SplitView resized externally removed content: before=%v after=%v", before, after)
	}
}

func TestDestroyedPaneContentIsRemovedFromOwnershipAndAutomation(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal).SetAutomationID("content.delete.split")
	content := input.New(0, 0, 100, 30, "value")
	content.View().SetAutomationID("content.delete.input")
	sv.SetLeftView(content)

	destroyer, ok := content.View().Raw().(interface{ Destroy() })
	if !ok {
		t.Fatalf("%T cannot be destroyed", content.View().Raw())
	}
	destroyer.Destroy()
	fltk_bridge.Check()
	if sv.leftView != nil {
		t.Fatal("destroyed pane content remained owned by SplitView")
	}
	if children := sv.LeadingPane().AutomationSnapshot().Children; len(children) != 0 {
		t.Fatalf("destroyed content remained in automation tree: %#v", children)
	}
	_ = sv.View().AutomationSnapshot() // stale handlers must not panic.
}

func TestSplitViewAutomationUnregistersOnNativeDestruction(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal).SetAutomationID("destroyed.split")
	sv.raw.Destroy()
	fltk_bridge.Check()

	for _, id := range []string{"destroyed.split", "destroyed.split.leading", "destroyed.split.divider", "destroyed.split.trailing"} {
		if _, ok := view.AutomationLookup(id); ok {
			t.Fatalf("automation id %q remained after native destruction", id)
		}
	}
}

func TestSplitViewAutomationDescribesDividerAndPanes(t *testing.T) {
	sv := New(0, 0, 800, 500, Horizontal).SetAutomationID("workspace.split")
	defer sv.SetAutomationID("")
	sv.SetMinimumSizes(220, 300)
	sv.SetDividerSize(7)
	sv.SetPosition(0.42)

	divider := sv.DividerView().AutomationSnapshot()
	if divider.ID != "workspace.split.divider" || divider.Role != "separator" {
		t.Fatalf("divider automation = %#v", divider)
	}
	if divider.Properties["size"] != "7" || divider.Properties["resizable"] != "true" {
		t.Fatalf("divider properties = %#v", divider.Properties)
	}
	if sv.LeadingPane().AutomationSnapshot().ID != "workspace.split.leading" {
		t.Fatalf("leading pane automation id not assigned")
	}
	if sv.TrailingPane().AutomationSnapshot().ID != "workspace.split.trailing" {
		t.Fatalf("trailing pane automation id not assigned")
	}
}

type geometryWidget interface {
	X() int
	Y() int
	W() int
	H() int
}

type geometrySnapshot struct {
	x, y, width, height int
}

func widgetGeometry(t *testing.T, raw any) geometrySnapshot {
	t.Helper()
	widget, ok := raw.(geometryWidget)
	if !ok {
		t.Fatalf("%T does not expose widget geometry", raw)
	}
	return geometrySnapshot{x: widget.X(), y: widget.Y(), width: widget.W(), height: widget.H()}
}

func assertWidgetGeometry(t *testing.T, raw any, x, y, width, height int) {
	t.Helper()
	widget, ok := raw.(geometryWidget)
	if !ok {
		t.Fatalf("%T does not expose widget geometry", raw)
	}
	if widget.X() != x || widget.Y() != y || widget.W() != width || widget.H() != height {
		t.Fatalf("geometry = (%d,%d %dx%d), want (%d,%d %dx%d)", widget.X(), widget.Y(), widget.W(), widget.H(), x, y, width, height)
	}
}
