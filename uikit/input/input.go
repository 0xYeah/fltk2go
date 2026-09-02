package input

import (
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

// Input 输入框组件
type Input struct {
	// 底层FLTK输入框
	raw fltk_bridge.Widget

	// 基础视图
	v view.UIView

	onChange func()

	onNavigation func(NavigationAction) bool
}

// NavigationAction identifies list-navigation commands commonly owned by a
// search or launcher input. Ordinary editing keys remain with the native input.
type NavigationAction int

const (
	NavigationSubmit NavigationAction = iota + 1
	NavigationNext
	NavigationPrevious
	NavigationCancel
)

// InputType 输入框类型
type InputType int

const (
	// TextInput 文本输入框
	TextInput InputType = iota
	// IntInput 整数输入框
	IntInput
	// FloatInput 浮点数输入框
	FloatInput
	// SecretInput 密码输入框
	SecretInput
)

// New 创建一个新的输入框
func New(x, y, width, height int, placeholder string) *Input {
	return NewWithType(x, y, width, height, placeholder, TextInput)
}

// NewWithType 创建一个指定类型的输入框
func NewWithType(x, y, width, height int, placeholder string, inputType InputType) *Input {
	var input fltk_bridge.Widget

	switch inputType {
	case IntInput:
		input = fltk_bridge.NewIntInput(x, y, width, height, placeholder)
	case FloatInput:
		input = fltk_bridge.NewFloatInput(x, y, width, height, placeholder)
	case SecretInput:
		input = fltk_bridge.NewSecretInput(x, y, width, height, placeholder)
	default:
		input = fltk_bridge.NewInput(x, y, width, height, placeholder)
	}

	in := &Input{
		raw: input,
	}

	// 绑定底层widget到view
	in.v.BindRaw(input)
	in.v.SetAutomationRole("textbox").SetAutomationName(placeholder)
	secure := inputType == SecretInput
	if secure {
		in.v.SetAutomationProperty("secure", "true")
	}
	in.v.SetAutomationTextHandlers(func(text string) error {
		in.SetText(text)
		if in.onChange != nil {
			in.onChange()
		}
		return nil
	}, func() (string, bool) {
		// Automation may write credentials for an end-to-end login test, but
		// snapshots must never serialize them. Callers that own the widget can
		// still read Text() explicitly on the GUI thread.
		if secure {
			return "", false
		}
		return in.Text(), true
	})

	return in
}

// SetText 设置输入框文本
func (in *Input) SetText(text string) {
	if in != nil && in.raw != nil {
		if input, ok := in.raw.(interface{ SetValue(value string) bool }); ok {
			input.SetValue(text)
		}
	}
}

// Text 获取输入框文本
func (in *Input) Text() string {
	if in != nil && in.raw != nil {
		if input, ok := in.raw.(interface{ Value() string }); ok {
			return input.Value()
		}
	}
	return ""
}

// SetPlaceholder 设置占位文本
func (in *Input) SetPlaceholder(placeholder string) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetLabel(label string) }); ok {
			widget.SetLabel(placeholder)
		}
	}
}

// Placeholder 获取占位文本
func (in *Input) Placeholder() string {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ Label() string }); ok {
			return widget.Label()
		}
	}
	return ""
}

// SetFontSize 设置字体大小
func (in *Input) SetFontSize(size int) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetTextSize(size int) }); ok {
			widget.SetTextSize(size)
		}
	}
}

// SetFont 设置字体
func (in *Input) SetFont(font fltk_bridge.Font) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetTextFont(font fltk_bridge.Font) }); ok {
			widget.SetTextFont(font)
		}
	}
}

// SetTextColor 设置文本颜色
func (in *Input) SetTextColor(color uint) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetTextColor(col fltk_bridge.Color) }); ok {
			widget.SetTextColor(fltk_bridge.Color(color))
		}
	}
}

// SetCursorColor styles the native insertion caret independently from text and
// background colors so dark and high-contrast input themes remain usable.
func (in *Input) SetCursorColor(color uint) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetCursorColor(col fltk_bridge.Color) }); ok {
			widget.SetCursorColor(fltk_bridge.Color(color))
		}
	}
}

// SetBackgroundColor 设置背景颜色
func (in *Input) SetBackgroundColor(color uint) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ SetColor(c fltk_bridge.Color) }); ok {
			widget.SetColor(fltk_bridge.Color(color))
		}
	}
}

// SetEnabled 设置是否可用
func (in *Input) SetEnabled(enabled bool) {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ Activate() }); ok {
			if enabled {
				widget.Activate()
			} else {
				if widget, ok := in.raw.(interface{ Deactivate() }); ok {
					widget.Deactivate()
				}
			}
		}
	}
}

// IsEnabled 获取是否可用
func (in *Input) IsEnabled() bool {
	if in != nil && in.raw != nil {
		if widget, ok := in.raw.(interface{ IsActive() bool }); ok {
			return widget.IsActive()
		}
	}
	return false
}

// OnChange 设置文本变化回调
func (in *Input) OnChange(callback func()) {
	if in != nil && in.raw != nil {
		in.onChange = callback
		if widget, ok := in.raw.(interface{ SetCallback(f func()) }); ok {
			widget.SetCallback(callback)
		}
	}
}

// OnNavigation routes Enter, Down, Up and Escape through one native-input
// callback. Return true when the owner handled the command, or false to let the
// underlying input retain its normal behavior.
func (in *Input) OnNavigation(callback func(NavigationAction) bool) {
	if in == nil {
		return
	}
	in.onNavigation = callback
	in.v.On(fltk_bridge.KEYDOWN, func(fltk_bridge.Event) bool {
		return in.dispatchNavigation(fltk_bridge.EventKey())
	})
}

func (in *Input) dispatchNavigation(key int) bool {
	if in == nil || in.onNavigation == nil {
		return false
	}
	var action NavigationAction
	switch key {
	case fltk_bridge.ENTER_KEY:
		action = NavigationSubmit
	case fltk_bridge.DOWN:
		action = NavigationNext
	case fltk_bridge.UP:
		action = NavigationPrevious
	case fltk_bridge.ESCAPE:
		action = NavigationCancel
	default:
		return false
	}
	return in.onNavigation(action)
}

// View 返回基础视图，实现view.Viewable接口
func (in *Input) View() *view.UIView {
	if in == nil {
		return nil
	}
	return &in.v
}

// Raw 返回底层FLTK输入框
func (in *Input) Raw() fltk_bridge.Widget {
	if in == nil {
		return nil
	}
	return in.raw
}
