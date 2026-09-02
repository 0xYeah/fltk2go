package main

import (
	"fmt"
	"runtime"

	"github.com/0xdevelop/fltk2go"
	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit"
)

func main() {
	runtime.LockOSThread()
	win := uikit.NewUIWindow(760, 460, "Closable TabView Example")
	root := win.RootView()

	title := uikit.NewUILabel(&foundation.Rect{X: 32, Y: 24, Width: 696, Height: 32}, "Owner-controlled closable tabs")
	title.SetFontSize(22)
	root.AddSubview(title)

	status := uikit.NewUILabel(&foundation.Rect{X: 32, Y: 66, Width: 392, Height: 28}, "Select a tab, then reorder or close it.")
	status.SetFrame(fltk_bridge.FLAT_BOX)
	status.SetBackgroundColor(uint(fltk_bridge.BACKGROUND_COLOR))
	status.View().SetAutomationID("closable-tabs.status")
	root.AddSubview(status)

	tabs := uikit.NewUITabView(&foundation.Rect{X: 32, Y: 110, Width: 696, Height: 300})
	tabs.SetAutomationID("closable-tabs")
	for index, name := range []string{"Local Shell", "Production", "Logs"} {
		panel := uikit.NewUIGroup(&foundation.Rect{})
		message := uikit.NewUILabel(&foundation.Rect{X: 52, Y: 190, Width: 656, Height: 36}, fmt.Sprintf("%s content", name))
		message.SetFontSize(18)
		panel.AddSubview(message)
		tabs.AddTabWithID(fmt.Sprintf("session-%d", index+1), name, panel)
	}
	tabs.SetTabsClosable(true)
	tabs.OnTabCloseRequested(func(index int) {
		id := tabs.TabID(index)
		if tabs.RemoveTab(index) {
			status.SetText(fmt.Sprintf("Closed %s · %d tab(s) remain", id, tabs.Count()))
		}
	})
	root.AddSubview(tabs)

	moveLeft := uikit.NewUIButton(&foundation.Rect{X: 440, Y: 64, Width: 136, Height: 32}, "Move left")
	moveLeft.View().SetAutomationID("closable-tabs.move-left")
	moveLeft.OnTouchUpInside(func() {
		from := tabs.ActiveIndex()
		if tabs.MoveTab(from, from-1) {
			status.SetText(fmt.Sprintf("Moved %s left", tabs.TabID(tabs.ActiveIndex())))
		}
	})
	root.AddSubview(moveLeft)

	moveRight := uikit.NewUIButton(&foundation.Rect{X: 592, Y: 64, Width: 136, Height: 32}, "Move right")
	moveRight.View().SetAutomationID("closable-tabs.move-right")
	moveRight.OnTouchUpInside(func() {
		from := tabs.ActiveIndex()
		if tabs.MoveTab(from, from+1) {
			status.SetText(fmt.Sprintf("Moved %s right", tabs.TabID(tabs.ActiveIndex())))
		}
	})
	root.AddSubview(moveRight)

	win.Show()
	fltk2go.Run()
}
