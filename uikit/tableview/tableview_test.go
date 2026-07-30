package tableview

import (
	"testing"

	"github.com/0xdevelop/fltk2go/fltk_bridge"
)

type fakeBridgeTable struct {
	rows       int
	redraws    int
	background fltk_bridge.Color
	draw       func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int)
	event      func(row int) bool
}

func (f *fakeBridgeTable) SetRows(rows int) { f.rows = rows }
func (f *fakeBridgeTable) Redraw()          { f.redraws++ }
func (f *fakeBridgeTable) SetDrawCellHandler(fn func(ctx fltk_bridge.TableContext, row, col int, x, y, w, h int)) {
	f.draw = fn
}
func (f *fakeBridgeTable) SetEventHandler(fn func(row int) bool)      { f.event = fn }
func (f *fakeBridgeTable) GetSelectedRow() int                        { return 3 }
func (f *fakeBridgeTable) Widget() fltk_bridge.Widget                 { return nil }
func (f *fakeBridgeTable) SetColumnCount(int)                         {}
func (f *fakeBridgeTable) SetColumnWidth(int, int)                    {}
func (f *fakeBridgeTable) AllowColumnResizing()                       {}
func (f *fakeBridgeTable) EnableColumnHeaders()                       {}
func (f *fakeBridgeTable) SetColumnHeaderHeight(int)                  {}
func (f *fakeBridgeTable) SetBackgroundColor(color fltk_bridge.Color) { f.background = color }

func TestSetBackgroundColorForwardsToBridge(t *testing.T) {
	bridge := &fakeBridgeTable{}
	tv := newWithBridgeTable(bridge)
	want := fltk_bridge.Color(0x12345600)
	tv.SetBackgroundColor(want)
	if bridge.background != want {
		t.Fatalf("background = %#x, want %#x", bridge.background, want)
	}
}

type sliceDataSource struct {
	rows  int
	cells map[int]*TableViewCell
}

func (s *sliceDataSource) NumberOfRows(*TableView) int { return s.rows }
func (s *sliceDataSource) CellForColumn(_ *TableView, row, col int) *TableViewCell {
	if s.cells == nil {
		s.cells = map[int]*TableViewCell{}
	}
	cell := NewCell("row")
	s.cells[row] = cell
	return cell
}

type recordingDelegate struct {
	selected []int
}

func (r *recordingDelegate) DidSelectRow(_ *TableView, row int) {
	r.selected = append(r.selected, row)
}
func (r *recordingDelegate) RowHeight(*TableView, int) int { return 0 }

func TestReloadDataUsesDataSourceAndClearsVisibleCells(t *testing.T) {
	bridge := &fakeBridgeTable{}
	tv := newWithBridgeTable(bridge)
	ds := &sliceDataSource{rows: 2}
	tv.SetDataSource(ds)

	tv.ReloadData()

	if bridge.rows != 2 {
		t.Fatalf("rows = %d, want 2", bridge.rows)
	}
	if bridge.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", bridge.redraws)
	}

	tv.cellFor(1, 0)
	if got := tv.visible["1_0"]; got == nil || got.Row() != 1 {
		t.Fatalf("visible row was not cached with row index: %#v", got)
	}

	ds.rows = 1
	tv.ReloadData()
	if len(tv.visible) != 0 {
		t.Fatalf("visible cells were not cleared: %d", len(tv.visible))
	}
	if got := len(tv.reusePool["row"]); got != 1 {
		t.Fatalf("reuse pool size = %d, want 1", got)
	}
}

func TestReloadDataWithoutDataSourceClearsRows(t *testing.T) {
	bridge := &fakeBridgeTable{}
	tv := newWithBridgeTable(bridge)

	tv.ReloadData()

	if bridge.rows != 0 {
		t.Fatalf("rows = %d, want 0", bridge.rows)
	}
	if bridge.redraws != 1 {
		t.Fatalf("redraws = %d, want 1", bridge.redraws)
	}
}

func TestEventDelegateIgnoresNegativeRows(t *testing.T) {
	bridge := &fakeBridgeTable{}
	tv := newWithBridgeTable(bridge)
	delegate := &recordingDelegate{}
	tv.SetDelegate(delegate)

	if bridge.event(-1) {
		t.Fatal("event(-1) = true, want false")
	}
	if !bridge.event(4) {
		t.Fatal("event(4) = false, want true")
	}
	if len(delegate.selected) != 1 || delegate.selected[0] != 4 {
		t.Fatalf("selected rows = %#v, want [4]", delegate.selected)
	}
}

func TestTableViewNilSafety(t *testing.T) {
	var tv *TableView
	tv.SetDataSource(nil)
	tv.SetDelegate(nil)
	tv.SetDefaultRowHeight(10)
	tv.SetCustomDraw(nil)
	tv.Enqueue(nil)
	tv.ReloadData()
	tv.onDrawCell(fltk_bridge.ContextCell, 0, 0, 0, 0, 0, 0)

	if tv.View() != nil {
		t.Fatal("nil TableView View() returned non-nil")
	}
	if tv.Raw() != nil {
		t.Fatal("nil TableView Raw() returned non-nil")
	}
	if tv.GetSelectedRow() != -1 {
		t.Fatal("nil TableView selected row should be -1")
	}
	if !tv.onEvent(0) {
		// expected false; keep branch explicit to ensure no panic above
		return
	}
	t.Fatal("nil TableView onEvent returned true")
}
