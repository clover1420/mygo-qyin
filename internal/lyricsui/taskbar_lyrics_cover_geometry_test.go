package lyricsui

import "testing"

// The clickable child stays on the cover at every DPI, without covering lyrics.
func TestTaskbarLyricsCoverPlacement(t *testing.T) {
	tests := []struct {
		name               string
		dpi                uint32
		height, x, y, size int
	}{
		{"100 percent", 96, 42, 106, 24, 34},
		{"150 percent", 144, 63, 109, 26, 51},
		{"200 percent", 192, 84, 112, 28, 68},
		{"short taskbar", 96, 28, 106, 22, 24},
		{"fractional DPI", 120, 53, 108, 25, 43},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := TaskbarCoverPlacement(Placement{Parent: 1, Taskbar: 2, X: 100, Y: 20, Width: 600, Height: tt.height, DPI: tt.dpi, Dark: true})
			if p.X != tt.x || p.Y != tt.y || p.Width != tt.size || p.Height != tt.size || p.Parent != 1 || p.Taskbar != 2 || p.DPI != tt.dpi || !p.Dark {
				t.Fatalf("unexpected placement: %+v", p)
			}
		})
	}
}

// Hidden or impossibly small taskbars never leave an interactive cover child.
func TestTaskbarLyricsCoverPlacementUnavailable(t *testing.T) {
	for _, p := range []Placement{{}, {Parent: 1, Width: 600, Height: 4, DPI: 96}, {Parent: 1, Width: 10, Height: 42, DPI: 96}} {
		if got := TaskbarCoverPlacement(p); got != (Placement{}) {
			t.Fatalf("unexpected interactive area: %+v", got)
		}
	}
}
