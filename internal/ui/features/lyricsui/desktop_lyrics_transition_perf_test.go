package lyricsui

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// A sentence change requests a bounded burst of frames and becomes idle after both layers settle.
func TestDesktopLyricsTransitionFrameBudget(t *testing.T) {
	var state desktopLyricViewState
	started, builds := false, 0
	tt := ui.NewTester(func(c *ui.Context) {
		builds++
		if !started {
			state.changed, started = c.Now(), true
		}
		state.blend(c)
	}, 100, 100)
	tt.Frame()
	before := builds
	tt.Advance(DesktopFade + 40*time.Millisecond)
	if got := builds - before; got < 12 || got > 18 {
		t.Fatalf("sentence transition rebuilt %d times, want a short 60 Hz burst", got)
	}
	before = builds
	tt.Advance(time.Second)
	if builds != before {
		t.Fatalf("settled transition kept requesting frames: %d", builds-before)
	}
}
