package fltk2go

import (
	"runtime"

	"github.com/0xYeah/fltk2go/fltk_bridge"
)

func App(build func()) {
	runtime.LockOSThread()
	if build != nil {
		build()
	}
	fltk_bridge.Run()
}
