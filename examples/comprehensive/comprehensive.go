package comprehensive

import (
	"strconv"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
	"github.com/0xdevelop/fltk2go/foundation"
	"github.com/0xdevelop/fltk2go/uikit/button"
	"github.com/0xdevelop/fltk2go/uikit/label"
	"github.com/0xdevelop/fltk2go/uikit/tableview"
	"github.com/0xdevelop/fltk2go/uikit/view"
)

const (
	Primary     uint = 0x2563EB00
	Slate       uint = 0x1E293B00
	Muted       uint = 0x64748B00
	Panel       uint = 0xF8FAFC00
	Card        uint = 0xFFFFFF00
	Line        uint = 0xE2E8F000
	Green       uint = 0x22C55E00
	Red         uint = 0xEF444400
	ButtonMuted uint = 0x47556900
	White       uint = 0xFFFFFFFF
)

type Item struct {
	ID    string
	Name  string
	Value string
}

type ItemDataSource struct {
	items []Item
}

func NewItemDataSource() *ItemDataSource {
	return &ItemDataSource{items: []Item{
		{ID: "1", Name: "Navigation", Value: "Ready"},
		{ID: "2", Name: "Forms", Value: "Validated"},
		{ID: "3", Name: "Buttons", Value: "Interactive"},
		{ID: "4", Name: "Table rows", Value: "Reusable"},
		{ID: "5", Name: "Layout", Value: "Spacious"},
	}}
}

func (ds *ItemDataSource) NumberOfRows(tv *tableview.TableView) int { return len(ds.items) }

func (ds *ItemDataSource) CellForColumn(tv *tableview.TableView, row, col int) *tableview.TableViewCell {
	cell := tv.Dequeue("itemCell")
	if cell == nil {
		cell = tableview.NewCell("itemCell")
	}
	_ = ds.items[row]
	return cell
}

type ItemDelegate struct{}

func (d *ItemDelegate) DidSelectRow(tv *tableview.TableView, row int) {
	print("Selected row:", row, "\n")
}
func (d *ItemDelegate) RowHeight(tv *tableview.TableView, row int) int { return 48 }

func BuildView(parent *view.UIView) view.Viewable {
	background := label.NewUILabel(&foundation.Rect{X: 0, Y: 0, Width: 1060, Height: 740}, "")
	background.SetBackgroundColor(Panel)
	background.SetFrame(fltk_bridge.FLAT_BOX)
	parent.AddSubview(background)

	title := label.NewUILabel(&foundation.Rect{X: 56, Y: 30, Width: 948, Height: 34}, "FLTK2Go Component Gallery")
	title.SetFontSize(25)
	title.SetFont(fltk_bridge.HELVETICA_BOLD)
	title.SetTextColor(Slate)
	parent.AddSubview(title)

	description := label.NewUILabel(&foundation.Rect{X: 56, Y: 68, Width: 948, Height: 28}, "A cleaner desktop demo: grouped controls, spacious rows, and semantic action colors.")
	description.SetFontSize(14)
	description.SetTextColor(Muted)
	parent.AddSubview(description)

	buttonCard := label.NewUILabel(&foundation.Rect{X: 56, Y: 124, Width: 948, Height: 150}, "")
	buttonCard.SetBackgroundColor(Card)
	buttonCard.SetFrame(fltk_bridge.ROUNDED_BOX)
	parent.AddSubview(buttonCard)

	buttonTitle := label.NewUILabel(&foundation.Rect{X: 88, Y: 150, Width: 420, Height: 28}, "Button states")
	buttonTitle.SetFontSize(18)
	buttonTitle.SetFont(fltk_bridge.HELVETICA_BOLD)
	buttonTitle.SetTextColor(Slate)
	parent.AddSubview(buttonTitle)

	buttonHint := label.NewUILabel(&foundation.Rect{X: 88, Y: 180, Width: 780, Height: 24}, "Minimum 44px height, clear gaps, and one primary visual action per row.")
	buttonHint.SetFontSize(13)
	buttonHint.SetTextColor(Muted)
	parent.AddSubview(buttonHint)

	systemBtn := button.NewUIButton(&foundation.Rect{X: 88, Y: 214, Width: 150, Height: 44}, "Primary")
	systemBtn.SetBackgroundColor(Primary)
	systemBtn.SetTitleColor(White)
	systemBtn.View().SetTooltip("Run the primary action")
	parent.AddSubview(systemBtn)

	checkBtn := button.NewUIButtonWithType(&foundation.Rect{X: 264, Y: 214, Width: 150, Height: 44}, "Checkbox", button.CheckboxButton)
	parent.AddSubview(checkBtn)

	radioBtn := button.NewUIButtonWithType(&foundation.Rect{X: 440, Y: 214, Width: 150, Height: 44}, "Radio", button.RadioButton)
	parent.AddSubview(radioBtn)

	toggleBtn := button.NewUIButtonWithType(&foundation.Rect{X: 616, Y: 214, Width: 150, Height: 44}, "Toggle", button.ToggleButton)
	parent.AddSubview(toggleBtn)

	tableCard := label.NewUILabel(&foundation.Rect{X: 56, Y: 304, Width: 948, Height: 342}, "")
	tableCard.SetBackgroundColor(Card)
	tableCard.SetFrame(fltk_bridge.ROUNDED_BOX)
	parent.AddSubview(tableCard)

	tableTitle := label.NewUILabel(&foundation.Rect{X: 88, Y: 330, Width: 500, Height: 28}, "TableView data source")
	tableTitle.SetFontSize(18)
	tableTitle.SetFont(fltk_bridge.HELVETICA_BOLD)
	tableTitle.SetTextColor(Slate)
	parent.AddSubview(tableTitle)

	tableHint := label.NewUILabel(&foundation.Rect{X: 88, Y: 360, Width: 780, Height: 24}, "Custom row rendering with stronger contrast and better scan rhythm.")
	tableHint.SetFontSize(13)
	tableHint.SetTextColor(Muted)
	parent.AddSubview(tableHint)

	tv, _ := tableview.New(88, 396, 884, 218)
	dataSource := NewItemDataSource()
	delegate := &ItemDelegate{}
	tv.SetDataSource(dataSource)
	tv.SetDelegate(delegate)

	tv.SetCustomDraw(func(ctx fltk_bridge.TableContext, row, col, x, y, w, h int) {
		if ctx != fltk_bridge.ContextCell || row < 0 || row >= len(dataSource.items) {
			return
		}
		item := dataSource.items[row]
		bg := fltk_bridge.Color(0xFFFFFF00)
		if row%2 == 1 {
			bg = fltk_bridge.Color(0xF8FAFC00)
		}
		fltk_bridge.DrawRectfWithColor(x, y, w, h, bg)
		fltk_bridge.SetDrawColor(fltk_bridge.Color(Line))
		fltk_bridge.DrawRect(x, y, w, h)
		fltk_bridge.SetDrawColor(fltk_bridge.Color(Slate))
		fltk_bridge.SetDrawFont(fltk_bridge.HELVETICA_BOLD, 14)
		fltk_bridge.Draw(item.Name, x+18, y, w/2-24, h, fltk_bridge.ALIGN_LEFT|fltk_bridge.ALIGN_INSIDE)
		fltk_bridge.SetDrawColor(fltk_bridge.Color(Muted))
		fltk_bridge.SetDrawFont(fltk_bridge.HELVETICA, 13)
		fltk_bridge.Draw(item.Value, x+w/2, y, w/2-18, h, fltk_bridge.ALIGN_LEFT|fltk_bridge.ALIGN_INSIDE)
	})
	parent.AddSubview(tv)
	tv.ReloadData()

	addBtn := button.NewUIButton(&foundation.Rect{X: 56, Y: 670, Width: 150, Height: 44}, "Add item")
	addBtn.SetBackgroundColor(Green)
	addBtn.SetTitleColor(White)
	addBtn.View().SetTooltip("Add a new table row")

	removeBtn := button.NewUIButton(&foundation.Rect{X: 226, Y: 670, Width: 150, Height: 44}, "Remove")
	removeBtn.SetBackgroundColor(Red)
	removeBtn.SetTitleColor(White)
	removeBtn.View().SetTooltip("Remove the selected table row")

	refreshBtn := button.NewUIButton(&foundation.Rect{X: 396, Y: 670, Width: 150, Height: 44}, "Refresh")
	refreshBtn.SetBackgroundColor(ButtonMuted)
	refreshBtn.SetTitleColor(White)
	refreshBtn.View().SetTooltip("Reload the table contents")

	addBtn.OnTouchUpInside(func() {
		newItem := Item{
			ID:    strconv.Itoa(len(dataSource.items) + 1),
			Name:  "Item " + strconv.Itoa(len(dataSource.items)+1),
			Value: "New row",
		}
		dataSource.items = append(dataSource.items, newItem)
		tv.ReloadData()
	})
	removeBtn.OnTouchUpInside(func() {
		selectedRow := tv.GetSelectedRow()
		if selectedRow >= 0 && selectedRow < len(dataSource.items) {
			dataSource.items = append(dataSource.items[:selectedRow], dataSource.items[selectedRow+1:]...)
			tv.ReloadData()
		}
	})
	refreshBtn.OnTouchUpInside(func() { tv.ReloadData() })

	parent.AddSubview(addBtn)
	parent.AddSubview(removeBtn)
	parent.AddSubview(refreshBtn)

	return nil
}
