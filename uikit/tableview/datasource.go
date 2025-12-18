package tableview

// DataSource：提供行数与 Cell
type DataSource interface {
	NumberOfRows(tv *TableView) int
	CellForRow(tv *TableView, row int) *TableViewCell
}

// Delegate：交互与布局（可选实现）
type Delegate interface {
	DidSelectRow(tv *TableView, row int)
	RowHeight(tv *TableView, row int) int // 返回 0 表示使用默认高度
}
