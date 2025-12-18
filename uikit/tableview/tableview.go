package tableview

type TableView struct {
	table BridgeTable

	dataSource DataSource
	delegate   Delegate

	defaultRowHeight int

	// 复用池：reuseID -> cells
	reusePool map[string][]*TableViewCell
	// 可见缓存（可选）：row -> cell
	visible map[int]*TableViewCell
}

func New(x, y, w, h int) (*TableView, error) {
	bt, err := newBridgeTable(x, y, w, h)
	if err != nil {
		return nil, err
	}

	tv := &TableView{
		table:            bt,
		defaultRowHeight: 24,
		reusePool:        map[string][]*TableViewCell{},
		visible:          map[int]*TableViewCell{},
	}

	// 绑定底层回调
	tv.table.SetDrawCellHandler(tv.onDrawCell)
	tv.table.SetEventHandler(tv.onEvent)

	return tv, nil
}

func (tv *TableView) SetDataSource(ds DataSource) { tv.dataSource = ds }
func (tv *TableView) SetDelegate(d Delegate)      { tv.delegate = d }

func (tv *TableView) SetDefaultRowHeight(h int) {
	if h > 0 {
		tv.defaultRowHeight = h
	}
}

func (tv *TableView) Dequeue(reuseID string) *TableViewCell {
	list := tv.reusePool[reuseID]
	if n := len(list); n > 0 {
		c := list[n-1]
		tv.reusePool[reuseID] = list[:n-1]
		c.PrepareForReuse()
		return c
	}
	return NewCell(reuseID)
}

func (tv *TableView) Enqueue(c *TableViewCell) {
	if c == nil || c.ReuseID == "" {
		return
	}
	tv.reusePool[c.ReuseID] = append(tv.reusePool[c.ReuseID], c)
}

func (tv *TableView) ReloadData() {
	if tv.dataSource == nil {
		tv.table.SetRows(0)
		tv.table.Redraw()
		return
	}

	// 回收旧可见 cell
	for row, cell := range tv.visible {
		_ = row
		tv.Enqueue(cell)
	}
	tv.visible = map[int]*TableViewCell{}

	rows := tv.dataSource.NumberOfRows(tv)
	if rows < 0 {
		rows = 0
	}
	tv.table.SetRows(rows)
	tv.table.Redraw()
}

// ============ callbacks ============

// onDrawCell：每行绘制时取 cell 并交给业务层配置
func (tv *TableView) onDrawCell(row int, x, y, w, h int) {
	if tv.dataSource == nil {
		return
	}

	cell, ok := tv.visible[row]
	if !ok {
		cell = tv.dataSource.CellForRow(tv, row)
		if cell == nil {
			return
		}
		cell.row = row
		tv.visible[row] = cell
	}

	// 这里是“绘制型 cell”的默认策略：TableViewCell 自身不画，
	// 交由你桥接层/绘图 API 在 dataSource 的 cell 内完成（比如设置文本、颜色等）
	//
	// 如果你想做“UIKit 风格 cell.layoutSubviews + 子控件”，
	// 后续可以在这里把 cell 的 ContentView frame 设置为 (x,y,w,h) 并 show/hide。
	_ = x
	_ = y
	_ = w
	_ = h
}

func (tv *TableView) onEvent(row int) bool {
	// 点击选中之类事件：交给 delegate
	if tv.delegate != nil && row >= 0 {
		tv.delegate.DidSelectRow(tv, row)
		return true
	}
	return false
}
