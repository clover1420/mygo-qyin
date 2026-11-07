package lyricsui

import (
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/settings"
	"lilt/internal/ui/base/theme"
)

// DesktopLyricsColors are the soft, readable tints offered over the desktop,
// from the dark scales of the Radix palette.
var DesktopLyricsColors = []struct {
	Name string
	Hex  string
}{
	{"月白", "#ffffff"},
	{"薄荷", "#b1f1cb"},
	{"雾蓝", "#c2e6ff"},
	{"淡紫", "#e0dffe"},
	{"暖杏", "#ffe7b3"},
	{"浅粉", "#fdd1ea"},
}

// DesktopLyricsColorWheel is the hue ring of the custom color well.
var DesktopLyricsColorWheel = ui.ConicGradient{Stops: []ui.GradientStop{
	{Color: ui.Hex("#e5484d"), At: 0},
	{Color: ui.Hex("#f76b15"), At: 0.125},
	{Color: ui.Hex("#ffc53d"), At: 0.25},
	{Color: ui.Hex("#30a46c"), At: 0.375},
	{Color: ui.Hex("#00a2c7"), At: 0.5},
	{Color: ui.Hex("#0090ff"), At: 0.625},
	{Color: ui.Hex("#8e4ec6"), At: 0.75},
	{Color: ui.Hex("#d6409f"), At: 0.875},
	{Color: ui.Hex("#e5484d"), At: 1},
}}

// DesktopLyricsColor reads the saved color, falling back to white.
func DesktopLyricsColor(value string) ui.Color {
	value = strings.TrimPrefix(strings.TrimSpace(value), "#")
	if len(value) == 6 || len(value) == 8 {
		if b, err := hex.DecodeString(value); err == nil {
			color := ui.RGB(b[0], b[1], b[2])
			if len(b) == 4 {
				color.A = b[3]
			}
			return color
		}
	}
	return theme.Paper
}

// SetDesktopLyricsColor stores the color in the settings, empty for white.
func SetDesktopLyricsColor(s *settings.Settings, color ui.Color) {
	value := fmt.Sprintf("#%02x%02x%02x", color.R, color.G, color.B)
	if color.A != 255 {
		value += fmt.Sprintf("%02x", color.A)
	}
	if color == theme.Paper {
		value = ""
	}
	s.DesktopLyricsColor = value
}

// desktopLyricWordColor keeps white's original dim-to-bright treatment; other
// colors fade from white to the selected color as each word is sung.
// Opacity belongs to the whole layer.
func desktopLyricWordColor(color ui.Color, progress float32) ui.Color {
	if color == theme.Paper {
		brightness := DesktopRestAlpha + (DesktopSungAlpha-DesktopRestAlpha)*progress
		return color.Alpha(brightness)
	}
	return theme.Paper.Mix(color, progress).Alpha(DesktopSungAlpha)
}
