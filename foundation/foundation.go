package foundation

// Rect：UI 通用几何类型（int 像素），UIKit 风格用它拼 UI。
// 不依赖任何 UI 包，所有控件都能用。
type Rect struct {
	X      int
	Y      int
	Width  int
	Height int
}

// New：便捷构造（类似 CGRectMake）
func NewRect(x, y, width, height int) Rect {
	return Rect{X: x, Y: y, Width: width, Height: height}
}
