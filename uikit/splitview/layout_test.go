package splitview

import (
	"math"
	"testing"
)

func TestLayoutModelClampsRatioToIndependentMinima(t *testing.T) {
	m := newLayoutModel(1000, 8)
	m.setMinimumSizes(250, 320)
	m.setRatio(0.9)

	got := m.geometry()
	if got.Leading != 672 || got.Trailing != 320 || got.Divider != 8 {
		t.Fatalf("geometry = %#v, want leading=672 divider=8 trailing=320", got)
	}
	if got.Leading+got.Divider+got.Trailing != 1000 {
		t.Fatalf("geometry does not fill extent: %#v", got)
	}
}

func TestLayoutModelPreservesRatioAcrossResize(t *testing.T) {
	m := newLayoutModel(1000, 8)
	m.setMinimumSizes(100, 100)
	m.setRatio(0.35)
	before := m.geometry()

	m.setExtent(1200)
	after := m.geometry()
	if math.Abs(before.Ratio-after.Ratio) > 0.002 {
		t.Fatalf("ratio changed across resize: before=%f after=%f", before.Ratio, after.Ratio)
	}
	if after.Leading != 420 || after.Trailing != 772 {
		t.Fatalf("resized geometry = %#v, want leading=420 trailing=772", after)
	}
}

func TestLayoutModelSupportsLeadingAndTrailingPixelPolicies(t *testing.T) {
	leading := newLayoutModel(1000, 8)
	leading.setMinimumSizes(100, 100)
	leading.setLeadingPixels(320)
	leading.setExtent(1200)
	if got := leading.geometry(); got.Leading != 320 || got.Trailing != 872 {
		t.Fatalf("leading pixel policy geometry = %#v", got)
	}

	trailing := newLayoutModel(1000, 8)
	trailing.setMinimumSizes(100, 100)
	trailing.setTrailingPixels(400)
	trailing.setExtent(1200)
	if got := trailing.geometry(); got.Leading != 792 || got.Trailing != 400 {
		t.Fatalf("trailing pixel policy geometry = %#v", got)
	}
}

func TestLayoutModelDragPreservesSelectedResizePolicy(t *testing.T) {
	tests := []struct {
		name         string
		configure    func(*layoutModel)
		dragLeading  int
		wantLeading  int
		wantTrailing int
	}{
		{name: "ratio", configure: func(m *layoutModel) { m.setRatio(0.5) }, dragLeading: 300, wantLeading: 360, wantTrailing: 832},
		{name: "leading pixels", configure: func(m *layoutModel) { m.setLeadingPixels(300) }, dragLeading: 400, wantLeading: 400, wantTrailing: 792},
		{name: "trailing pixels", configure: func(m *layoutModel) { m.setTrailingPixels(400) }, dragLeading: 500, wantLeading: 700, wantTrailing: 492},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newLayoutModel(1000, 8)
			tt.configure(m)
			m.setLeadingForCurrentPolicy(tt.dragLeading)
			m.setExtent(1200)
			got := m.geometry()
			if got.Leading != tt.wantLeading || got.Trailing != tt.wantTrailing {
				t.Fatalf("resized drag geometry = %#v, want leading=%d trailing=%d", got, tt.wantLeading, tt.wantTrailing)
			}
		})
	}
}

func TestLayoutModelClampedDragPersistsActualGeometryForPolicy(t *testing.T) {
	tests := []struct {
		name         string
		configure    func(*layoutModel)
		wantLeading  int
		wantTrailing int
	}{
		{name: "ratio", configure: func(m *layoutModel) { m.setRatio(0.5) }, wantLeading: 120, wantTrailing: 1072},
		{name: "leading pixels", configure: func(m *layoutModel) { m.setLeadingPixels(500) }, wantLeading: 100, wantTrailing: 1092},
		{name: "trailing pixels", configure: func(m *layoutModel) { m.setTrailingPixels(400) }, wantLeading: 300, wantTrailing: 892},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newLayoutModel(1000, 8)
			m.setMinimumSizes(100, 200)
			tt.configure(m)
			m.setLeadingForCurrentPolicy(-100)
			if got := m.geometry(); got.Leading != 100 {
				t.Fatalf("clamped drag leading = %d, want 100", got.Leading)
			}
			m.setExtent(1200)
			got := m.geometry()
			if got.Leading != tt.wantLeading || got.Trailing != tt.wantTrailing {
				t.Fatalf("resized clamped geometry = %#v, want leading=%d trailing=%d", got, tt.wantLeading, tt.wantTrailing)
			}
		})
	}
}

func TestLayoutModelCompressesImpossibleMinimaDeterministically(t *testing.T) {
	m := newLayoutModel(500, 8)
	m.setMinimumSizes(300, 400)
	m.setRatio(0.5)

	got := m.geometry()
	if got.Leading != 211 || got.Trailing != 281 || got.Divider != 8 {
		t.Fatalf("constrained geometry = %#v, want proportional 211/8/281", got)
	}
	if got.Leading < 0 || got.Trailing < 0 || got.Leading+got.Divider+got.Trailing != 500 {
		t.Fatalf("invalid constrained geometry: %#v", got)
	}
}

func TestLayoutModelHandlesDividerLargerThanExtent(t *testing.T) {
	m := newLayoutModel(4, 8)
	m.setMinimumSizes(10, 10)
	got := m.geometry()
	if got.Leading != 0 || got.Divider != 4 || got.Trailing != 0 {
		t.Fatalf("tiny geometry = %#v, want 0/4/0", got)
	}
}

func TestLayoutModelDoesNotOverflowExtremeMinimums(t *testing.T) {
	m := newLayoutModel(500, 8)
	maximum := int(^uint(0) >> 1)
	m.setMinimumSizes(maximum, maximum)
	got := m.geometry()
	if got.Leading != 246 || got.Trailing != 246 || got.Divider != 8 {
		t.Fatalf("extreme-minimum geometry = %#v", got)
	}
}

func TestLayoutModelClampsInvalidValues(t *testing.T) {
	m := newLayoutModel(-1, -1)
	m.setMinimumSizes(-20, -30)
	m.setRatio(math.NaN())
	got := m.geometry()
	if got.Leading != 0 || got.Divider != 0 || got.Trailing != 0 || got.Ratio != 0 {
		t.Fatalf("invalid inputs were not normalized: %#v", got)
	}
}
