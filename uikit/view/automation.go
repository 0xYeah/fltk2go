package view

import (
	"errors"
	"sort"
	"strings"
	"sync"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

var (
	// ErrAutomationIDRequired is returned when an automation action is attempted without an id.
	ErrAutomationIDRequired = errors.New("automation id is required")
	// ErrAutomationNodeNotFound is returned when no registered view has the requested id.
	ErrAutomationNodeNotFound = errors.New("automation node not found")
	// ErrAutomationActionUnsupported is returned when the target view does not expose the requested action.
	ErrAutomationActionUnsupported = errors.New("automation action unsupported")
	// ErrAutomationNodeUnavailable is returned when a hidden or disabled node is targeted.
	ErrAutomationNodeUnavailable = errors.New("automation node is hidden or disabled")
)

type automationState struct {
	id       string
	name     string
	role     string
	props    map[string]string
	click    func() error
	setText  func(string) error
	getText  func() (string, bool)
	getValue func() (string, bool)
	children []*UIView
	parents  map[*UIView]struct{}
}

var automationRegistry = struct {
	sync.RWMutex
	byID map[string]*UIView
}{byID: map[string]*UIView{}}

// AutomationBounds describes a view's current FLTK bounds.
type AutomationBounds struct {
	X      int `json:"x"`
	Y      int `json:"y"`
	Width  int `json:"width"`
	Height int `json:"height"`
}

// AutomationNode is a serializable snapshot of a registered UIView.
type AutomationNode struct {
	ID         string            `json:"id,omitempty"`
	Role       string            `json:"role,omitempty"`
	Name       string            `json:"name,omitempty"`
	Label      string            `json:"label,omitempty"`
	Text       string            `json:"text,omitempty"`
	Value      string            `json:"value,omitempty"`
	Actions    []string          `json:"actions,omitempty"`
	Enabled    bool              `json:"enabled"`
	Visible    bool              `json:"visible"`
	Bounds     AutomationBounds  `json:"bounds"`
	Properties map[string]string `json:"properties,omitempty"`
	Children   []AutomationNode  `json:"children,omitempty"`
}

// SetAutomationID assigns a stable automation id and registers the view for debug automation.
func (v *UIView) SetAutomationID(id string) *UIView {
	if v == nil {
		return nil
	}
	automationRegistry.Lock()
	defer automationRegistry.Unlock()
	if v.automation.id != "" && automationRegistry.byID[v.automation.id] == v {
		delete(automationRegistry.byID, v.automation.id)
	}
	v.automation.id = id
	if id != "" {
		automationRegistry.byID[id] = v
	}
	return v
}

func (v *UIView) AutomationID() string {
	if v == nil {
		return ""
	}
	return v.automation.id
}

func (v *UIView) SetAutomationName(name string) *UIView {
	if v != nil {
		v.automation.name = name
	}
	return v
}

func (v *UIView) AutomationName() string {
	if v == nil {
		return ""
	}
	return v.automation.name
}

func (v *UIView) SetAutomationRole(role string) *UIView {
	if v != nil {
		v.automation.role = role
	}
	return v
}

func (v *UIView) AutomationRole() string {
	if v == nil {
		return ""
	}
	return v.automation.role
}

func (v *UIView) SetAutomationProperty(key, value string) *UIView {
	if v == nil || key == "" {
		return v
	}
	if v.automation.props == nil {
		v.automation.props = map[string]string{}
	}
	v.automation.props[key] = value
	return v
}

func (v *UIView) OnAutomationClick(handler func() error) *UIView {
	if v != nil {
		v.automation.click = handler
	}
	return v
}

func (v *UIView) SetAutomationTextHandlers(set func(string) error, get func() (string, bool)) *UIView {
	if v != nil {
		v.automation.setText = set
		v.automation.getText = get
	}
	return v
}

func (v *UIView) SetAutomationValueHandler(get func() (string, bool)) *UIView {
	if v != nil {
		v.automation.getValue = get
	}
	return v
}

func (v *UIView) AddAutomationChild(child Viewable) *UIView {
	if v == nil || child == nil {
		return v
	}
	cv := child.View()
	if cv == nil || cv == v {
		return v
	}
	for _, existing := range v.automation.children {
		if existing == cv {
			return v
		}
	}
	v.automation.children = append(v.automation.children, cv)
	if cv.automation.parents == nil {
		cv.automation.parents = map[*UIView]struct{}{}
	}
	cv.automation.parents[v] = struct{}{}
	return v
}

// RemoveAutomationChild removes one child from the semantic automation tree.
func (v *UIView) RemoveAutomationChild(child Viewable) *UIView {
	if v == nil || child == nil || child.View() == nil {
		return v
	}
	cv := child.View()
	children := v.automation.children[:0]
	for _, existing := range v.automation.children {
		if existing != cv {
			children = append(children, existing)
		}
	}
	v.automation.children = children
	delete(cv.automation.parents, v)
	return v
}

func (v *UIView) ClearAutomationChildren() *UIView {
	if v == nil {
		return v
	}
	for _, child := range v.automation.children {
		if child != nil {
			delete(child.automation.parents, v)
		}
	}
	v.automation.children = nil
	return v
}

func (v *UIView) clearAutomationOnDelete() {
	if v == nil {
		return
	}
	for parent := range v.automation.parents {
		parent.RemoveAutomationChild(v)
	}
	v.ClearAutomationChildren()
	v.SetAutomationID("")
	v.automation.click = nil
	v.automation.setText = nil
	v.automation.getText = nil
	v.automation.getValue = nil
	v.automation.props = nil
}

func AutomationLookup(id string) (*UIView, bool) {
	automationRegistry.RLock()
	defer automationRegistry.RUnlock()
	v, ok := automationRegistry.byID[id]
	return v, ok
}

func AutomationUnregisterPrefix(prefix string) {
	if prefix == "" {
		return
	}
	automationRegistry.Lock()
	defer automationRegistry.Unlock()
	for id := range automationRegistry.byID {
		if strings.HasPrefix(id, prefix) {
			delete(automationRegistry.byID, id)
		}
	}
}

func AutomationClick(id string) error {
	if id == "" {
		return ErrAutomationIDRequired
	}
	v, ok := AutomationLookup(id)
	if !ok || v == nil {
		return ErrAutomationNodeNotFound
	}
	if !v.automationEffectiveVisible(nil) || !v.automationEffectiveEnabled(nil) {
		return ErrAutomationNodeUnavailable
	}
	if v.automation.click == nil {
		return ErrAutomationActionUnsupported
	}
	return v.automation.click()
}

func AutomationSetText(id, text string) error {
	if id == "" {
		return ErrAutomationIDRequired
	}
	v, ok := AutomationLookup(id)
	if !ok || v == nil {
		return ErrAutomationNodeNotFound
	}
	if !v.automationEffectiveVisible(nil) || !v.automationEffectiveEnabled(nil) {
		return ErrAutomationNodeUnavailable
	}
	if v.automation.setText == nil {
		return ErrAutomationActionUnsupported
	}
	return v.automation.setText(text)
}

func AutomationSnapshot() []AutomationNode {
	automationRegistry.RLock()
	ids := make([]string, 0, len(automationRegistry.byID))
	for id := range automationRegistry.byID {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	views := make([]*UIView, 0, len(ids))
	for _, id := range ids {
		views = append(views, automationRegistry.byID[id])
	}
	automationRegistry.RUnlock()

	nodes := make([]AutomationNode, 0, len(views))
	for _, v := range views {
		if v != nil {
			nodes = append(nodes, v.AutomationSnapshot())
		}
	}
	return nodes
}

func (v *UIView) AutomationSnapshot() AutomationNode {
	if v == nil {
		return AutomationNode{}
	}
	node := AutomationNode{
		ID:      v.automation.id,
		Role:    v.automation.role,
		Name:    v.automation.name,
		Enabled: true,
		Visible: true,
	}
	if len(v.automation.props) > 0 {
		node.Properties = make(map[string]string, len(v.automation.props))
		for k, val := range v.automation.props {
			node.Properties[k] = val
		}
	}
	if v.raw != nil {
		node.Bounds = AutomationBounds{X: widgetX(v.raw), Y: widgetY(v.raw), Width: widgetW(v.raw), Height: widgetH(v.raw)}
		node.Label = widgetLabel(v.raw)
	}
	node.Enabled = v.automationEffectiveEnabled(nil)
	node.Visible = v.automationEffectiveVisible(nil)
	if v.automation.click != nil {
		node.Actions = append(node.Actions, "click")
	}
	if v.automation.setText != nil {
		node.Actions = append(node.Actions, "set_text")
	}
	if v.automation.getText != nil {
		if text, ok := v.automation.getText(); ok {
			node.Text = text
		}
	}
	if v.automation.getValue != nil {
		if value, ok := v.automation.getValue(); ok {
			node.Value = value
		}
	}
	if len(v.automation.children) > 0 {
		node.Children = make([]AutomationNode, 0, len(v.automation.children))
		for _, child := range v.automation.children {
			node.Children = append(node.Children, child.AutomationSnapshot())
		}
	}
	return node
}

func (v *UIView) automationEffectiveVisible(visiting map[*UIView]bool) bool {
	if v == nil {
		return false
	}
	if v.raw != nil && !widgetVisible(v.raw) {
		return false
	}
	if len(v.automation.parents) == 0 {
		return true
	}
	if visiting == nil {
		visiting = make(map[*UIView]bool)
	}
	if visiting[v] {
		return false
	}
	visiting[v] = true
	defer delete(visiting, v)
	for parent := range v.automation.parents {
		if parent != nil && parent.automationEffectiveVisible(visiting) {
			return true
		}
	}
	return false
}

func (v *UIView) automationEffectiveEnabled(visiting map[*UIView]bool) bool {
	if v == nil {
		return false
	}
	if v.raw != nil && !widgetActive(v.raw) {
		return false
	}
	if len(v.automation.parents) == 0 {
		return true
	}
	if visiting == nil {
		visiting = make(map[*UIView]bool)
	}
	if visiting[v] {
		return false
	}
	visiting[v] = true
	defer delete(visiting, v)
	for parent := range v.automation.parents {
		if parent != nil && parent.automationEffectiveEnabled(visiting) {
			return true
		}
	}
	return false
}

func widgetX(w fltk_bridge.Widget) int {
	if x, ok := w.(interface{ X() int }); ok {
		return x.X()
	}
	return 0
}
func widgetY(w fltk_bridge.Widget) int {
	if y, ok := w.(interface{ Y() int }); ok {
		return y.Y()
	}
	return 0
}
func widgetW(w fltk_bridge.Widget) int {
	if ww, ok := w.(interface{ W() int }); ok {
		return ww.W()
	}
	return 0
}
func widgetH(w fltk_bridge.Widget) int {
	if h, ok := w.(interface{ H() int }); ok {
		return h.H()
	}
	return 0
}
func widgetLabel(w fltk_bridge.Widget) string {
	if l, ok := w.(interface{ Label() string }); ok {
		return l.Label()
	}
	return ""
}
func widgetActive(w fltk_bridge.Widget) bool {
	if a, ok := w.(interface{ IsActive() bool }); ok {
		return a.IsActive()
	}
	return true
}
func widgetVisible(w fltk_bridge.Widget) bool {
	if v, ok := w.(interface{ Visible() bool }); ok {
		return v.Visible()
	}
	return true
}
