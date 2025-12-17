package colors

import (
	"math"
	"strconv"
	"strings"
)

func toByte(v any) (uint8, bool) {
	switch n := v.(type) {
	case uint8:
		return n, true
	case int:
		return clampIntToByte(n), true
	case int32:
		return clampIntToByte(int(n)), true
	case int64:
		return clampIntToByte(int(n)), true
	case uint:
		if n > 255 {
			return 255, true
		}
		return uint8(n), true
	case uint32:
		if n > 255 {
			return 255, true
		}
		return uint8(n), true
	case uint64:
		if n > 255 {
			return 255, true
		}
		return uint8(n), true
	default:
		return 0, false
	}
}

func clampIntToByte(i int) uint8 {
	if i < 0 {
		return 0
	}
	if i > 255 {
		return 255
	}
	return uint8(i)
}

// 允许把 0xRRGGBB / 0xRRGGBBAA 这种整数直接传进来
func parseColorUint64(v uint64) (Color, bool) {
	// 如果大于 0xFFFFFF，当作 0xRRGGBBAA（忽略 AA）
	if v > 0xFFFFFF {
		r := uint8((v >> 24) & 0xFF)
		g := uint8((v >> 16) & 0xFF)
		b := uint8((v >> 8) & 0xFF)
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	}
	r := uint8((v >> 16) & 0xFF)
	g := uint8((v >> 8) & 0xFF)
	b := uint8(v & 0xFF)
	return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
}

func parseColorString(s string) (Color, bool) {
	in := strings.TrimSpace(strings.ToLower(s))
	if in == "" {
		return Color{}, false
	}

	// rgb()/rgba()
	if strings.HasPrefix(in, "rgb(") || strings.HasPrefix(in, "rgba(") {
		if c, ok := parseRGBFunc(in); ok {
			return c, true
		}
		return Color{}, false
	}

	// "r,g,b[,a]"
	if strings.Contains(in, ",") && !strings.Contains(in, "(") {
		if c, ok := parseCSV(in); ok {
			return c, true
		}
	}

	// hex：# / 0x / 纯hex（3/4/6/8）
	if c, ok := parseHexAny(in); ok {
		return c, true
	}

	return Color{}, false
}

func parseHexAny(s string) (Color, bool) {
	hex := strings.TrimSpace(s)
	hex = strings.TrimPrefix(hex, "#")
	hex = strings.TrimPrefix(hex, "0x")
	hex = strings.TrimPrefix(hex, "0X")
	hex = strings.ReplaceAll(hex, "_", "")

	switch len(hex) {
	case 3: // RGB
		r := hexNibbleDup(hex[0])
		g := hexNibbleDup(hex[1])
		b := hexNibbleDup(hex[2])
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	case 4: // RGBA（忽略 A）
		r := hexNibbleDup(hex[0])
		g := hexNibbleDup(hex[1])
		b := hexNibbleDup(hex[2])
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	case 6: // RRGGBB
		r, g, b, ok := parseHexRGB(hex)
		if !ok {
			return Color{}, false
		}
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	case 8: // RRGGBBAA（忽略 AA）
		r, g, b, ok := parseHexRGB(hex[:6])
		if !ok {
			return Color{}, false
		}
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	default:
		return Color{}, false
	}
}

func parseHexRGB(hex6 string) (r, g, b uint8, ok bool) {
	if len(hex6) != 6 {
		return 0, 0, 0, false
	}
	rb, ok1 := parseHexByte(hex6[0:2])
	gb, ok2 := parseHexByte(hex6[2:4])
	bb, ok3 := parseHexByte(hex6[4:6])
	if !(ok1 && ok2 && ok3) {
		return 0, 0, 0, false
	}
	return rb, gb, bb, true
}

func parseHexByte(h2 string) (uint8, bool) {
	v, err := strconv.ParseUint(h2, 16, 8)
	if err != nil {
		return 0, false
	}
	return uint8(v), true
}

func hexNibbleDup(ch byte) uint8 {
	v, err := strconv.ParseUint(string([]byte{ch}), 16, 8)
	if err != nil {
		return 0
	}
	return uint8(v*16 + v)
}

func parseRGBFunc(s string) (Color, bool) {
	open := strings.IndexByte(s, '(')
	close := strings.LastIndexByte(s, ')')
	if open < 0 || close <= open {
		return Color{}, false
	}
	fn := strings.TrimSpace(s[:open])
	body := s[open+1 : close]
	args := splitArgs(body)

	if fn == "rgb" {
		if len(args) != 3 {
			return Color{}, false
		}
		r, ok := parse255(args[0])
		if !ok {
			return Color{}, false
		}
		g, ok := parse255(args[1])
		if !ok {
			return Color{}, false
		}
		b, ok := parse255(args[2])
		if !ok {
			return Color{}, false
		}
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	}

	if fn == "rgba" {
		if len(args) != 4 {
			return Color{}, false
		}
		r, ok := parse255(args[0])
		if !ok {
			return Color{}, false
		}
		g, ok := parse255(args[1])
		if !ok {
			return Color{}, false
		}
		b, ok := parse255(args[2])
		if !ok {
			return Color{}, false
		}
		// 第 4 个 alpha 解析但忽略（最小改动；未来你要 alpha 再扩展 Color）
		_, _ = parseAlpha(args[3])
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
	}

	return Color{}, false
}

func parseCSV(s string) (Color, bool) {
	args := splitArgs(s)
	if len(args) != 3 && len(args) != 4 {
		return Color{}, false
	}
	r, ok := parse255(args[0])
	if !ok {
		return Color{}, false
	}
	g, ok := parse255(args[1])
	if !ok {
		return Color{}, false
	}
	b, ok := parse255(args[2])
	if !ok {
		return Color{}, false
	}
	if len(args) == 4 {
		_, _ = parseAlpha(args[3]) // 解析但忽略
	}
	return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}, true
}

func splitArgs(s string) []string {
	raw := strings.Split(s, ",")
	out := make([]string, 0, len(raw))
	for _, p := range raw {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

func parse255(s string) (uint8, bool) {
	s = strings.TrimSpace(s)

	// 允许百分比：50%
	if strings.HasSuffix(s, "%") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return 0, false
		}
		f = clampFloat(f, 0, 100)
		return uint8(math.Round(f * 255 / 100)), true
	}

	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, false
	}
	if i < 0 {
		i = 0
	}
	if i > 255 {
		i = 255
	}
	return uint8(i), true
}

func parseAlpha(s string) (uint8, bool) {
	s = strings.TrimSpace(s)

	if strings.HasSuffix(s, "%") {
		f, err := strconv.ParseFloat(strings.TrimSuffix(s, "%"), 64)
		if err != nil {
			return 0, false
		}
		f = clampFloat(f, 0, 100)
		return uint8(math.Round(f * 255 / 100)), true
	}

	if strings.Contains(s, ".") {
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return 0, false
		}
		f = clampFloat(f, 0, 1)
		return uint8(math.Round(f * 255)), true
	}

	return parse255(s)
}

func clampFloat(v, min, max float64) float64 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}
