package window

import (
	"github.com/0xYeah/fltk2go/fltk_bridge"
	"github.com/0xYeah/fltk2go/foundation"
	"github.com/0xYeah/fltk2go/uikit/view"
)

type UIWindow struct {
	raw  *fltk_bridge.Window
	root *view.UIView
}

func NewUIWindow(rect *foundation.Rect, title string) *UIWindow {
	// 给 nil 一个默认值，避免后面 rect.X 崩溃
	if rect == nil {
		rect = &foundation.Rect{X: 100, Y: 100, Width: 800, Height: 600}
	}
	if rect.Width <= 0 {
		rect.Width = 800
	}
	if rect.Height <= 0 {
		rect.Height = 600
	}

	win := fltk_bridge.NewWindowWithPosition(rect.X, rect.Y, rect.Width, rect.Height, title)

	u := &UIWindow{
		raw:  win,
		root: &view.UIView{},
	}

	// root view 不一定需要 raw（它是“逻辑根”），但必须有 host（window）
	u.root.BindHost(win)

	return u
}

func (w *UIWindow) RootView() *view.UIView { return w.root }

func (w *UIWindow) Show() {
	if w == nil || w.raw == nil {
		return
	}
	w.raw.Show()
}

func (w *UIWindow) Raw() *fltk_bridge.Window { return w.raw }
