package splitview

import (
	"strconv"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

// Orientation describes whether panes are arranged left/right or top/bottom.
type Orientation int

const (
	Horizontal Orientation = iota
	Vertical
)

// PositionPolicy controls what remains stable when the split view is resized.
type PositionPolicy int

const (
	PreserveRatio PositionPolicy = iota
	PreserveLeadingSize
	PreserveTrailingSize
)

// PositionChangeReason identifies why pane geometry changed.
type PositionChangeReason string

const (
	ChangeProgrammatic PositionChangeReason = "programmatic"
	ChangeDrag         PositionChangeReason = "drag"
	ChangeResize       PositionChangeReason = "resize"
)

// SplitGeometry reports pane sizes along the split axis.
type SplitGeometry struct {
	Leading  int
	Divider  int
	Trailing int
	Ratio    float64
}

// PositionChange is delivered after an observable split-position change.
type PositionChange struct {
	Geometry SplitGeometry
	Reason   PositionChangeReason
}

// SplitView is a native two-pane container with a draggable divider.
type SplitView struct {
	raw *fltk_bridge.Group
	v   view.UIView

	leadingRaw       *fltk_bridge.Group
	trailingRaw      *fltk_bridge.Group
	dividerRaw       *fltk_bridge.Box
	leadingPane      view.UIView
	trailingPane     view.UIView
	divider          *view.UIView
	leftView         view.Viewable
	rightView        view.Viewable
	leftHostCancel   func()
	rightHostCancel  func()
	leftPreDetached  bool
	rightPreDetached bool

	orientation Orientation
	model       *layoutModel
	resizable   bool

	dividerNormal fltk_bridge.Color
	dividerHover  fltk_bridge.Color
	dividerActive fltk_bridge.Color

	isDragging      bool
	isHovering      bool
	lastMouse       int
	initialSplitPos int
	dragStart       SplitGeometry
	lastGeometry    SplitGeometry
	onPosition      func(PositionChange)
}

// New creates a split view. Horizontal arranges left/right panes; Vertical
// arranges top/bottom panes.
func New(x, y, width, height int, orientation Orientation) *SplitView {
	grp := fltk_bridge.NewGroup(x, y, width, height)
	grp.End()

	leading := fltk_bridge.NewGroup(x, y, 0, 0)
	leading.End()
	trailing := fltk_bridge.NewGroup(x, y, 0, 0)
	trailing.End()
	dividerBox := fltk_bridge.NewBox(fltk_bridge.FLAT_BOX, x, y, 0, 0, "")

	extent := width
	if orientation == Vertical {
		extent = height
	}
	sv := &SplitView{
		raw:           grp,
		leadingRaw:    leading,
		trailingRaw:   trailing,
		dividerRaw:    dividerBox,
		orientation:   orientation,
		model:         newLayoutModel(extent, 6),
		resizable:     true,
		dividerNormal: fltk_bridge.Color(0xDDDDDD00),
		dividerHover:  fltk_bridge.Color(0xCCCCCC00),
		dividerActive: fltk_bridge.Color(0xAAAAAA00),
	}
	sv.model.setMinimumSizes(50, 50)

	sv.v.BindRaw(grp)
	sv.leadingPane.BindRaw(leading)
	sv.leadingPane.BindHost(grp)
	sv.trailingPane.BindRaw(trailing)
	sv.trailingPane.BindHost(grp)
	sv.divider = &view.UIView{}
	sv.divider.BindRaw(dividerBox)
	sv.divider.BindHost(grp)
	sv.leadingPane.SetAutomationRole("split-pane-leading")
	sv.trailingPane.SetAutomationRole("split-pane-trailing")
	sv.divider.SetAutomationRole("separator")
	sv.v.AddAutomationChild(&sv.leadingPane)
	sv.v.AddAutomationChild(sv.divider)
	sv.v.AddAutomationChild(&sv.trailingPane)

	dividerBox.SetColor(sv.dividerNormal)
	sv.divider.On(fltk_bridge.PUSH, sv.onPush)
	sv.divider.On(fltk_bridge.DRAG, sv.onDrag)
	sv.divider.On(fltk_bridge.RELEASE, sv.onRelease)
	sv.divider.On(fltk_bridge.ENTER, sv.onEnter)
	sv.divider.On(fltk_bridge.LEAVE, sv.onLeave)

	grp.Add(leading)
	grp.Add(trailing)
	grp.Add(dividerBox)
	// SplitView owns layout deterministically; disable FLTK's proportional
	// child scaling so each pane/content receives exactly one final resize.
	grp.Resizable(nil)
	leading.Resizable(nil)
	trailing.Resizable(nil)
	grp.SetPreResizeHandler(sv.prepareForNativeResize)
	grp.SetResizeHandler(func() { sv.layout(ChangeResize, true) })
	sv.layout("", false)
	return sv
}

// SetLeftView sets the leading (left/top) pane content.
func (sv *SplitView) SetLeftView(content view.Viewable) {
	if sv == nil || sv.leadingRaw == nil {
		return
	}
	sv.clearLeftContent(true)
	if sameView(content, sv.rightView) {
		sv.clearRightContent(true)
	}
	if !validContent(content) {
		sv.layout("", false)
		return
	}
	addPaneContent(sv.leadingRaw, content)
	sv.leftView = content
	sv.leadingPane.AddAutomationChild(content)
	expected := content
	sv.leftHostCancel = content.View().ObserveSuperviewChanges(func(_, next view.Container) {
		if next != sv.leadingRaw && sameView(expected, sv.leftView) {
			sv.releaseLeftReference(expected)
		}
	})
	sv.layout("", false)
}

// SetRightView sets the trailing (right/bottom) pane content.
func (sv *SplitView) SetRightView(content view.Viewable) {
	if sv == nil || sv.trailingRaw == nil {
		return
	}
	sv.clearRightContent(true)
	if sameView(content, sv.leftView) {
		sv.clearLeftContent(true)
	}
	if !validContent(content) {
		sv.layout("", false)
		return
	}
	addPaneContent(sv.trailingRaw, content)
	sv.rightView = content
	sv.trailingPane.AddAutomationChild(content)
	expected := content
	sv.rightHostCancel = content.View().ObserveSuperviewChanges(func(_, next view.Container) {
		if next != sv.trailingRaw && sameView(expected, sv.rightView) {
			sv.releaseRightReference(expected)
		}
	})
	sv.layout("", false)
}

func validContent(content view.Viewable) bool {
	return content != nil && content.View() != nil && content.View().Raw() != nil
}

func sameView(a, b view.Viewable) bool {
	return a != nil && b != nil && a.View() != nil && a.View() == b.View()
}

func (sv *SplitView) clearLeftContent(removeNative bool) {
	if sv.leftHostCancel != nil {
		sv.leftHostCancel()
		sv.leftHostCancel = nil
	}
	content := sv.leftView
	sv.leftView = nil
	sv.leadingPane.RemoveAutomationChild(content)
	if removeNative {
		removePaneContent(sv.leadingRaw, content)
	}
}

func (sv *SplitView) clearRightContent(removeNative bool) {
	if sv.rightHostCancel != nil {
		sv.rightHostCancel()
		sv.rightHostCancel = nil
	}
	content := sv.rightView
	sv.rightView = nil
	sv.trailingPane.RemoveAutomationChild(content)
	if removeNative {
		removePaneContent(sv.trailingRaw, content)
	}
}

func (sv *SplitView) releaseLeftReference(content view.Viewable) {
	if sameView(content, sv.leftView) {
		sv.clearLeftContent(false)
	}
}

func (sv *SplitView) releaseRightReference(content view.Viewable) {
	if sameView(content, sv.rightView) {
		sv.clearRightContent(false)
	}
}

func removePaneContent(host *fltk_bridge.Group, content view.Viewable) {
	if host == nil || !validContent(content) {
		return
	}
	host.Remove(content.View().Raw())
	if content.View().Superview() == host {
		content.View().BindHost(nil)
	}
}

func addPaneContent(host *fltk_bridge.Group, content view.Viewable) {
	if host == nil || !validContent(content) {
		return
	}
	if previous := content.View().Superview(); previous != nil && previous != host {
		previous.Remove(content.View().Raw())
	}
	host.Add(content.View().Raw())
	content.View().BindHost(host)
}

// LeadingPane returns the stable internal leading-pane container.
func (sv *SplitView) LeadingPane() *view.UIView {
	if sv == nil {
		return nil
	}
	return &sv.leadingPane
}

// TrailingPane returns the stable internal trailing-pane container.
func (sv *SplitView) TrailingPane() *view.UIView {
	if sv == nil {
		return nil
	}
	return &sv.trailingPane
}

// DividerView exposes divider automation and appearance hooks.
func (sv *SplitView) DividerView() *view.UIView {
	if sv == nil {
		return nil
	}
	return sv.divider
}

// SetLeftViewFixed preserves a fixed leading size across parent resizes.
func (sv *SplitView) SetLeftViewFixed(size int) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setLeadingPixels(size)
	sv.layout(ChangeProgrammatic, true)
}

// SetRightViewFixed preserves a fixed trailing size across parent resizes.
func (sv *SplitView) SetRightViewFixed(size int) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setTrailingPixels(size)
	sv.layout(ChangeProgrammatic, true)
}

// SetPosition sets a ratio-based position and preserves that ratio on resize.
func (sv *SplitView) SetPosition(pos float64) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setRatio(pos)
	sv.layout(ChangeProgrammatic, true)
}

// SetPositionPixels sets the leading pane size in pixels.
func (sv *SplitView) SetPositionPixels(position int) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setLeadingPixels(position)
	sv.layout(ChangeProgrammatic, true)
}

// SetPositionPolicy changes resize behavior without changing current geometry.
func (sv *SplitView) SetPositionPolicy(policy PositionPolicy) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setPolicy(positionPolicy(policy))
}

// Position returns the current leading position as a ratio of total extent.
func (sv *SplitView) Position() float64 { return sv.Geometry().Ratio }

// PositionPixels returns the leading pane size in pixels.
func (sv *SplitView) PositionPixels() int { return sv.Geometry().Leading }

// Geometry returns current sizes along the split axis.
func (sv *SplitView) Geometry() SplitGeometry {
	if sv == nil || sv.model == nil {
		return SplitGeometry{}
	}
	return publicGeometry(sv.model.geometry())
}

// SetMinimumSizes configures independent leading and trailing minima.
func (sv *SplitView) SetMinimumSizes(leading, trailing int) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setMinimumSizes(leading, trailing)
	sv.layout(ChangeProgrammatic, true)
}

// SetDividerSize configures the divider thickness in pixels.
func (sv *SplitView) SetDividerSize(size int) {
	if sv == nil || sv.model == nil {
		return
	}
	sv.model.setDividerSize(size)
	sv.layout(ChangeProgrammatic, true)
}

// SetDividerColors configures normal, hover, and active divider colors.
func (sv *SplitView) SetDividerColors(normal, hover, active uint) {
	if sv == nil {
		return
	}
	sv.dividerNormal = fltk_bridge.Color(normal)
	sv.dividerHover = fltk_bridge.Color(hover)
	sv.dividerActive = fltk_bridge.Color(active)
	sv.setDividerColor(sv.dividerNormal)
}

// SetResizable enables or disables divider dragging.
func (sv *SplitView) SetResizable(enabled bool) {
	if sv == nil {
		return
	}
	sv.resizable = enabled
	if !enabled {
		wasDragging := sv.isDragging
		sv.isDragging = false
		geometry := sv.Geometry()
		sv.setDividerColor(sv.dividerNormal)
		sv.setResizeCursor(false)
		if wasDragging && geometry != sv.dragStart && sv.onPosition != nil {
			sv.onPosition(PositionChange{Geometry: geometry, Reason: ChangeDrag})
		}
	} else if sv.isHovering {
		sv.setDividerColor(sv.dividerHover)
		sv.setResizeCursor(true)
	}
	sv.updateAutomationGeometry(sv.Geometry())
}

// Resizable reports whether divider dragging is enabled.
func (sv *SplitView) Resizable() bool { return sv != nil && sv.resizable }

// OnPositionChanged installs the split-position callback.
func (sv *SplitView) OnPositionChanged(handler func(PositionChange)) {
	if sv != nil {
		sv.onPosition = handler
	}
}

// SetAutomationID assigns stable IDs to the split view and its parts.
func (sv *SplitView) SetAutomationID(id string) *SplitView {
	if sv == nil {
		return nil
	}
	sv.v.SetAutomationID(id).SetAutomationRole("split-view")
	if id == "" {
		sv.leadingPane.SetAutomationID("")
		sv.divider.SetAutomationID("")
		sv.trailingPane.SetAutomationID("")
		return sv
	}
	sv.leadingPane.SetAutomationID(id + ".leading")
	sv.divider.SetAutomationID(id + ".divider")
	sv.trailingPane.SetAutomationID(id + ".trailing")
	sv.updateAutomationGeometry(sv.Geometry())
	return sv
}

func (sv *SplitView) View() *view.UIView {
	if sv == nil {
		return nil
	}
	return &sv.v
}

func (sv *SplitView) Raw() fltk_bridge.Widget {
	if sv == nil {
		return nil
	}
	return sv.raw
}

func (sv *SplitView) On(event fltk_bridge.Event, handler func(fltk_bridge.Event) bool) {
	if sv != nil {
		sv.v.On(event, handler)
	}
}

func (sv *SplitView) AddSubview(child view.Viewable) {
	if sv != nil {
		sv.v.AddSubview(child)
	}
}

func (sv *SplitView) prepareForNativeResize() {
	if sv == nil {
		return
	}
	sv.leftPreDetached = detachPaneContent(sv.leadingRaw, sv.leftView)
	sv.rightPreDetached = detachPaneContent(sv.trailingRaw, sv.rightView)
}

func (sv *SplitView) layout(reason PositionChangeReason, notify bool) {
	if sv == nil || sv.raw == nil || sv.model == nil {
		return
	}
	extent := sv.raw.W()
	if sv.orientation == Vertical {
		extent = sv.raw.H()
	}
	sv.model.setExtent(extent)
	geometry := publicGeometry(sv.model.geometry())
	x, y, w, h := sv.raw.X(), sv.raw.Y(), sv.raw.W(), sv.raw.H()

	leftDetached := sv.leftPreDetached || detachPaneContent(sv.leadingRaw, sv.leftView)
	rightDetached := sv.rightPreDetached || detachPaneContent(sv.trailingRaw, sv.rightView)
	sv.leftPreDetached = false
	sv.rightPreDetached = false
	if sv.orientation == Horizontal {
		sv.leadingRaw.Resize(x, y, geometry.Leading, h)
		sv.dividerRaw.Resize(x+geometry.Leading, y, geometry.Divider, h)
		sv.trailingRaw.Resize(x+geometry.Leading+geometry.Divider, y, geometry.Trailing, h)
	} else {
		sv.leadingRaw.Resize(x, y, w, geometry.Leading)
		sv.dividerRaw.Resize(x, y+geometry.Leading, w, geometry.Divider)
		sv.trailingRaw.Resize(x, y+geometry.Leading+geometry.Divider, w, geometry.Trailing)
	}
	reattachPaneContent(sv.leadingRaw, sv.leftView, leftDetached)
	reattachPaneContent(sv.trailingRaw, sv.rightView, rightDetached)
	resizePaneContent(sv.leadingRaw, sv.leftView)
	resizePaneContent(sv.trailingRaw, sv.rightView)
	sv.leadingRaw.Redraw()
	sv.dividerRaw.Redraw()
	sv.trailingRaw.Redraw()
	sv.updateAutomationGeometry(geometry)

	previous := sv.lastGeometry
	sv.lastGeometry = geometry
	if notify && geometry != previous && sv.onPosition != nil {
		sv.onPosition(PositionChange{Geometry: geometry, Reason: reason})
	}
}

func detachPaneContent(host *fltk_bridge.Group, content view.Viewable) bool {
	if host == nil || !validContent(content) || content.View().Superview() != host {
		return false
	}
	host.Remove(content.View().Raw())
	return true
}

func reattachPaneContent(host *fltk_bridge.Group, content view.Viewable, detached bool) {
	if detached && host != nil && validContent(content) && content.View().Superview() == host {
		host.Add(content.View().Raw())
	}
}

func resizePaneContent(host *fltk_bridge.Group, content view.Viewable) {
	if host == nil || content == nil || content.View() == nil || content.View().Raw() == nil {
		return
	}
	if raw, ok := content.View().Raw().(interface{ Resize(x, y, width, height int) }); ok {
		raw.Resize(host.X(), host.Y(), host.W(), host.H())
	}
}

func publicGeometry(g splitGeometry) SplitGeometry {
	return SplitGeometry{Leading: g.Leading, Divider: g.Divider, Trailing: g.Trailing, Ratio: g.Ratio}
}

func (sv *SplitView) updateAutomationGeometry(g SplitGeometry) {
	if sv == nil || sv.divider == nil {
		return
	}
	sv.v.SetAutomationProperty("orientation", sv.orientationName())
	sv.v.SetAutomationProperty("ratio", strconv.FormatFloat(g.Ratio, 'f', 6, 64))
	sv.leadingPane.SetAutomationProperty("size", strconv.Itoa(g.Leading))
	sv.divider.SetAutomationProperty("size", strconv.Itoa(g.Divider))
	sv.divider.SetAutomationProperty("resizable", strconv.FormatBool(sv.resizable))
	sv.trailingPane.SetAutomationProperty("size", strconv.Itoa(g.Trailing))
}

func (sv *SplitView) orientationName() string {
	if sv.orientation == Vertical {
		return "vertical"
	}
	return "horizontal"
}

func (sv *SplitView) onPush(fltk_bridge.Event) bool {
	if sv == nil || !sv.resizable {
		return false
	}
	sv.isDragging = true
	sv.isHovering = true
	sv.lastMouse = sv.eventAxisRoot()
	sv.initialSplitPos = sv.Geometry().Leading
	sv.dragStart = sv.Geometry()
	sv.setDividerColor(sv.dividerActive)
	return true
}

func (sv *SplitView) onDrag(fltk_bridge.Event) bool {
	if sv == nil || !sv.resizable || !sv.isDragging {
		return false
	}
	delta := sv.eventAxisRoot() - sv.lastMouse
	sv.model.setLeadingForCurrentPolicy(sv.initialSplitPos + delta)
	sv.layout("", false)
	return true
}

func (sv *SplitView) onRelease(fltk_bridge.Event) bool {
	if sv == nil || !sv.resizable || !sv.isDragging {
		return false
	}
	sv.isDragging = false
	geometry := sv.Geometry()
	if sv.isHovering {
		sv.setDividerColor(sv.dividerHover)
		sv.setResizeCursor(true)
	} else {
		sv.setDividerColor(sv.dividerNormal)
		sv.setResizeCursor(false)
	}
	if geometry != sv.dragStart && sv.onPosition != nil {
		sv.onPosition(PositionChange{Geometry: geometry, Reason: ChangeDrag})
	}
	return true
}

func (sv *SplitView) onEnter(fltk_bridge.Event) bool {
	if sv == nil || !sv.resizable {
		return false
	}
	sv.isHovering = true
	sv.setResizeCursor(true)
	if !sv.isDragging {
		sv.setDividerColor(sv.dividerHover)
	}
	return true
}

func (sv *SplitView) onLeave(fltk_bridge.Event) bool {
	if sv == nil {
		return false
	}
	sv.isHovering = false
	if sv.isDragging {
		sv.setResizeCursor(true)
		return sv.resizable
	}
	sv.setResizeCursor(false)
	sv.setDividerColor(sv.dividerNormal)
	return sv.resizable
}

func (sv *SplitView) eventAxisRoot() int {
	if sv.orientation == Vertical {
		return fltk_bridge.EventYRoot()
	}
	return fltk_bridge.EventXRoot()
}

func (sv *SplitView) setDividerColor(color fltk_bridge.Color) {
	if sv == nil || sv.divider == nil || sv.divider.Raw() == nil {
		return
	}
	sv.dividerRaw.SetColor(color)
	sv.dividerRaw.Redraw()
}

func (sv *SplitView) setResizeCursor(resize bool) {
	if sv == nil || sv.raw == nil {
		return
	}
	cursor := fltk_bridge.CURSOR_DEFAULT
	if resize {
		if sv.orientation == Vertical {
			cursor = fltk_bridge.CURSOR_NS
		} else {
			cursor = fltk_bridge.CURSOR_WE
		}
	}
	sv.raw.SetWindowCursor(cursor)
}
