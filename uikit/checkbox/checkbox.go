package checkbox

import (
	"strconv"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

type CheckboxStyle struct {
	Font      fltk_bridge.Font
	FontSize  int
	TextColor uint
	Color     uint
}

func DefaultCheckboxStyle() CheckboxStyle {
	return CheckboxStyle{
		Font:      fltk_bridge.HELVETICA,
		FontSize:  14,
		TextColor: 0,
		Color:     0,
	}
}

type UICheckbox struct {
	v   view.UIView
	raw *fltk_bridge.CheckButton

	style          CheckboxStyle
	onValueChanged func(bool)
}

func NewUICheckbox(r *foundation.Rect, title string) *UICheckbox {
	return NewUICheckboxWithOptions(r, title, DefaultCheckboxStyle())
}

func NewUICheckboxWithOptions(r *foundation.Rect, title string, style CheckboxStyle) *UICheckbox {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 120, Height: 30}
	}

	cb := fltk_bridge.NewCheckButton(r.X, r.Y, r.Width, r.Height, title)

	checkbox := &UICheckbox{
		raw:   cb,
		style: style,
	}

	checkbox.v.BindRaw(cb)
	checkbox.v.SetAutomationRole("checkbox")
	checkbox.v.SetAutomationValueHandler(func() (string, bool) {
		return strconv.FormatBool(checkbox.Value()), true
	})
	checkbox.v.OnAutomationClick(func() error {
		checkbox.SetValue(!checkbox.Value())
		checkbox.notifyValueChanged()
		return nil
	})
	checkbox.ApplyStyle(style)

	cb.SetCallback(func() {
		checkbox.notifyValueChanged()
	})

	return checkbox
}

func (c *UICheckbox) View() *view.UIView { return &c.v }

func (c *UICheckbox) Raw() *fltk_bridge.CheckButton { return c.raw }

func (c *UICheckbox) ApplyStyle(style CheckboxStyle) {
	c.style = style
	c.raw.SetLabelFont(style.Font)
	c.raw.SetLabelSize(style.FontSize)
	c.raw.SetLabelColor(fltk_bridge.Color(style.TextColor))
	if style.Color != 0 {
		c.raw.SetColor(fltk_bridge.Color(style.Color))
	}
	c.raw.Redraw()
}

func (c *UICheckbox) Style() CheckboxStyle { return c.style }

func (c *UICheckbox) SetValue(val bool) {
	c.raw.SetValue(val)
}

func (c *UICheckbox) Value() bool {
	return c.raw.Value()
}

func (c *UICheckbox) OnValueChanged(cb func(bool)) {
	c.onValueChanged = cb
}

func (c *UICheckbox) notifyValueChanged() {
	if c != nil && c.onValueChanged != nil {
		c.onValueChanged(c.Value())
	}
}
