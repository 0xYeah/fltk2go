package button

import (
	"github.com/0xYeah/fltk2go/fltk_bridge"
	"github.com/0xYeah/fltk2go/foundation"
	"github.com/0xYeah/fltk2go/uikit/view"
)

type UIButton struct {
	v   view.UIView
	raw *fltk_bridge.Button
}

func NewUIButton(r *foundation.Rect, title string) *UIButton {
	if r == nil {
		r = &foundation.Rect{X: 0, Y: 0, Width: 120, Height: 36}
	}

	btn := fltk_bridge.NewButton(r.X, r.Y, r.Width, r.Height, title)

	b := &UIButton{raw: btn}
	b.v.BindRaw(btn)
	return b
}

func (b *UIButton) View() *view.UIView { return &b.v }

func (b *UIButton) SetTitle(s string) {
	if b.raw != nil {
		b.raw.SetLabel(s)
	}
}

func (b *UIButton) SetBackgroundColor(rgb uint) {
	if b.raw != nil {
		b.raw.SetColor(fltk_bridge.Color(rgb))
	}
}

func (b *UIButton) OnTouchUpInside(cb func()) {
	if b.raw != nil {
		b.raw.SetCallback(cb)
	}
}

func (b *UIButton) Raw() *fltk_bridge.Button { return b.raw }
