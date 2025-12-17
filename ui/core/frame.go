package core

import "github.com/0xYeah/fltk2go/fltk_bridge"

type Frame struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

func InitStyles() {
	fltk_bridge.InitStyles()
}
