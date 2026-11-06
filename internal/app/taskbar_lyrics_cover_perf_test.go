package app

import (
	"image/color"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
	"lilt/internal/media/qqmusic"
)

// A ready taskbar cover must not introduce an artwork fade or keep paused lyrics rebuilding.
func TestTaskbarLyricsReadyCoverFrameBudget(t *testing.T) {
	a, d := taskbarLyricsViewFixture()
	track := *a.pl.Current()
	track.Album.PicURL = "https://example.invalid/taskbar-paused-cover.jpg"
	a.pl.Cue([]qqmusic.Track{track})
	a.art.Add(qqmusic.Sized(track.Album.PicURL, 96), &artwork{BMP: taskbarCoverBitmap(color.RGBA{R: 255, A: 255})})
	builds := 0
	tt := ui.NewTester(func(c *ui.Context) { builds++; d.View(c) }, 280, 42)
	tt.Frame()
	if red, _ := taskbarCoverColors(tt.Image()); red == 0 {
		t.Fatal("fixture did not paint the ready cover")
	}
	before := builds
	tt.Advance(time.Second)
	if builds != before {
		t.Fatalf("paused cover kept rebuilding: %d extra frames", builds-before)
	}
}
