package tabview

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/button"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

const (
	defaultTabBarHeight = 40
	indicatorHeight     = 3
)

var automationSegmentUnsafe = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

// Style controls the native tab strip without requiring application-local
// drawing patches. Zero values retain the platform-neutral defaults.
type Style struct {
	BarBackground     uint
	ContentBackground uint
	NormalText        uint
	ActiveText        uint
	Indicator         uint
	FontSize          int
}

// UITabView is a native segmented tab container with explicit dynamic-tab
// lifecycle, stable identity and semantic automation support.
type UITabView struct {
	v           view.UIView
	raw         *fltk_bridge.Group
	tabBar      *fltk_bridge.Group
	contentArea *fltk_bridge.Group
	highlight   *fltk_bridge.Box

	tabs         []*tabItem
	activeIndex  int
	automationID string
	style        Style

	onTabChanged func(index int)
}

type tabItem struct {
	id      string
	title   string
	btn     *button.UIButton
	content view.Viewable
}

func defaultStyle() Style {
	return Style{
		BarBackground:     uint(fltk_bridge.ColorFromRgb(245, 245, 245)),
		ContentBackground: uint(fltk_bridge.ColorFromRgb(255, 255, 255)),
		NormalText:        uint(fltk_bridge.ColorFromRgb(51, 51, 51)),
		ActiveText:        uint(fltk_bridge.ColorFromRgb(0, 122, 255)),
		Indicator:         uint(fltk_bridge.ColorFromRgb(0, 122, 255)),
		FontSize:          13,
	}
}

// NewUITabView creates a native tab container.
func NewUITabView(r *foundation.Rect) *UITabView {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 400, Height: 300}
	}
	style := defaultStyle()
	raw := fltk_bridge.NewGroup(r.X, r.Y, r.Width, r.Height, "")
	tabBar := fltk_bridge.NewGroup(r.X, r.Y, r.Width, defaultTabBarHeight, "")
	tabBar.SetBox(fltk_bridge.FLAT_BOX)
	tabBar.SetColor(fltk_bridge.Color(style.BarBackground))
	highlight := fltk_bridge.NewBox(fltk_bridge.FLAT_BOX, r.X, r.Y+defaultTabBarHeight-indicatorHeight, 0, indicatorHeight, "")
	highlight.SetColor(fltk_bridge.Color(style.Indicator))
	tabBar.Add(highlight)
	tabBar.End()

	contentHeight := r.Height - defaultTabBarHeight
	if contentHeight < 0 {
		contentHeight = 0
	}
	contentArea := fltk_bridge.NewGroup(r.X, r.Y+defaultTabBarHeight, r.Width, contentHeight, "")
	contentArea.SetBox(fltk_bridge.FLAT_BOX)
	contentArea.SetColor(fltk_bridge.Color(style.ContentBackground))
	contentArea.End()
	raw.End()

	tv := &UITabView{
		raw: raw, tabBar: tabBar, contentArea: contentArea, highlight: highlight,
		tabs: make([]*tabItem, 0), activeIndex: -1, style: style,
	}
	tv.v.BindRaw(raw)
	tv.v.SetAutomationRole("tablist").SetAutomationValueHandler(func() (string, bool) {
		if tv.activeIndex < 0 || tv.activeIndex >= len(tv.tabs) {
			return "", true
		}
		return tv.tabs[tv.activeIndex].id, true
	})
	return tv
}

func (tv *UITabView) View() *view.UIView {
	if tv == nil {
		return nil
	}
	return &tv.v
}

func (tv *UITabView) Raw() *fltk_bridge.Group {
	if tv == nil {
		return nil
	}
	return tv.raw
}

// SetStyle applies semantic native colors and typography to the strip.
func (tv *UITabView) SetStyle(style Style) {
	if tv == nil {
		return
	}
	defaults := defaultStyle()
	if style.BarBackground == 0 {
		style.BarBackground = defaults.BarBackground
	}
	if style.ContentBackground == 0 {
		style.ContentBackground = defaults.ContentBackground
	}
	if style.NormalText == 0 {
		style.NormalText = defaults.NormalText
	}
	if style.ActiveText == 0 {
		style.ActiveText = defaults.ActiveText
	}
	if style.Indicator == 0 {
		style.Indicator = defaults.Indicator
	}
	if style.FontSize <= 0 {
		style.FontSize = defaults.FontSize
	}
	tv.style = style
	tv.tabBar.SetColor(fltk_bridge.Color(style.BarBackground))
	tv.contentArea.SetColor(fltk_bridge.Color(style.ContentBackground))
	tv.highlight.SetColor(fltk_bridge.Color(style.Indicator))
	for _, item := range tv.tabs {
		item.btn.SetBackgroundColor(style.BarBackground)
		item.btn.Raw().SetLabelSize(style.FontSize)
	}
	tv.updateSelectionStyles()
	tv.raw.Redraw()
}

// SetAutomationID assigns the tab list and stable child ids in the form
// <id>.tab.<tab-id>.
func (tv *UITabView) SetAutomationID(id string) *UITabView {
	if tv == nil {
		return nil
	}
	tv.automationID = strings.TrimSpace(id)
	tv.v.SetAutomationID(tv.automationID)
	tv.updateAutomation()
	return tv
}

// AddTab preserves the original convenience API and assigns a stable generated id.
func (tv *UITabView) AddTab(title string, content view.Viewable) {
	if tv == nil {
		return
	}
	tv.AddTabWithID(fmt.Sprintf("tab-%d", len(tv.tabs)+1), title, content)
}

// AddTabWithID adds a tab with application-stable identity and returns its index.
// Duplicate ids select and update the existing tab rather than creating ambiguity.
func (tv *UITabView) AddTabWithID(id, title string, content view.Viewable) int {
	if tv == nil {
		return -1
	}
	id = normalizeTabID(id, len(tv.tabs)+1)
	if existing := tv.IndexOfID(id); existing >= 0 {
		tv.SetTabTitle(existing, title)
		tv.SelectTab(existing)
		return existing
	}
	item := &tabItem{id: id, title: title, content: content}
	btn := button.NewUIButton(&foundation.Rect{Width: 100, Height: defaultTabBarHeight}, title)
	item.btn = btn
	btn.Raw().SetBox(fltk_bridge.FLAT_BOX)
	btn.Raw().SetLabelSize(tv.style.FontSize)
	btn.SetBackgroundColor(tv.style.BarBackground)
	btn.View().SetAutomationRole("tab").SetAutomationProperty("tabID", id)
	btn.OnTouchUpInside(func() {
		if index := tv.indexOfItem(item); index >= 0 {
			tv.SelectTab(index)
		}
	})
	tv.tabBar.Add(btn.Raw())
	tv.v.AddAutomationChild(btn)

	if content != nil && content.View() != nil && content.View().Raw() != nil {
		if rw, ok := content.View().Raw().(interface {
			Resize(x, y, w, h int)
			Hide()
		}); ok {
			rw.Resize(tv.contentArea.X(), tv.contentArea.Y(), tv.contentArea.W(), tv.contentArea.H())
			rw.Hide()
		}
		tv.contentArea.Add(content.View().Raw())
		content.View().BindHost(tv.contentArea)
		tv.v.AddAutomationChild(content)
	}
	tv.tabs = append(tv.tabs, item)
	tv.relayoutTabs()
	tv.updateAutomation()
	if tv.activeIndex == -1 {
		tv.selectTab(0, true)
	}
	return len(tv.tabs) - 1
}

func normalizeTabID(id string, fallback int) string {
	id = automationSegmentUnsafe.ReplaceAllString(strings.TrimSpace(id), "-")
	id = strings.Trim(id, "-")
	if id == "" {
		return fmt.Sprintf("tab-%d", fallback)
	}
	return id
}

func (tv *UITabView) indexOfItem(target *tabItem) int {
	for i, item := range tv.tabs {
		if item == target {
			return i
		}
	}
	return -1
}

func (tv *UITabView) Count() int {
	if tv == nil {
		return 0
	}
	return len(tv.tabs)
}

func (tv *UITabView) ActiveIndex() int {
	if tv == nil {
		return -1
	}
	return tv.activeIndex
}

func (tv *UITabView) TabID(index int) string {
	if tv == nil || index < 0 || index >= len(tv.tabs) {
		return ""
	}
	return tv.tabs[index].id
}

func (tv *UITabView) IndexOfID(id string) int {
	if tv == nil {
		return -1
	}
	id = strings.TrimSpace(id)
	for i, item := range tv.tabs {
		if item.id == id {
			return i
		}
	}
	return -1
}

func (tv *UITabView) SetTabTitle(index int, title string) bool {
	if tv == nil || index < 0 || index >= len(tv.tabs) {
		return false
	}
	item := tv.tabs[index]
	item.title = title
	item.btn.SetTitle(title)
	tv.updateAutomation()
	return true
}

// RemoveTab removes native and semantic ownership, then selects the nearest
// remaining tab. Removing a tab before the active one preserves active identity.
func (tv *UITabView) RemoveTab(index int) bool {
	if tv == nil || index < 0 || index >= len(tv.tabs) {
		return false
	}
	item := tv.tabs[index]
	wasActive := index == tv.activeIndex
	if item.btn != nil {
		tv.tabBar.Remove(item.btn.Raw())
		tv.v.RemoveAutomationChild(item.btn)
		item.btn.View().SetAutomationID("")
		item.btn.View().BindHost(nil)
	}
	if item.content != nil && item.content.View() != nil && item.content.View().Raw() != nil {
		tv.contentArea.Remove(item.content.View().Raw())
		item.content.View().BindHost(nil)
		tv.v.RemoveAutomationChild(item.content)
	}
	tv.tabs = append(tv.tabs[:index], tv.tabs[index+1:]...)
	if len(tv.tabs) == 0 {
		tv.activeIndex = -1
		tv.highlight.Resize(tv.tabBar.X(), tv.tabBar.Y()+tv.tabBar.H()-indicatorHeight, 0, indicatorHeight)
	} else if wasActive {
		if index >= len(tv.tabs) {
			index = len(tv.tabs) - 1
		}
		tv.activeIndex = -1
		tv.selectTab(index, true)
	} else if index < tv.activeIndex {
		tv.activeIndex--
	}
	tv.relayoutTabs()
	tv.updateAutomation()
	tv.updateSelectionStyles()
	tv.raw.Redraw()
	return true
}

func (tv *UITabView) relayoutTabs() {
	count := len(tv.tabs)
	if count == 0 {
		return
	}
	baseWidth := tv.tabBar.W() / count
	x := tv.tabBar.X()
	for i, item := range tv.tabs {
		width := baseWidth
		if i == count-1 {
			width = tv.tabBar.X() + tv.tabBar.W() - x
		}
		item.btn.Raw().Resize(x, tv.tabBar.Y(), width, tv.tabBar.H()-indicatorHeight)
		x += width
	}
	tv.tabBar.Remove(tv.highlight)
	tv.tabBar.Add(tv.highlight)
	if tv.activeIndex >= 0 && tv.activeIndex < count {
		active := tv.tabs[tv.activeIndex].btn.Raw()
		tv.highlight.Resize(active.X(), tv.tabBar.Y()+tv.tabBar.H()-indicatorHeight, active.W(), indicatorHeight)
	}
}

func (tv *UITabView) SelectTab(index int) {
	if tv == nil || index < 0 || index >= len(tv.tabs) || index == tv.activeIndex {
		return
	}
	tv.selectTab(index, true)
}

func (tv *UITabView) selectTab(index int, notify bool) {
	if index < 0 || index >= len(tv.tabs) {
		return
	}
	if tv.activeIndex >= 0 && tv.activeIndex < len(tv.tabs) {
		tv.setContentVisible(tv.tabs[tv.activeIndex], false)
	}
	tv.activeIndex = index
	tv.setContentVisible(tv.tabs[index], true)
	tv.updateSelectionStyles()
	tv.relayoutTabs()
	tv.updateAutomation()
	tv.raw.Redraw()
	if notify && tv.onTabChanged != nil {
		tv.onTabChanged(index)
	}
}

func (tv *UITabView) setContentVisible(item *tabItem, visible bool) {
	if item == nil || item.content == nil || item.content.View() == nil || item.content.View().Raw() == nil {
		return
	}
	if visible {
		if rw, ok := item.content.View().Raw().(interface{ Show() }); ok {
			rw.Show()
		}
		return
	}
	if rw, ok := item.content.View().Raw().(interface{ Hide() }); ok {
		rw.Hide()
	}
}

func (tv *UITabView) updateSelectionStyles() {
	if tv == nil {
		return
	}
	for i, item := range tv.tabs {
		selected := i == tv.activeIndex
		color := tv.style.NormalText
		if selected {
			color = tv.style.ActiveText
		}
		item.btn.SetTitleColor(color)
		item.btn.View().SetAutomationProperty("selected", fmt.Sprintf("%t", selected))
	}
}

func (tv *UITabView) updateAutomation() {
	if tv == nil {
		return
	}
	for i, item := range tv.tabs {
		id := ""
		if tv.automationID != "" {
			id = tv.automationID + ".tab." + item.id
		}
		item.btn.View().SetAutomationID(id).SetAutomationName(item.title).SetAutomationProperty("index", fmt.Sprintf("%d", i))
	}
	tv.updateSelectionStyles()
}

func (tv *UITabView) OnTabChanged(cb func(index int)) {
	if tv != nil {
		tv.onTabChanged = cb
	}
}
