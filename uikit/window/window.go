package window

import (
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/screen"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

type UIWindow struct {
	raw            *fltk_bridge.Window
	root           *view.UIView
	closed         bool
	onClose        func()
	onCloseRequest func() bool
}

func NewUIWindow(width, height int, title string) *UIWindow {
	sSize := screen.GetScreenSize()
	aRect := &foundation.Rect{X: sSize.Width/2 - width/2, Y: sSize.Height/2 - height/2, Width: width, Height: height}

	return NewWindowWithRect(aRect, title)
}

func NewWindowWithRect(rect *foundation.Rect, title string) *UIWindow {
	// 给 nil 一个默认值，避免后面 rect.X 崩溃
	sSize := screen.GetScreenSize()
	if rect == nil {
		rect = &foundation.Rect{X: sSize.Width/2 - screen.DefaultWindowSize.Width/2, Y: sSize.Height/2 - screen.DefaultWindowSize.Height/2, Width: 800, Height: 600}
	}
	if rect.Width <= 0 {
		rect.Width = 800
	}
	if rect.Height <= 0 {
		rect.Height = 600
	}

	win := fltk_bridge.NewWindowWithPosition(rect.X, rect.Y, rect.Width, rect.Height, title)
	win.Resizable(win)

	u := &UIWindow{
		raw:  win,
		root: &view.UIView{},
	}
	// Route the native window-manager close action through the same lifecycle
	// used by application buttons. This gives owners one deterministic callback
	// and guarantees that native widget/automation registrations are released.
	win.SetCallback(func() { u.RequestClose() })

	// root view 不一定需要 raw（它是“逻辑根”），但必须有 host（window）
	u.root.BindHost(win)

	return u
}

func (w *UIWindow) RootView() *view.UIView {
	if w == nil {
		return nil
	}
	return w.root
}

func (w *UIWindow) Show() {
	if w == nil || w.raw == nil {
		return
	}
	w.raw.Show()
}

func (w *UIWindow) SetResizable(resizable bool) {
	if w == nil || w.raw == nil {
		return
	}
	if resizable {
		w.raw.Resizable(w.raw)
	} else {
		w.raw.Resizable(nil)
	}
}

func (w *UIWindow) Raw() *fltk_bridge.Window {
	if w == nil {
		return nil
	}
	return w.raw
}

// OnClose registers a lifecycle callback invoked exactly once when Close is
// first requested, including requests originating from the window manager.
func (w *UIWindow) OnClose(callback func()) {
	if w == nil || w.closed {
		return
	}
	w.onClose = callback
}

// OnCloseRequest registers a policy callback for user-originated close
// requests, such as the window-manager close control or RequestClose. Return
// false to keep the native window open. Owner teardown through Close bypasses
// this policy so application shutdown and post-save cleanup remain explicit.
func (w *UIWindow) OnCloseRequest(callback func() bool) {
	if w == nil || w.closed {
		return
	}
	w.onCloseRequest = callback
}

// RequestClose applies the user-close policy and closes the window only when
// accepted. It reports true when the window is closed after the request.
func (w *UIWindow) RequestClose() bool {
	if w == nil || w.closed {
		return true
	}
	if w.onCloseRequest != nil && !w.onCloseRequest() {
		return false
	}
	w.Close()
	return true
}

// Close hides and destroys the native window. It is safe to call repeatedly.
// The window must be recreated after Close; Raw returns nil once closed.
func (w *UIWindow) Close() {
	if w == nil || w.closed {
		return
	}
	w.closed = true
	raw := w.raw
	w.raw = nil
	callback := w.onClose
	w.onClose = nil
	w.onCloseRequest = nil
	if raw != nil {
		raw.Hide()
		raw.Destroy()
	}
	if callback != nil {
		callback()
	}
}

// IsClosed reports whether the managed close lifecycle has completed.
func (w *UIWindow) IsClosed() bool { return w == nil || w.closed }
