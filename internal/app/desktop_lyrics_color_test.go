package app

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/features/lyricsui"
)

// Presets and custom colors save across click and keyboard picker openings, Escape, and outside clicks.
func TestDesktopLyricsColorSettings(t *testing.T) {
	if !supportsDesktopLyrics() {
		t.Skip("no desktop lyrics settings on this platform")
	}
	a, _ := desktopLyricsViewFixture()
	a.settings.Path = filepath.Join(t.TempDir(), "settings.json")
	a.settingsTab = int(settingsPlayback)
	tt := ui.NewTester(a.settingsView, 560, 2000)
	tt.SetLocale("en")
	tt.Frame()
	check := func(want ui.Color) {
		t.Helper()
		restored := readDesktopLyricsEntrySettings(t, a)
		if got := lyricsui.DesktopLyricsColor(a.settings.DesktopLyricsColor); got != want {
			t.Fatalf("live lyrics color = %#v, want %#v", got, want)
		}
		if got := lyricsui.DesktopLyricsColor(restored.DesktopLyricsColor); got != want {
			t.Fatalf("saved lyrics color = %#v, want %#v", got, want)
		}
	}
	if err := tt.Click("桌面歌词颜色：薄荷"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	check(ui.Hex("#b1f1cb"))
	for i, value := range []string{"#9ab8d6", "#b7a1c980", "#648fba"} {
		switch i {
		case 0:
			if err := tt.Click("自定义桌面歌词颜色"); err != nil {
				t.Fatal(err)
			}
		case 1:
			tt.Key(0, ui.KeySpace)
		case 2:
			tt.Key(0, ui.KeyEnter)
		}
		tt.Frame()
		if err := tt.Click("Hex"); err != nil {
			t.Fatal(err)
		}
		tt.Key(ui.Cmd, ui.KeyA)
		tt.Type(value)
		tt.Frame()
		check(ui.Hex(value))
		tt.Key(0, ui.KeyEscape)
		tt.Frame()
		check(ui.Hex(value))
	}
	if err := tt.Click("自定义桌面歌词颜色"); err != nil {
		t.Fatal(err)
	}
	if err := tt.Click("桌面歌词颜色：月白"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	check(ui.Hex("#ffffff"))
}

// Export the settings row and open picker at narrow widths for visual inspection in both themes.
func TestDesktopLyricsColorFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip("set NM_FRAMES to inspect desktop lyric colors")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, width := range []int{480, 560} {
		for _, dark := range []bool{false, true} {
			a := newTestApp()
			a.settings = defaultSettings()
			a.settings.DesktopLyricsColor = "#b1f1cb"
			a.settingsTab = int(settingsPlayback)
			tt := ui.NewTester(a.settingsView, width, 1100)
			tt.SetLocale("zh-CN")
			tt.SetDark(dark)
			tt.Frame()
			prefix := fmt.Sprintf("lyrics-color-%d-dark-%t", width, dark)
			saveLyricsEntryFrame(t, dir, prefix+".png", tt)
			if err := tt.Click("自定义桌面歌词颜色"); err != nil {
				t.Fatal(err)
			}
			tt.Move(-10, -10)
			tt.Advance(200 * time.Millisecond)
			saveLyricsEntryFrame(t, dir, prefix+"-picker.png", tt)
		}
	}
}

// Zero-opacity custom colors must hide the complete lyric layer instead of inheriting white text or leaving outlines.
func TestDesktopLyricsColorTransparency(t *testing.T) {
	a, d := desktopLyricsViewFixture()
	a.settings.DesktopLyricsLocked = true
	tt := ui.NewTester(d.View, 760, 240)
	tt.SetPreferences(ui.Preferences{ReduceMotion: true})
	visible := func() bool {
		t.Helper()
		tt.Frame()
		frame := tt.Image()
		for y := 0; y < frame.Bounds().Dy(); y++ {
			for x := 0; x < frame.Bounds().Dx(); x++ {
				if _, _, _, alpha := frame.At(x, y).RGBA(); alpha != 0 {
					return true
				}
			}
		}
		return false
	}
	if !visible() {
		t.Fatal("fixture has no visible lyrics")
	}
	a.settings.DesktopLyricsColor = "#b1f1cb00"
	if visible() {
		t.Fatal("fully transparent lyrics retained text or outlines")
	}
	a.settings.DesktopLyricsColor = "#b1f1cb"
	if !visible() {
		t.Fatal("restoring opacity did not reveal the lyrics")
	}
}
