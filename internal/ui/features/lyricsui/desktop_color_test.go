package lyricsui

import (
	"testing"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/settings"
	"lilt/internal/ui/base/theme"
)

// Invalid persisted colors fall back to visible white while valid custom RGB and alpha values survive loading.
func TestDesktopLyricsColorParsing(t *testing.T) {
	for _, value := range []string{"", "#bad", "#gg1122", "#1122334455", "#1122334"} {
		s := settings.Settings{DesktopLyricsColor: value}
		if got := DesktopLyricsColor(s.DesktopLyricsColor); got != theme.Paper {
			t.Errorf("invalid color %q = %#v, want white", value, got)
		}
	}
	for _, tc := range []struct {
		value string
		want  ui.Color
	}{
		{"#9Ab8D6", ui.RGB(0x9a, 0xb8, 0xd6)},
		{"  9ab8d680  ", ui.Color{R: 0x9a, G: 0xb8, B: 0xd6, A: 0x80}},
	} {
		if got := DesktopLyricsColor(tc.value); got != tc.want {
			t.Errorf("color %q = %#v, want %#v", tc.value, got, tc.want)
		}
	}
}

// Colored words fade from white to the selected color without dimming, while the original white keeps its brightness ramp.
func TestDesktopLyricsColorProgress(t *testing.T) {
	for _, color := range []ui.Color{ui.Hex("#b1f1cb"), ui.Hex("#314159")} {
		for _, progress := range []float32{0, 0.5, 1} {
			got := desktopLyricWordColor(color, progress)
			if got.A != theme.Veil(12).A {
				t.Fatalf("colored word opacity changed at progress %v: %v", progress, got.A)
			}
			if progress == 0 && got != theme.Veil(12) {
				t.Fatalf("unsung word did not start white: %+v", got)
			}
			if progress == 1 && got != color.Alpha(DesktopSungAlpha) {
				t.Fatalf("sung word lost selected color: %+v", got)
			}
			if progress == 0.5 && (got.R <= color.R || got.G <= color.G || got.B <= color.B || got.R == 255 || got.G == 255 || got.B == 255) {
				t.Fatalf("word did not gradually blend toward the selected color: %+v", got)
			}
		}
	}
	if got := desktopLyricWordColor(theme.Paper, 0); got != theme.Paper.Alpha(DesktopRestAlpha) {
		t.Fatalf("default white lost its unsung brightness: %+v", got)
	}
	if got := desktopLyricWordColor(theme.Paper, 1); got != theme.Veil(12) {
		t.Fatalf("default white lost its sung brightness: %+v", got)
	}
}
