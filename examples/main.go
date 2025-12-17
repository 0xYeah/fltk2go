package main

import (
	_ "embed"
	"runtime"
	"strconv"

	"github.com/0xYeah/fltk2go"
	"github.com/0xYeah/fltk2go/foundation"
	uibutton "github.com/0xYeah/fltk2go/uikit/button"
	uilabel "github.com/0xYeah/fltk2go/uikit/label"
	uiwindow "github.com/0xYeah/fltk2go/uikit/window"
)

const (
	BLUE uint = 0x42A5F500
	GRAY uint = 0x75757500
)

//go:embed resources/imgs/Icon.png
var Icon []byte

func main() {
	// 将当前 goroutine 绑定到当前操作系统线程。
	// 对于 Win32 / OpenGL / GDI+ 等具有线程亲和性的系统或 C/C++ API，这是必须的。
	// 防止 goroutine 被调度到其他 OS 线程，导致 GUI / 图形上下文失效或异常。
	runtime.LockOSThread()

	win := uiwindow.NewUIWindow(&foundation.Rect{X: 50, Y: 50, Width: 600, Height: 400}, "Counter")
	root := win.RootView()

	title := uilabel.NewUILabel(&foundation.Rect{X: 20, Y: 20, Width: 560, Height: 40}, "你点击了 0 次")
	title.SetFontSize(20)
	title.SetTextColor(GRAY)

	btn := uibutton.NewUIButton(&foundation.Rect{X: 20, Y: 80, Width: 160, Height: 44}, "点我 +1")
	btn.SetBackgroundColor(BLUE)

	count := 0
	btn.OnTouchUpInside(func() {
		count++
		title.SetText("你点击了 " + strconv.Itoa(count) + " 次")
	})

	root.AddSubview(title)
	root.AddSubview(btn)

	win.Show()
	fltk2go.Run()
}
