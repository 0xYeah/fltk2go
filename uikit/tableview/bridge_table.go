package tableview

import "github.com/0xdevelop/fltk2go/fltk_bridge"

// BridgeTable is the minimal interface TableView needs from the underlying FLTK table.
type BridgeTable interface {
	SetRows(rows int)
	Redraw()
	SetDrawCellHandler(fn func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int))
	SetEventHandler(fn func(row int) bool)
	GetSelectedRow() int
	Widget() fltk_bridge.Widget

	SetColumnCount(cols int)
	SetColumnWidth(col, width int)
	AllowColumnResizing()
	EnableColumnHeaders()
	SetColumnHeaderHeight(h int)
	SetBackgroundColor(color fltk_bridge.Color)
}

func newBridgeTable(x, y, w, h int) (BridgeTable, error) {
	return newBridgeTableImpl(x, y, w, h), nil
}
