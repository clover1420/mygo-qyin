package app

import (
	"testing"

	"github.com/egoist/mygo/ui"
)

func TestAccentCoverDefaultsOff(t *testing.T) {
	if defaultSettings().AccentCover {
		t.Fatal("the cover accent should be off by default")
	}
}

func TestCoverAccent(t *testing.T) {
	for _, tc := range []struct {
		name string
		tint ui.Color
		want bool
	}{
		{"vivid red", ui.RGB(220, 40, 30), true},
		{"vivid green", ui.RGB(30, 180, 70), true},
		{"gray cover", ui.RGB(120, 120, 120), false},
		{"black cover", ui.RGB(0, 0, 0), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := newTestApp()
			a.coverArt = &artwork{tint: tc.tint}
			got, ok := a.coverAccent()
			if ok != tc.want {
				t.Fatalf("follows cover = %v, want %v", ok, tc.want)
			}
			if ok && got.A == 0 {
				t.Fatal("the accent is transparent")
			}
		})
	}
	a := newTestApp()
	if _, ok := a.coverAccent(); ok {
		t.Fatal("a missing cover should not set an accent")
	}
}

// The theme keeps the fixed red while off and falls back to it when a cover
// gives no usable color; only a vivid cover with the setting on changes it.
func TestApplyThemeAccentCover(t *testing.T) {
	accentOf := func(follow bool, cover *artwork) ui.Color {
		a := newTestApp()
		a.settings = defaultSettings()
		a.settings.AccentCover = follow
		a.coverArt = cover
		var got ui.Color
		tt := ui.NewTester(func(c *ui.Context) {
			a.applyTheme(c)
			got = a.pal.accent
		}, 400, 300)
		tt.Frame()
		return got
	}
	red := accentBase(false)
	same := func(x, y ui.Color) bool {
		return x.R == y.R && x.G == y.G && x.B == y.B && x.A == y.A
	}
	green := &artwork{tint: ui.RGB(30, 180, 70)}
	if off := accentOf(false, green); !same(off, red) {
		t.Fatalf("off: %+v, want the fixed red %+v", off, red)
	}
	if on := accentOf(true, green); same(on, red) {
		t.Fatalf("on: %+v still the fixed red", on)
	}
	if missing := accentOf(true, nil); !same(missing, red) {
		t.Fatalf("without a cover: %+v, want the fixed red %+v", missing, red)
	}
	if gray := accentOf(true, &artwork{tint: ui.RGB(120, 120, 120)}); !same(gray, red) {
		t.Fatalf("gray cover: %+v, want the fixed red %+v", gray, red)
	}
}

// The played stretch of the player bar takes the cover's accent while the
// setting is on, and keeps its neutral color when it is off or the cover is
// too gray to follow.
func TestProgressFillAccentCover(t *testing.T) {
	p := palette{text: ui.RGB(30, 30, 32)}
	neutral := p.text.Alpha(0.75)
	same := func(x, y ui.Color) bool {
		return x.R == y.R && x.G == y.G && x.B == y.B && x.A == y.A
	}
	a := newTestApp()
	a.settings = defaultSettings()
	a.coverArt = &artwork{tint: ui.RGB(30, 180, 70)}
	if got := a.progressFill(p); !same(got, neutral) {
		t.Fatalf("off: %+v, want neutral %+v", got, neutral)
	}
	a.settings.AccentCover = true
	want, ok := a.coverAccent()
	if !ok {
		t.Fatal("the cover should give an accent")
	}
	if got := a.progressFill(p); !same(got, want) {
		t.Fatalf("on: %+v, want the cover accent %+v", got, want)
	}
	a.coverArt = &artwork{tint: ui.RGB(120, 120, 120)}
	if got := a.progressFill(p); !same(got, neutral) {
		t.Fatalf("gray cover: %+v, want neutral %+v", got, neutral)
	}
}
