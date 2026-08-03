package tableview

import "github.com/0xdevelop/fltk2go/fltk_bridge"

// TableInteraction describes a native table selection callback. FLTK reports
// Clicks=1 for a double click (the number of clicks after the first one).
type TableInteraction struct {
	Row    int
	Clicks int
}

// BridgeTable is the minimal interface TableView needs from the underlying FLTK table.
type BridgeTable interface {
	SetRows(rows int)
	Redraw()
	SetDrawCellHandler(fn func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int))
	SetEventHandler(fn func(TableInteraction) bool)
	GetSelectedRow() int
	SelectRow(row int)
	ScrollToRow(row int)
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
