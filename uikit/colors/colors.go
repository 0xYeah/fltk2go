package colors

import (
	"fmt"
	"strings"

	"github.com/0xYeah/fltk2go/fltk_bridge"
)

var (
	// System colors
	Foreground  = colorByName("foreground")
	Background  = colorByName("background")
	Background2 = colorByName("background2")
	Inactive    = colorByName("inactive")
	Selection   = colorByName("selection")

	// Grays
	Gray0  = colorByName("gray0")
	Dark3  = colorByName("dark3")
	Dark2  = colorByName("dark2")
	Dark1  = colorByName("dark1")
	Light1 = colorByName("light1")
	Light2 = colorByName("light2")
	Light3 = colorByName("light3")

	// Base colors
	Black   = colorByName("black")
	White   = colorByName("white")
	Red     = colorByName("red")
	Green   = colorByName("green")
	Blue    = colorByName("blue")
	Yellow  = colorByName("yellow")
	Magenta = colorByName("magenta")
	Cyan    = colorByName("cyan")

	// Dark variants
	DarkRed     = colorByName("dark-red")
	DarkGreen   = colorByName("dark-green")
	DarkYellow  = colorByName("dark-yellow")
	DarkBlue    = colorByName("dark-blue")
	DarkMagenta = colorByName("dark-magenta")
	DarkCyan    = colorByName("dark-cyan")
)

type Color struct {
	Rgb       uint   // 0xRRGGBB
	RgbHex    string // "#RRGGBB"
	RgbStr    string // "rgb(r,g,b)"
	ColorName string // optional, for display
}

// ColorWithRGB supports multiple input forms while keeping a single exported API.
//
// Supported:
//   - ColorWithRGB(uint8,uint8,uint8)
//   - ColorWithRGB(int,int,int) / uint/uint8 混用也可以（会 clamp 到 0~255）
//   - ColorWithRGB("#RGB" / "#RGBA" / "#RRGGBB" / "#RRGGBBAA")
//   - ColorWithRGB("0xRRGGBB" / "0xRRGGBBAA")
//   - ColorWithRGB("RRGGBB" / "RRGGBBAA")
//   - ColorWithRGB("rgb(r,g,b)" / "rgba(r,g,b,a)")
//   - ColorWithRGB("r,g,b" / "r,g,b,a")
func ColorWithRGB(args ...any) Color {
	// default white
	def := Color{Rgb: packRGB(255, 255, 255), RgbHex: rgbToHex(255, 255, 255), RgbStr: rgbToStr(255, 255, 255), ColorName: "rgb"}

	// 1) (r,g,b) 或 (r,g,b,a)
	if len(args) == 3 || len(args) == 4 {
		r, ok := toByte(args[0])
		if !ok {
			return def
		}
		g, ok := toByte(args[1])
		if !ok {
			return def
		}
		b, ok := toByte(args[2])
		if !ok {
			return def
		}
		// 第 4 个 alpha 允许传入，但 Color 结构不存 alpha（最小化改动）；你将来要用再扩展结构即可
		return Color{Rgb: packRGB(r, g, b), RgbHex: rgbToHex(r, g, b), RgbStr: rgbToStr(r, g, b), ColorName: "rgb"}
	}

	// 2) 单参数：string 或 整数 0xRRGGBB/0xRRGGBBAA
	if len(args) == 1 {
		switch v := args[0].(type) {
		case string:
			if c, ok := parseColorString(v); ok {
				return c
			}
			// 也允许走你现有的命名色（不额外暴露任何 API）
			return colorByName(strings.ToLower(strings.TrimSpace(v)))
		case uint:
			if c, ok := parseColorUint64(uint64(v)); ok {
				return c
			}
		case uint32:
			if c, ok := parseColorUint64(uint64(v)); ok {
				return c
			}
		case uint64:
			if c, ok := parseColorUint64(v); ok {
				return c
			}
		case int:
			if v < 0 {
				return def
			}
			if c, ok := parseColorUint64(uint64(v)); ok {
				return c
			}
		case int64:
			if v < 0 {
				return def
			}
			if c, ok := parseColorUint64(uint64(v)); ok {
				return c
			}
		}
	}

	return def
}

func rgbToStr(r, g, b uint8) string {
	return fmt.Sprintf("rgb(%d,%d,%d)", r, g, b)
}

func rgbToHex(r, g, b uint8) string {
	return fmt.Sprintf("#%02X%02X%02X", r, g, b)
}

func packRGB(r, g, b uint8) uint {
	return (uint(r) << 16) | (uint(g) << 8) | uint(b)
}

func unpackRGB(rgb uint) (r, g, b uint8) {
	r = uint8((rgb >> 16) & 0xFF)
	g = uint8((rgb >> 8) & 0xFF)
	b = uint8(rgb & 0xFF)
	return
}

func colorByName(name string) Color {
	// default white
	ac := ColorWithRGB(255, 255, 255)

	var fltkColor fltk_bridge.Color

	switch name {
	// System
	case "foreground":
		fltkColor = fltk_bridge.FOREGROUND_COLOR
	case "background":
		fltkColor = fltk_bridge.BACKGROUND_COLOR
	case "background2":
		fltkColor = fltk_bridge.BACKGROUND2_COLOR
	case "inactive":
		fltkColor = fltk_bridge.INACTIVE_COLOR
	case "selection":
		fltkColor = fltk_bridge.SELECTION_COLOR

	// Grays
	case "gray0":
		fltkColor = fltk_bridge.GRAY0
	case "dark3":
		fltkColor = fltk_bridge.DARK3
	case "dark2":
		fltkColor = fltk_bridge.DARK2
	case "dark1":
		fltkColor = fltk_bridge.DARK1
	case "light1":
		fltkColor = fltk_bridge.LIGHT1
	case "light2":
		fltkColor = fltk_bridge.LIGHT2
	case "light3":
		fltkColor = fltk_bridge.LIGHT3

	// Base
	case "black":
		fltkColor = fltk_bridge.BLACK
	case "white":
		fltkColor = fltk_bridge.WHITE
	case "red":
		fltkColor = fltk_bridge.RED
	case "green":
		fltkColor = fltk_bridge.GREEN
	case "blue":
		fltkColor = fltk_bridge.BLUE
	case "yellow":
		fltkColor = fltk_bridge.YELLOW
	case "magenta":
		fltkColor = fltk_bridge.MAGENTA
	case "cyan":
		fltkColor = fltk_bridge.CYAN

	// Dark variants
	case "dark-red":
		fltkColor = fltk_bridge.DARK_RED
	case "dark-green":
		fltkColor = fltk_bridge.DARK_GREEN
	case "dark-yellow":
		fltkColor = fltk_bridge.DARK_YELLOW
	case "dark-blue":
		fltkColor = fltk_bridge.DARK_BLUE
	case "dark-magenta":
		fltkColor = fltk_bridge.DARK_MAGENTA
	case "dark-cyan":
		fltkColor = fltk_bridge.DARK_CYAN
	default:
		// Unknown color name, return default white
		return ac
	}

	r, g, b := unpackRGB(uint(fltkColor))
	ac.Rgb = uint(fltkColor)
	ac.RgbHex = rgbToHex(r, g, b)
	ac.RgbStr = rgbToStr(r, g, b)
	return ac
}
