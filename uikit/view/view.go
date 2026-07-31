package view

import "github.com/0xdevelop/fltk2go/fltk_bridge"

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
	raw                   fltk_bridge.Widget
	host                  Container
	eventHandlers         map[fltk_bridge.Event]func(fltk_bridge.Event) bool
	automation            automationState
	nextSuperviewObserver uint64
	superviewObservers    map[uint64]func(Container, Container)
}

func (v *UIView) View() *UIView {
	if v == nil {
		return nil
	}
	return v
}

// BindHost：框架内部使用，为 view 绑定父容器
func (v *UIView) BindHost(host Container) {
	if v == nil || v.host == host {
		return
	}
	previous := v.host
	v.host = host
	observers := make([]func(Container, Container), 0, len(v.superviewObservers))
	for _, observer := range v.superviewObservers {
		observers = append(observers, observer)
	}
	for _, observer := range observers {
		observer(previous, host)
	}
}

// ObserveSuperviewChanges observes logical UIKit ownership changes. The
// returned function unregisters the observer.
func (v *UIView) ObserveSuperviewChanges(observer func(Container, Container)) func() {
	if v == nil || observer == nil {
		return func() {}
	}
	v.nextSuperviewObserver++
	id := v.nextSuperviewObserver
	if v.superviewObservers == nil {
		v.superviewObservers = map[uint64]func(Container, Container){}
	}
	v.superviewObservers[id] = observer
	return func() {
		if v != nil && v.superviewObservers != nil {
			delete(v.superviewObservers, id)
		}
	}
}

// BindRaw：框架内部使用，为 view 绑定底层 widget
func (v *UIView) BindRaw(raw fltk_bridge.Widget) {
	if v == nil || v.raw == raw {
		return
	}
	v.raw = raw
	if observable, ok := raw.(interface{ OnDelete(func()) }); ok {
		boundRaw := raw
		observable.OnDelete(func() {
			if v.raw != boundRaw {
				return
			}
			v.BindHost(nil)
			v.clearAutomationOnDelete()
			v.eventHandlers = nil
			v.superviewObservers = nil
			v.raw = nil
		})
	}
	if v.eventHandlers != nil {
		if eh, ok := v.raw.(interface {
			SetEventHandler(func(fltk_bridge.Event) bool)
		}); ok {
			eh.SetEventHandler(v.handleEvent)
		}
	}
}

func (v *UIView) Raw() fltk_bridge.Widget {
	if v == nil {
		return nil
	}
	return v.raw
}

// Superview returns the container currently hosting this view, if any.
func (v *UIView) Superview() Container {
	if v == nil {
		return nil
	}
	return v.host
}

// On 绑定闭包事件流
func (v *UIView) On(event fltk_bridge.Event, handler func(fltk_bridge.Event) bool) {
	if v == nil {
		return
	}
	if v.eventHandlers == nil {
		v.eventHandlers = make(map[fltk_bridge.Event]func(fltk_bridge.Event) bool)
		if v.raw != nil {
			if eh, ok := v.raw.(interface {
				SetEventHandler(func(fltk_bridge.Event) bool)
			}); ok {
				eh.SetEventHandler(v.handleEvent)
			}
		}
	}
	v.eventHandlers[event] = handler
}

func (v *UIView) handleEvent(e fltk_bridge.Event) bool {
	if v == nil {
		return false
	}
	if v.eventHandlers != nil {
		if handler, ok := v.eventHandlers[e]; ok {
			return handler(e)
		}
	}
	return false
}

// AddSubview：iOS 语义。如果当前 view 本身是容器，则将 child 添加到当前 view 中；
// 否则，退回到将 child 添加到当前 view 的父容器中（虽然不推荐，但兼容旧逻辑）。
func (v *UIView) AddSubview(child Viewable) {
	if v == nil || child == nil {
		return
	}
	cv := child.View()
	if cv == nil || cv.raw == nil {
		return
	}

	// 如果 v.raw 本身实现了 Container (比如 fltk_bridge.Group)
	if container, ok := v.raw.(Container); ok {
		container.Add(cv.raw)
		cv.BindHost(container)
		v.AddAutomationChild(child)
	} else if v.host != nil {
		// 回退逻辑，添加给 host
		v.host.Add(cv.raw)
		cv.BindHost(v.host)
		v.AddAutomationChild(child)
	}
}

// RemoveFromSuperview removes the view's raw widget from its current host.
func (v *UIView) RemoveFromSuperview() {
	if v == nil || v.host == nil || v.raw == nil {
		return
	}
	host := v.host
	host.Remove(v.raw)
	v.BindHost(nil)
}
