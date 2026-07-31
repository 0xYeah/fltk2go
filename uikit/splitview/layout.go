package splitview

import "math"

type positionPolicy int

const (
	preserveRatio positionPolicy = iota
	preserveLeading
	preserveTrailing
)

type splitGeometry struct {
	Leading  int
	Divider  int
	Trailing int
	Ratio    float64
}

type layoutModel struct {
	extent        int
	dividerSize   int
	leadingMin    int
	trailingMin   int
	policy        positionPolicy
	ratio         float64
	leadingPixel  int
	trailingPixel int
}

func newLayoutModel(extent, dividerSize int) *layoutModel {
	m := &layoutModel{ratio: 0.5, policy: preserveRatio}
	m.setExtent(extent)
	m.setDividerSize(dividerSize)
	return m
}

func (m *layoutModel) setExtent(extent int) {
	m.extent = max(0, extent)
}

func (m *layoutModel) setDividerSize(size int) {
	m.dividerSize = max(0, size)
}

func (m *layoutModel) setMinimumSizes(leading, trailing int) {
	m.leadingMin = max(0, leading)
	m.trailingMin = max(0, trailing)
}

func (m *layoutModel) setRatio(ratio float64) {
	if math.IsNaN(ratio) || math.IsInf(ratio, 0) {
		ratio = 0
	}
	m.ratio = min(1, max(0, ratio))
	m.policy = preserveRatio
}

func (m *layoutModel) setLeadingPixels(position int) {
	m.leadingPixel = max(0, position)
	m.policy = preserveLeading
}

func (m *layoutModel) setTrailingPixels(size int) {
	m.trailingPixel = max(0, size)
	m.policy = preserveTrailing
}

// setLeadingForCurrentPolicy applies a drag position without changing the
// caller-selected resize policy. It stores the actual clamped geometry so a
// later resize cannot jump toward an earlier unreachable request.
func (m *layoutModel) setLeadingForCurrentPolicy(position int) {
	divider := min(m.dividerSize, m.extent)
	available := max(0, m.extent-divider)
	minimumTotal := float64(m.leadingMin) + float64(m.trailingMin)
	if minimumTotal > float64(available) {
		position = m.geometry().Leading
	} else {
		position = min(available-m.trailingMin, max(m.leadingMin, position))
	}
	switch m.policy {
	case preserveLeading:
		m.leadingPixel = position
	case preserveTrailing:
		m.trailingPixel = available - position
	default:
		if m.extent == 0 {
			m.ratio = 0
		} else {
			m.ratio = min(1, float64(position)/float64(m.extent))
		}
	}
}

func (m *layoutModel) setPolicy(policy positionPolicy) {
	g := m.geometry()
	switch policy {
	case preserveLeading:
		m.leadingPixel = g.Leading
	case preserveTrailing:
		m.trailingPixel = g.Trailing
	default:
		policy = preserveRatio
		m.ratio = g.Ratio
	}
	m.policy = policy
}

func (m *layoutModel) geometry() splitGeometry {
	if m == nil {
		return splitGeometry{}
	}
	divider := min(m.dividerSize, m.extent)
	available := max(0, m.extent-divider)
	if available == 0 {
		return splitGeometry{Divider: divider}
	}

	minimumTotal := float64(m.leadingMin) + float64(m.trailingMin)
	if minimumTotal > float64(available) {
		leading := int(math.Round(float64(available) * float64(m.leadingMin) / minimumTotal))
		leading = min(available, max(0, leading))
		return splitGeometry{
			Leading: leading, Divider: divider, Trailing: available - leading,
			Ratio: m.ratioForLeading(leading),
		}
	}

	var target int
	switch m.policy {
	case preserveLeading:
		target = m.leadingPixel
	case preserveTrailing:
		target = available - m.trailingPixel
	default:
		target = int(math.Round(float64(m.extent) * m.ratio))
	}
	leading := min(available-m.trailingMin, max(m.leadingMin, target))
	return splitGeometry{
		Leading: leading, Divider: divider, Trailing: available - leading,
		Ratio: m.ratioForLeading(leading),
	}
}

func (m *layoutModel) ratioForLeading(leading int) float64 {
	if m == nil || m.extent == 0 {
		return 0
	}
	return float64(leading) / float64(m.extent)
}
