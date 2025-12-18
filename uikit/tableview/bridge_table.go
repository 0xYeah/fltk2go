package tableview

// BridgeTable 是对 fltk_bridge “Table/RowTable” 的最小需求抽象
// 你只需要在 newBridgeTable(...) 里用你项目现有 fltk_bridge/table.go 实现它。
type BridgeTable interface {
	SetRows(rows int)
	Redraw()

	// 回调：绘制与事件
	SetDrawCellHandler(fn func(row int, x, y, w, h int))
	SetEventHandler(fn func(row int) bool) // 返回 true 表示已处理（例如 click 选中）
}

// newBridgeTable：在这里把 fltk_bridge.Table 适配成 BridgeTable
func newBridgeTable(x, y, w, h int) (BridgeTable, error) {
	// TODO(你来对齐): 用你现有 fltk_bridge/table.go 创建 table 实例并返回
	//
	// 例如（伪代码）：
	//   t := fltk_bridge.NewTable(x,y,w,h)
	//   return &bridgeTableImpl{t: t}, nil
	//
	return nil, ErrBridgeNotImplemented
}

var ErrBridgeNotImplemented = &bridgeErr{"bridge_table.go: newBridgeTable() not implemented"}

type bridgeErr struct{ s string }

func (e *bridgeErr) Error() string { return e.s }
