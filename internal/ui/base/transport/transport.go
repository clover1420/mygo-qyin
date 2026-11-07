// Package transport holds the playback controls views share: the progress
// bar, its smoke and wave, and the equalizer bars.
package transport

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/settings"
	"lilt/internal/media/player"
	"lilt/internal/ui/base/theme"
)

// State is a progress bar's own: the value dragged, while it is, and the
// levels of the music its smoke follows.
type State struct {
	V float64
	// Held is set while the pointer pressed on the bar is down, wherever
	// it has gone since: Pressed is false off the bar, which is thin.
	Held   bool
	Active bool // dragged since the last frame, to seek once let go
	Level  Bar
}

func Secs(s float64) time.Duration { return time.Duration(s * float64(time.Second)) }

// Scrubber shows how far the track is, and seeks where it is dragged or
// clicked: the bar thickens under the pointer, and its knob shows. With a
// smoke color, the bar is a comet while music plays.
func Scrubber(c *ui.Context, pl *player.Player, effect string, st *State, track, fill ui.Color, height float32, smoke ui.Color) ui.Element {
	dur := pl.Duration().Seconds()
	if !st.Active && !st.Held {
		st.V = pl.Position().Seconds()
	}
	s := ui.SliderBase(c, &st.V, 0, max(dur, 1)).Height(height).FocusRing(false).Label("播放进度").Cursor(ui.CursorPointer)
	// The press and its release come to the bar wherever the pointer is,
	// and go on to the slider, which follows the pointer.
	s.HandleInput(func(ev ui.InputEvent) bool {
		switch ev.Kind {
		case ui.InputPointerDown:
			st.Held = ev.Button == 0
		case ui.InputPointerUp:
			st.Held = false
		}
		return false
	})
	pressed := st.Held
	switch {
	case s.PressCanceled():
		// Losing capture ends the preview without committing a seek.
		st.Held, st.Active = false, false
		st.V = pl.Position().Seconds()
	case pressed:
		st.Active = true
	case st.Active:
		// Let go: there, but not right at the end, which would end the
		// song as it begins to play.
		st.Active = false
		pl.Seek(Secs(min(st.V, max(0, dur-1))))
	case s.Changed():
		pl.Seek(Secs(min(st.V, max(0, dur-1))))
	}
	wavy := effect == settings.ProgressSquiggly
	hot := s.Hovered() || pressed || s.FocusVisible()
	thick := float32(4)
	if hot {
		thick = 6
	}
	thick = s.AnimateTo("thick", thick, ui.Snappy)
	knob := float32(0)
	// The wave keeps its playhead visible even without hover or playback.
	if wavy {
		knob = 2.0 / 3 // 8 DIPs, growing to 12 under the pointer.
	}
	if hot {
		knob = 1
	}
	knob = s.AnimateTo("knob", knob, ui.Snappy)
	waveHeight := float32(0)
	waveAnimation := ui.Animation{Duration: 550 * time.Millisecond}
	if wavy && pl.Playing() && !pl.Loading() && !c.Preferences().ReduceMotion {
		waveHeight = 1
		waveAnimation.Duration = 800 * time.Millisecond
		waveAnimation.Delay = 60 * time.Millisecond
	}
	waveHeight = s.AnimateTo("wave-height", waveHeight, waveAnimation)
	// The smoke settles away as the music pauses, and while dragging.
	smoking := float32(0)
	if smoke.A > 0 && effect == settings.ProgressSmoke && pl.Playing() && !pl.Loading() && !st.Active {
		smoking = 1
	}
	smoking = s.AnimateTo("smoke", smoking, ui.Animation{Duration: 600 * time.Millisecond})
	s.Draw(func(p *ui.Painter, r ui.Rect) {
		at := st.V
		if !st.Active {
			at = pl.Position().Seconds()
		}
		frac := float32(0)
		if dur > 0 {
			frac = float32(min(1, max(0, at/dur)))
		}
		y := r.Y + (r.H-thick)/2
		head, cy := r.X+r.W*frac, r.Y+r.H/2
		if wavy {
			DrawSquiggly(p, r, head, thick, waveHeight, track, fill)
		} else {
			p.Fill(ui.Rect{X: r.X, Y: y, W: r.W, H: thick}, track, thick/2)
			// What has downloaded, a shade lighter.
			if got, ok := pl.Downloaded(); ok && float32(got) < 1 {
				p.Fill(ui.Rect{X: r.X, Y: y, W: r.W * float32(got), H: thick}, track, thick/2)
			}
			played := ui.Rect{X: r.X, Y: y, W: max(thick, head-r.X), H: thick}
			if smoking > 0.01 {
				// A streak of light, brightest at the head.
				p.FillGradient(played, ui.LinearGradient{Angle: 90, Oklab: true,
					From: fill.Mix(smoke, smoking).Alpha(0.3), To: fill.Mix(smoke, 0.8*smoking)}, thick/2)
				lv, live := pl.Levels()
				st.Level.Step(p.Now(), lv, live)
				bass := st.Level.Group(0, 6)
				DrawSmoke(p, head, cy, p.Now(), bass, smoke, smoking)
				// The glow of the head, swelling with the bass, and its core.
				glow := 9 + 8*bass
				p.Shadow(ui.Rect{X: head, Y: cy}, 0, 0, 0, glow, 2.5+2.5*bass, smoke.Alpha(0.95*smoking))
				core := thick * (1 + 0.25*bass)
				p.Fill(ui.Rect{X: head - core/2, Y: cy - core/2, W: core, H: core}, theme.Paper.Mix(smoke, 0.15).Alpha(smoking), core/2)
			} else {
				p.Fill(played, fill, thick/2)
			}
		}
		if knob > 0.01 {
			k := 12 * knob
			p.Shadow(ui.Rect{X: head - k/2, Y: r.Y + (r.H-k)/2, W: k, H: k}, k/2, 0, 1, 3, 0, theme.Ink.Alpha(0.3))
			p.Fill(ui.Rect{X: head - k/2, Y: r.Y + (r.H-k)/2, W: k, H: k}, fill, k/2)
		}
		switch {
		case wavy && waveHeight > 0.01 && frac > 0:
			p.AnimationFrameRate(30)
		case smoking > 0.01 && (pl.Playing() || smoking < 1):
			// The smoke drifts slowly, under a point a frame at 30 frames
			// a second, which show it smooth for a quarter of what a 120 Hz
			// display would draw.
			p.After(time.Second / 30)
		case pl.Playing() && !st.Active && dur > 0:
			// Again when the fill has moved a pixel of the screen, which
			// takes a good part of a second, rather than every frame.
			perPixel := dur / float64(r.W*p.Scale())
			p.After(max(16*time.Millisecond, Secs(perPixel)))
		}
	})
	return s
}

// Smoke from the playhead: wisps shed one after another, each drifting
// back along the bar and rising, growing and fading over its life. Where a
// wisp is follows from the time alone, so nothing is kept between frames.
const (
	smokeLife  = 1.4 // seconds
	smokeWisps = 36  // alive at once, at most
)

// DrawSmoke draws the smoke of a playhead at (x, y) at now, thicker as
// loud (0 to 1) is higher, at opacity shown.
func DrawSmoke(p *ui.Painter, x, y float32, now time.Time, loud float32, color ui.Color, shown float32) {
	t := float64(now.UnixMilli()) / 1000
	interval := smokeLife / smokeWisps
	first := math.Floor(t / interval)
	for j := range smokeWisps {
		k := first - float64(j)
		age := float32(t - k*interval)
		if age < 0 || age > smokeLife {
			continue
		}
		// A wisp of its own: how fast it drifts and rises, how it curls,
		// how large it grows. Quiet music sheds fewer of them.
		r := func(o float64) float32 { return LineRandom(k, o) }
		if r(5) > 0.6+0.4*loud {
			continue
		}
		f := age / smokeLife
		wx := x - age*(12+22*r(1))
		wy := y - age*(12+18*r(2)) + float32(math.Sin(float64(age)*(2.5+2*float64(r(3)))+float64(r(4))*2*math.Pi))*6*f
		size := 3 + f*(12+10*r(4))
		alpha := 0.5 * float32(math.Pow(float64(1-f), 1.4)) * shown * (0.75 + 0.5*loud)
		p.Shadow(ui.Rect{X: wx, Y: wy}, 0, 0, 0, size, size*0.45, color.Alpha(alpha))
	}
}

// TickTimes redraws the view as the next second of the track begins, for
// the times shown.
func TickTimes(c *ui.Context, pl *player.Player) {
	if pl.Playing() && pl.Ready() {
		pos := pl.Position()
		c.After(time.Second - pos%time.Second + 10*time.Millisecond)
	}
}
