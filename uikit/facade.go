package uikit

import (
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/button"
	"github.com/0xdevelop/fltk2go/uikit/dialog"
	"github.com/0xdevelop/fltk2go/uikit/group"
	"github.com/0xdevelop/fltk2go/uikit/input"
	"github.com/0xdevelop/fltk2go/uikit/label"
	"github.com/0xdevelop/fltk2go/uikit/menubar"
	"github.com/0xdevelop/fltk2go/uikit/progress"
	"github.com/0xdevelop/fltk2go/uikit/scrollview"
	"github.com/0xdevelop/fltk2go/uikit/slider"
	"github.com/0xdevelop/fltk2go/uikit/splitview"
	"github.com/0xdevelop/fltk2go/uikit/stackview"
	switchview "github.com/0xdevelop/fltk2go/uikit/switch"
	"github.com/0xdevelop/fltk2go/uikit/tableview"
	"github.com/0xdevelop/fltk2go/uikit/tabview"
	"github.com/0xdevelop/fltk2go/uikit/terminalview"
	"github.com/0xdevelop/fltk2go/uikit/textfield"
	"github.com/0xdevelop/fltk2go/uikit/textview"
	"github.com/0xdevelop/fltk2go/uikit/treeview"
	"github.com/0xdevelop/fltk2go/uikit/view"
	"github.com/0xdevelop/fltk2go/uikit/window"
)

type UIView = view.UIView
type Viewable = view.Viewable
type UIWindow = window.UIWindow
type UIGroup = group.UIGroup
type UILabel = label.UILabel
type UIButton = button.UIButton
type ButtonType = button.ButtonType
type Input = input.Input
type InputType = input.InputType
type InputNavigationAction = input.NavigationAction
type UITableView = tableview.TableView
type UITableViewCell = tableview.TableViewCell
type TableViewDataSource = tableview.DataSource
type TableViewDelegate = tableview.Delegate
type UITabView = tabview.UITabView
type UITreeView = treeview.UITreeView
type TreeDataSource = treeview.TreeDataSource
type UIMenuBar = menubar.UIMenuBar
type UIContextMenu = menubar.UIContextMenu
type MenuItem = menubar.MenuItem
type UITextField = textfield.UITextField
type UISlider = slider.UISlider
type UIProgressView = progress.UIProgressView
type UISwitch = switchview.UISwitch
type UIScrollView = scrollview.UIScrollView
type UISplitView = splitview.SplitView
type SplitOrientation = splitview.Orientation
type SplitPositionPolicy = splitview.PositionPolicy
type SplitPositionChange = splitview.PositionChange
type SplitPositionChangeReason = splitview.PositionChangeReason
type SplitGeometry = splitview.SplitGeometry
type UIStackView = stackview.UIStackView
type StackAxis = stackview.Axis
type UITextView = textview.UITextView
type KeyEvent = textview.KeyEvent
type UITerminalView = terminalview.UITerminalView
type TerminalSize = terminalview.Size

const (
	SystemButton   = button.SystemButton
	CheckboxButton = button.CheckboxButton
	RadioButton    = button.RadioButton
	ToggleButton   = button.ToggleButton

	TextInput   = input.TextInput
	IntInput    = input.IntInput
	FloatInput  = input.FloatInput
	SecretInput = input.SecretInput

	InputNavigationSubmit   = input.NavigationSubmit
	InputNavigationNext     = input.NavigationNext
	InputNavigationPrevious = input.NavigationPrevious
	InputNavigationCancel   = input.NavigationCancel

	AxisVertical   = stackview.AxisVertical
	AxisHorizontal = stackview.AxisHorizontal

	SplitHorizontal           = splitview.Horizontal
	SplitVertical             = splitview.Vertical
	SplitPreserveRatio        = splitview.PreserveRatio
	SplitPreserveLeadingSize  = splitview.PreserveLeadingSize
	SplitPreserveTrailingSize = splitview.PreserveTrailingSize
	SplitChangeProgrammatic   = splitview.ChangeProgrammatic
	SplitChangeDrag           = splitview.ChangeDrag
	SplitChangeResize         = splitview.ChangeResize
)

func NewUIWindow(width, height int, title string) *UIWindow {
	return window.NewUIWindow(width, height, title)
}

func NewWindowWithRect(rect *foundation.Rect, title string) *UIWindow {
	return window.NewWindowWithRect(rect, title)
}

func NewUIGroup(r *foundation.Rect) *UIGroup {
	return group.NewUIGroup(r)
}

func NewUILabel(r *foundation.Rect, text string) *UILabel {
	return label.NewUILabel(r, text)
}

func NewUIButton(r *foundation.Rect, title string) *UIButton {
	return button.NewUIButton(r, title)
}

func NewUIButtonWithType(r *foundation.Rect, title string, buttonType ButtonType) *UIButton {
	return button.NewUIButtonWithType(r, title, buttonType)
}

func NewInput(x, y, width, height int, placeholder string) *Input {
	return input.New(x, y, width, height, placeholder)
}

func NewInputWithType(x, y, width, height int, placeholder string, inputType InputType) *Input {
	return input.NewWithType(x, y, width, height, placeholder, inputType)
}

func NewUITableView(x, y, width, height int) (*UITableView, error) {
	return tableview.New(x, y, width, height)
}

func NewUITableViewCell(reuseID string) *UITableViewCell {
	return tableview.NewCell(reuseID)
}

func NewUITabView(r *foundation.Rect) *UITabView {
	return tabview.NewUITabView(r)
}

func NewUITreeView(r *foundation.Rect) *UITreeView {
	return treeview.NewUITreeView(r)
}

func NewUIMenuBar(r *foundation.Rect) *UIMenuBar {
	return menubar.NewUIMenuBar(r)
}

func NewUIContextMenu(r *foundation.Rect) *UIContextMenu {
	return menubar.NewUIContextMenu(r)
}

func NewUITextField(x, y, width, height int, placeholder string) *UITextField {
	return textfield.NewUITextField(x, y, width, height, placeholder)
}

func NewUISlider(r *foundation.Rect) *UISlider {
	return slider.NewUISlider(r)
}

func NewUIProgressView(r *foundation.Rect) *UIProgressView {
	return progress.NewUIProgressView(r)
}

func NewUISwitch(r *foundation.Rect) *UISwitch {
	return switchview.NewUISwitch(r)
}

func NewUIScrollView(r *foundation.Rect) *UIScrollView {
	return scrollview.NewUIScrollView(r)
}

func NewUISplitView(x, y, width, height int, orientation SplitOrientation) *UISplitView {
	return splitview.New(x, y, width, height, orientation)
}

func NewUIStackView(r *foundation.Rect, axis StackAxis) *UIStackView {
	return stackview.NewUIStackView(r, axis)
}

func NewUITextView(r *foundation.Rect) *UITextView {
	return textview.NewUITextView(r)
}

func NewUITerminalView(r *foundation.Rect) *UITerminalView {
	return terminalview.NewUITerminalView(r)
}

func Message(title, message string) {
	dialog.Message(title, message)
}

func Alert(title, message string) {
	dialog.Alert(title, message)
}

func Choice(message string, options ...string) int {
	return dialog.Choice(message, options...)
}

func TitledChoice(title, message string, options ...string) int {
	return dialog.TitledChoice(title, message, options...)
}
