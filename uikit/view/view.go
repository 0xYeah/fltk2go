package view

import "github.com/0xYeah/fltk2go/fltk_bridge"

// Container：任何能容纳子控件的对象（Window/Group）
// 你 fltk_bridge.Window 继承了 Group，天然满足 Add/Remove
type Container interface {
	Add(w fltk_bridge.Widget)
	Remove(w fltk_bridge.Widget)
}

// Viewable：UIKit 风格，“能提供 UIView 的对象”
type Viewable interface {
	View() *UIView
}

// UIView：UIKit 风格的基础 View
// raw：底层 FLTK widget（Box/Button/Group/Window 都是 Widget）
// host：父容器（Window 或 Group）
type UIView struct {
	raw  fltk_bridge.Widget
	host Container
}

// BindHost：框架内部使用，为 view 绑定父容器
func (v *UIView) BindHost(host Container) {
	v.host = host
}

// BindRaw：框架内部使用，为 view 绑定底层 widget
func (v *UIView) BindRaw(raw fltk_bridge.Widget) {
	v.raw = raw
}

func (v *UIView) Raw() fltk_bridge.Widget { return v.raw }

// AddSubview：iOS 语义。核心就一件事：host.Add(child.raw)
func (v *UIView) AddSubview(child Viewable) {
	if v == nil || v.host == nil || child == nil {
		return
	}
	cv := child.View()
	if cv == nil || cv.raw == nil {
		return
	}
	v.host.Add(cv.raw)
}
