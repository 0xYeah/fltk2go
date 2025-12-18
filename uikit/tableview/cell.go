package tableview

// TableViewCell：UITableViewCell 的最小抽象
type TableViewCell struct {
	ReuseID string

	// 你未来若要“控件型 cell”（cell 里放 label/button），
	// 可以在这里挂一个 uikit/view 的容器，如 ContentView *view.View
	// 但为了不强依赖你当前 uikit/view 的具体 API，这里先留空。
	row int
}

func NewCell(reuseID string) *TableViewCell {
	return &TableViewCell{ReuseID: reuseID}
}

// PrepareForReuse：复用前清理状态
func (c *TableViewCell) PrepareForReuse() {
	// TODO: 你需要的清理逻辑（文本、颜色、回调解绑等）
}

func (c *TableViewCell) Row() int { return c.row }
