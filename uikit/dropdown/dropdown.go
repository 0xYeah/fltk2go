package dropdown

import (
	"fmt"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

type UIDropdown struct {
	v                  view.UIView
	raw                *fltk_bridge.Choice
	options            []string
	onSelectionChanged func(index int, option string)
}

func NewUIDropdown(r *foundation.Rect) *UIDropdown {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 120, Height: 30}
	}

	choice := fltk_bridge.NewChoice(r.X, r.Y, r.Width, r.Height, "")

	dd := &UIDropdown{
		raw: choice,
	}
	dd.v.BindRaw(choice)
	dd.v.SetAutomationRole("combobox")
	dd.v.SetAutomationTextHandlers(func(option string) error {
		for index, candidate := range dd.options {
			if candidate == option {
				dd.selectOption(index)
				return nil
			}
		}
		return fmt.Errorf("dropdown option %q not found", option)
	}, nil)
	dd.v.SetAutomationValueHandler(func() (string, bool) { return dd.SelectedOption(), true })

	return dd
}

func (dd *UIDropdown) View() *view.UIView { return &dd.v }

func (dd *UIDropdown) Raw() *fltk_bridge.Choice { return dd.raw }

func (dd *UIDropdown) SetOptions(options []string) {
	dd.raw.Clear()
	dd.options = append(dd.options[:0], options...)
	for index, opt := range dd.options {
		index, opt := index, opt
		// Fl_Choice dispatches the selected menu item's callback instead of
		// the parent widget callback when items have callbacks. Wire each item
		// to the dropdown-level handler so real native popup selections notify
		// consumers consistently.
		dd.raw.Add(opt, func() { dd.selectOption(index) })
	}
	if len(options) > 0 {
		dd.raw.SetValue(0)
	}
}

func (dd *UIDropdown) Options() []string {
	return dd.options
}

func (dd *UIDropdown) SelectedIndex() int {
	return dd.raw.Value()
}

func (dd *UIDropdown) SelectedOption() string {
	idx := dd.SelectedIndex()
	if idx >= 0 && idx < len(dd.options) {
		return dd.options[idx]
	}
	return ""
}

func (dd *UIDropdown) SetSelectedIndex(index int) {
	if index >= 0 && index < len(dd.options) {
		dd.raw.SetValue(index)
	}
}

func (dd *UIDropdown) OnSelectionChanged(cb func(index int, option string)) {
	dd.onSelectionChanged = cb
}

func (dd *UIDropdown) selectOption(index int) {
	if index < 0 || index >= len(dd.options) {
		return
	}
	dd.raw.SetValue(index)
	if dd.onSelectionChanged != nil {
		dd.onSelectionChanged(index, dd.options[index])
	}
}

// On 绑定事件
func (dd *UIDropdown) On(event fltk_bridge.Event, handler func(fltk_bridge.Event) bool) {
	dd.v.On(event, handler)
}
