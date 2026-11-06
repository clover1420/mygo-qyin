package app

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// How the record comes in as the app starts.
const (
	splashTime = 1150 * time.Millisecond
	splashDisc = 168
)

// easeRange is how far v has gone from a to b, from 0 to 1.
func easeRange(v, a, b float32) float32 { return clamp01((v - a) / max(0.0001, b-a)) }

// splash is a record over the window as the app starts, spinning up out of
// nothing and fading into the app, whose first page is loading under it. It
// takes no clicks, for nothing to wait on it.
func (a *app) splash(c *ui.Context) {
	if a.splashed {
		return
	}
	// The record comes over the app's own window as it opens: a view built
	// without one has nothing to cover.
	if a.win == nil || c.Preferences().ReduceMotion {
		a.splashed = true
		return
	}
	now := c.Now()
	if a.splashAt.IsZero() {
		a.splashAt = now
	}
	t := float32(now.Sub(a.splashAt)) / float32(splashTime)
	if t >= 1 {
		a.splashed = true
		return
	}
	c.AnimationFrame()
	in := 1 - easeRange(1-t, 0.45, 1)
	gone := easeRange(t, 0.68, 1)
	s := lerp(0.72, 1, in) + 0.06*gone
	ui.Box(c.Key("splash")).Absolute().Top(0).Left(0).Right(0).Bottom(0).PassThrough().Center().
		Background(a.pal.Bg).Opacity(1 - gone).Children(func() {
		disc := ui.Box(c.Key("disc")).Size(splashDisc, splashDisc).Radius(999).Center().PassThrough().
			Background(ui.Hex("#232327")).Border(1, white.Alpha(0.14)).
			Shadow(0, 16, 44, -10, ui.RGBA(0, 0, 0, 0.5)).
			Opacity(easeRange(t, 0, 0.3)).Scale(s, s).Rotate(lerp(-150, 0, in))
		disc.Draw(func(p *ui.Painter, r ui.Rect) {
			// Its grooves, and the light across them.
			for i := float32(0); i < 5; i++ {
				inset := 8 + i*6
				p.Stroke(ui.Rect{X: r.X + inset, Y: r.Y + inset, W: r.W - 2*inset, H: r.H - 2*inset},
					white.Alpha(0.09), (r.W-2*inset)/2, 1)
			}
		})
		disc.Children(func() {
			// The label at its middle, in the app's red.
			ui.Box(c).Size(64, 64).Radius(999).Center().Background(a.pal.Accent).Children(func() {
				ui.Icon(c, iconNote).Size(26, 26).TextColor(a.pal.OnAccent)
			})
		})
	})
}
