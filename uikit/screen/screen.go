package screen

import (
	"sync"

	"github.com/kbinani/screenshot"
)

var (
	once              sync.Once
	DefaultWindowSize = ScreenSize{
		Width:  1440,
		Height: 900,
	}
	screenSizeInstance *ScreenSize
)

type ScreenSize struct {
	Width  int
	Height int
}

func GetScreenSize() *ScreenSize {
	once.Do(func() {
		bounds := screenshot.GetDisplayBounds(0)
		screenSizeInstance = &ScreenSize{
			Width:  bounds.Dx(),
			Height: bounds.Dy(),
		}
	})
	return screenSizeInstance
}
