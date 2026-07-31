package splitview

import (
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/button"
	"github.com/0xdevelop/fltk2go/uikit/label"
	"github.com/0xdevelop/fltk2go/uikit/splitview"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

const (
	BLUE   uint = 0x42A5F500
	GRAY   uint = 0x75757500
	ORANGE uint = 0xFFA72600
	GREEN  uint = 0x4CAF5000
	RED    uint = 0xF4433600
	WHITE  uint = 0xFFFFFFFF
)

func BuildView(parent *view.UIView) view.Viewable {

	title := label.NewUILabel(&foundation.Rect{X: 50, Y: 20, Width: 900, Height: 40}, "Split View Example")
	title.SetFontSize(24)
	title.SetFont(fltk_bridge.HELVETICA_BOLD)
	title.SetAlignment(fltk_bridge.ALIGN_CENTER)
	parent.AddSubview(title)

	hSplit := splitview.New(50, 80, 900, 500, splitview.Horizontal)

	leftPanel := label.NewUILabel(&foundation.Rect{X: 0, Y: 0, Width: 300, Height: 500}, "Left Panel")
	leftPanel.SetFontSize(18)
	leftPanel.SetFont(fltk_bridge.HELVETICA_BOLD)
	leftPanel.SetAlignment(fltk_bridge.ALIGN_CENTER)
	leftPanel.SetBackgroundColor(0xE3F2FD00)
	leftPanel.SetFrame(fltk_bridge.ENGRAVED_BOX)

	rightPanel := label.NewUILabel(&foundation.Rect{X: 0, Y: 0, Width: 600, Height: 500}, "Right Panel")
	rightPanel.SetFontSize(18)
	rightPanel.SetFont(fltk_bridge.HELVETICA_BOLD)
	rightPanel.SetAlignment(fltk_bridge.ALIGN_CENTER)
	rightPanel.SetBackgroundColor(0xE8F5E800)
	rightPanel.SetFrame(fltk_bridge.ENGRAVED_BOX)

	hSplit.SetLeftView(leftPanel)
	hSplit.SetRightView(rightPanel)
	hSplit.SetMinimumSizes(220, 320)
	hSplit.SetDividerSize(8)
	hSplit.SetDividerColors(0xCBD5E100, 0x94A3B800, 0x64748B00)
	hSplit.SetPosition(0.35)
	hSplit.SetAutomationID("example.workspace")
	parent.AddSubview(hSplit)

	vSplit := splitview.New(50, 600, 900, 80, splitview.Vertical)

	topPanel := button.NewUIButton(&foundation.Rect{X: 0, Y: 0, Width: 450, Height: 80}, "Top Button")
	topPanel.SetBackgroundColor(BLUE)
	topPanel.SetTitleColor(WHITE)

	bottomPanel := button.NewUIButton(&foundation.Rect{X: 0, Y: 0, Width: 450, Height: 80}, "Bottom Button")
	bottomPanel.SetBackgroundColor(GREEN)
	bottomPanel.SetTitleColor(WHITE)

	vSplit.SetLeftView(topPanel)
	vSplit.SetRightView(bottomPanel)
	vSplit.SetMinimumSizes(20, 20)
	vSplit.SetDividerSize(6)
	vSplit.SetPosition(0.5)
	parent.AddSubview(vSplit)

	topPanel.OnTouchUpInside(func() {
		leftPanel.SetText("Left Panel - Top Button Clicked!")
	})
	bottomPanel.OnTouchUpInside(func() {
		rightPanel.SetText("Right Panel - Bottom Button Clicked!")
	})

	return nil
}
