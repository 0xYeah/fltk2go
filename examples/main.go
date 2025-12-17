package main

import (
	"github.com/0xYeah/fltk2go"
	"github.com/0xYeah/fltk2go/core"
	"github.com/0xYeah/fltk2go/window"
)

// FLTK uses an RGBI color representation, the I is an index into FLTK's color map
// Passing 00 as I will use the RGB part of the value
const GRAY = 0x75757500
const LIGHT_GRAY = 0xeeeeee00
const BLUE = 0x42A5F500
const SEL_BLUE = 0x2196F300
const WIDTH = 600
const HEIGHT = 400

func main() {
	// 创建窗口
	w := window.NewWindow(core.Frame{0.0,0,1024,768} "Hello fltk2go")
	w.CenterScreen()

	// 创建按钮
	btn := button.New(20, 20, 200, 40, "Hello World")
	btn.SetCallback(func() {
		fltk2go.Message("Hello, FLTK + Go 👋")
	})

	w.Add(btn)
	w.Show()

	// 进入事件循环
	fltk2go.Run()
}
