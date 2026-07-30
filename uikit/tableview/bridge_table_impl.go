package tableview

import (
	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

// bridgeTableImpl 实现了 BridgeTable 接口，使用 fltk_bridge.TableRow 作为底层实现
type bridgeTableImpl struct {
	table           *fltk_bridge.TableRow
	drawCellHandler func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int)
	eventHandler    func(row int) bool
}

// newBridgeTableImpl 创建一个新的 bridgeTableImpl 实例
func newBridgeTableImpl(x, y, w, h int) *bridgeTableImpl {
	// 创建 FLTK TableRow 实例
	table := fltk_bridge.NewTableRow(x, y, w, h)

	// 设置初始属性
	table.SetColumnCount(1)                       // 默认1列，可根据需要调整
	table.SetColumnWidthAll(w)                    // 单列默认占满可用宽度，避免示例行挤在左侧
	table.SetRowHeightAll(30)                     // 默认行高30
	table.EnableColumnHeaders()                   // 启用列头
	table.SetColumnHeaderHeight(30)               // 列头高度
	table.SetColor(fltk_bridge.Color(0xFFFFFF00)) // 白色背景

	// 创建 bridgeTableImpl 实例
	bt := &bridgeTableImpl{
		table: table,
	}

	// 设置 FLTK TableRow 的绘制回调
	table.SetDrawCellCallback(func(context fltk_bridge.TableContext, row, col, x, y, w, h int) {
		if bt.drawCellHandler != nil {
			bt.drawCellHandler(context, row, col, x, y, w, h)
		}
	})

	// 设置 FLTK TableRow 的事件回调
	table.SetCallback(func() {
		if bt.eventHandler != nil {
			row := table.CallbackRow()
			bt.eventHandler(row)
		}
	})

	return bt
}

// SetRows 设置表格行数
func (bt *bridgeTableImpl) SetRows(rows int) {
	bt.table.SetRowCount(rows)
}

// Redraw 重绘表格
func (bt *bridgeTableImpl) Redraw() {
	bt.table.Redraw()
}

// SetDrawCellHandler 设置绘制单元格的回调函数
func (bt *bridgeTableImpl) SetDrawCellHandler(fn func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int)) {
	bt.drawCellHandler = fn
}

func (bt *bridgeTableImpl) SetColumnCount(cols int) {
	bt.table.SetColumnCount(cols)
}

func (bt *bridgeTableImpl) SetColumnWidth(col, width int) {
	bt.table.SetColumnWidth(col, width)
}

func (bt *bridgeTableImpl) AllowColumnResizing() {
	bt.table.AllowColumnResizing()
}

func (bt *bridgeTableImpl) EnableColumnHeaders() {
	bt.table.EnableColumnHeaders()
}

func (bt *bridgeTableImpl) SetColumnHeaderHeight(h int) {
	bt.table.SetColumnHeaderHeight(h)
}

func (bt *bridgeTableImpl) SetBackgroundColor(color fltk_bridge.Color) {
	if bt != nil && bt.table != nil {
		bt.table.SetColor(color)
	}
}

// SetEventHandler 设置处理事件的回调函数
func (bt *bridgeTableImpl) SetEventHandler(fn func(row int) bool) {
	bt.eventHandler = fn
}

// Widget 返回底层的 FLTK Widget
func (bt *bridgeTableImpl) Widget() fltk_bridge.Widget {
	return bt.table
}

// ScrollToRow 滚动到指定行
func (bt *bridgeTableImpl) ScrollToRow(row int) {
	bt.table.SetTopRow(row + 1) // +1: FLTK row 0 is the column header
}

// InsertRow 插入一行
func (bt *bridgeTableImpl) InsertRow(row int) {
	// 插入一行的逻辑
	// 注意：TableRow 没有直接的插入行方法，需要重新设置行数
	// 实际应用中可能需要更复杂的逻辑来保存和恢复数据
	currentRows := bt.table.RowCount() - 1 // 减去表头
	if row >= 0 && row <= currentRows {
		bt.table.SetRowCount(currentRows + 2) // +1 用于新行，+1 用于表头
	}
}

// DeleteRow 删除一行
func (bt *bridgeTableImpl) DeleteRow(row int) {
	// 删除一行的逻辑
	// 注意：TableRow 没有直接的删除行方法，需要重新设置行数
	// 实际应用中可能需要更复杂的逻辑来保存和恢复数据
	currentRows := bt.table.RowCount() - 1 // 减去表头
	if row >= 0 && row < currentRows && currentRows > 0 {
		bt.table.SetRowCount(currentRows) // 减少一行，保持表头
	}
}

// SelectRow 选中指定行
func (bt *bridgeTableImpl) SelectRow(row int) {
	// 选中指定行的逻辑
	bt.table.SelectRow(row, fltk_bridge.Select)
}

// GetSelectedRow 获取当前选中的行（0-based 数据行索引，无选中返回 -1）
func (bt *bridgeTableImpl) GetSelectedRow() int {
	top, _, _, _ := bt.table.Selection()
	if top <= 0 { // -1 = no selection, 0 = column header row
		return -1
	}
	return top - 1 // convert from FLTK 1-based data row to 0-based
}
