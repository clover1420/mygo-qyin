package app

import (
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/features/lyricsui"
)

// Paused or dismissed taskbar controls must not ask for frames after their interaction animations settle.
func TestTaskbarLyricsControlsFrameBudget(t *testing.T) {
	for _, mode := range []string{"paused", "closed", "hidden"} {
		t.Run(mode, func(t *testing.T) {
			_, d := taskbarControlsFixture()
			d.ControlsOpen = true
			if mode == "closed" {
				d.CloseControls()
			}
			if mode == "hidden" {
				d.Hide()
			}
			builds := 0
			tt := ui.NewTester(func(c *ui.Context) { builds++; d.ControlsView(c) }, lyricsui.TaskbarControlsWidth, lyricsui.TaskbarControlsHeight)
			tt.Move(-10, -10)
			tt.Frame()
			tt.Advance(time.Second)
			before := builds
			tt.Advance(time.Second)
			if builds != before {
				t.Fatalf("%s taskbar controls kept rebuilding: %d extra frames", mode, builds-before)
			}
		})
	}
}
