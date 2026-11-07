package appui

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// NowPlayingTransition is Now Playing as it grows out of the player bar's
// cover and shrinks back into it. Its fields are written once a build by
// the layer and read by everything the layer shows.
type NowPlayingTransition struct {
	T     float32 // how far open, 0 to 1
	Shift float32 // how far down while pulled
	Pull  float32 // how far the pointer has pulled it down
	Held  bool    // whether the pointer is pulling it
	Still bool    // reduced motion: fade in place
	From  ui.Rect // the player bar's cover, which it grows out of
}

// NPFromRadius is the radius of the player bar's cover it grows from.
const NPFromRadius = 8

// AtRest reports whether the last frame was all open without a pull, or
// all closed. The covered page can stop building only at rest.
func (s *NowPlayingTransition) AtRest(open bool) bool {
	if open {
		return s.T >= 1 && s.Shift == 0 && !s.Held
	}
	return s.T <= 0
}

// FromRect is where the bar last laid out its cover.
func (s *NowPlayingTransition) FromRect() ui.Rect {
	return s.From
}

// Morph moves a cover from the bar to its place as the card opens, and
// back to the bar as it closes.
func (s *NowPlayingTransition) Morph(c *ui.Context, e ui.Element) ui.Element {
	if s.T >= 1 || s.Still {
		return e
	}
	b := e.LayoutBounds()
	if b.W <= 0 {
		c.AnimationFrame()
		return e.Opacity(0)
	}
	from := s.FromRect()
	if from.W <= 0 {
		c.AnimationFrame()
		return e.Opacity(0)
	}
	k := 1 - s.T
	scale := 1 + (from.W/b.W-1)*k
	return e.TransformOrigin(0, 0).Scale(scale, scale).Translate((from.X-b.X)*k, (from.Y-b.Y)*k)
}

// Breathe blends the cover's resting scale into its flight, so even a
// paused cover meets the bar's cover at its full size.
func (s *NowPlayingTransition) Breathe(e ui.Element, scale float32) {
	if s.T < 1 && !s.Still {
		v := lerpNP(1, scale, clampNP01(s.T))
		e.Scale(v, v)
		return
	}
	e.AnimatePose(ui.Pose{Scale: scale}, ui.Animation{Spring: ui.SpringOptions{Duration: 600 * time.Millisecond, Bounce: 0.25}})
}

// Radius keeps the scaled cover's corners round as it meets the bar.
func (s *NowPlayingTransition) Radius(c *ui.Context, r, size float32) float32 {
	if s.T >= 1 || s.Still {
		return r
	}
	from := s.FromRect()
	if from.W <= 0 {
		return r
	}
	k := 1 - clampNP01(s.T)
	scale := 1 + (from.W/size-1)*k
	return min(lerpNP(r, NPFromRadius, k)/scale, size/2)
}

// Rise brings controls in after the cover, in order, and reverses that
// order as the card closes.
func (s *NowPlayingTransition) Rise(e ui.Element, n int) ui.Element {
	if s.T >= 1 || s.Still {
		return e
	}
	from := 0.3 + 0.07*float32(n)
	k := smoothstepNP(from, from+0.4, s.T)
	return e.Opacity(k).Translate(0, 24*(1-k))
}

func clampNP01(v float32) float32 { return min(1, max(0, v)) }

func lerpNP(a, b, t float32) float32 { return a + (b-a)*t }

func smoothstepNP(a, b, t float32) float32 {
	k := clampNP01((t - a) / (b - a))
	return k * k * (3 - 2*k)
}
