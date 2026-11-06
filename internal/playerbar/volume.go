package playerbar

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/player"
	"lilt/internal/theme"
)

// VolumeButton is the speaker, which opens the volume in a popover above
// it, as an AppKit popover: a native slider standing up, the volume in
// percent over it. The wheel over the speaker turns it up and down
// without opening it.
func VolumeButton(c *ui.Context, volume *float64, open *bool, pl *player.Player, p theme.Palette, color ui.Color) {
	surface := appui.PlayerSurface
	if *open {
		surface = appui.PlayerSelectedSurface
	}
	b := appui.IconButton(c.Key("volume-button"), volumeIcon(*volume), 16, color, "音量", surface)
	if b.Clicked() {
		*open = !*open
	}
	b.HandleInput(func(ev ui.InputEvent) bool { return ScrollVolume(volume, pl, ev) })
	pop := PopoverColors(p.Dark)
	ui.PopoverBase(c.Key("volume-popover"), b, open, func(panel ui.Element) {
		const arrow = 7
		panel.AttachTo(b, ui.AnchorTop, ui.AnchorBottom).Margin(0, 0, arrow+3, 0).Width(52).Padding(11, 0, 12, 0).
			Radius(14).Background(pop.fill).Border(0.5, pop.edge).Shadow(0, 8, 24, -4, theme.Ink.Alpha(0.2)).
			AlignItems(ui.Center).Gap(8).Label("音量")
		panel.Draw(func(p *ui.Painter, r ui.Rect) {
			// The arrow pointing at the speaker, its sides edged as the
			// panel is, over the panel's border.
			cx, foot := r.X+r.W/2, r.Y+r.H
			tip := (&ui.Path{}).MoveTo(cx-arrow-1, foot-0.75).LineTo(cx-1.2, foot+arrow-1.2).
				QuadTo(cx, foot+arrow, cx+1.2, foot+arrow-1.2).LineTo(cx+arrow+1, foot-0.75).Close()
			p.FillPath(tip, pop.fill)
			edge := (&ui.Path{}).MoveTo(cx-arrow-1, foot).LineTo(cx-1.2, foot+arrow-1.2).
				QuadTo(cx, foot+arrow, cx+1.2, foot+arrow-1.2).LineTo(cx+arrow+1, foot)
			p.StrokePath(edge, 0.5, pop.edge)
		})
		ui.Textf(c, "%d%%", int(*volume*100+0.5)).FontSize(11).FontWeight(500).FontFeatures("tnum").TextColor(p.Muted)
		NativeVolumeSlider(c, volume, pl, pop, p.Accent)
	})
}

// volumeButton is VolumeButton, with the bar's state.
func volumeButton(c *ui.Context, d Deps, color ui.Color) {
	VolumeButton(c, d.Volume, d.VolumeOpen, d.Player, d.Palette, color)
}

// Popover holds the colors of a popover as AppKit draws them.
type Popover struct {
	fill, edge, track, knob ui.Color
}

// PopoverColors returns the colors of a popover.
func PopoverColors(dark bool) Popover {
	return Popover{fill: theme.Neutral(dark, 2), edge: theme.Neutral(dark, 6), track: theme.Neutral(dark, 5), knob: theme.Paper}
}

// NativeVolumeSlider is a vertical slider as AppKit draws one: a thin
// track, filled with the accent from the foot to a round knob.
func NativeVolumeSlider(c *ui.Context, volume *float64, pl *player.Player, pop Popover, accent ui.Color) {
	const knob, track = 20, 4
	s := ui.SliderBase(c.Key("volume"), volume, 0, 1).Vertical().Size(knob+8, 124).PaddingY(knob / 2).
		Step(0.01).AutoFocus().FocusRing(false).Label("音量")
	if s.Changed() {
		pl.SetVolume(*volume)
	}
	s.HandleInput(func(ev ui.InputEvent) bool { return ScrollVolume(volume, pl, ev) })
	pressed := s.Pressed()
	v := float32(*volume)
	s.Draw(func(p *ui.Painter, r ui.Rect) {
		top, foot := r.Y+knob/2, r.Y+r.H-knob/2
		x := r.X + (r.W-track)/2
		y := foot - (foot-top)*v
		p.Fill(ui.Rect{X: x, Y: top - track/2, W: track, H: foot - top + track}, pop.track, track/2)
		p.Fill(ui.Rect{X: x, Y: y, W: track, H: foot - y + track/2}, accent, track/2)
		k := ui.Rect{X: r.X + (r.W-knob)/2, Y: y - knob/2, W: knob, H: knob}
		p.Shadow(k, knob/2, 0, 0.5, 1.5, 0, theme.Ink.Alpha(0.3))
		p.Shadow(k, knob/2, 0, 1, 4, 0, theme.Ink.Alpha(0.12))
		fill := pop.knob
		if pressed {
			fill = fill.Mix(theme.Ink, 0.06)
		}
		p.Fill(k, fill, knob/2)
		p.Stroke(k, theme.Ink.Alpha(0.08), knob/2, 0.5)
	})
}

// ScrollVolume turns the volume up as the wheel or the fingers scroll up.
func ScrollVolume(volume *float64, pl *player.Player, ev ui.InputEvent) bool {
	if ev.Kind != ui.InputScroll || ev.DY == 0 {
		return false
	}
	delta := -float64(ev.DY) / 400
	if !ev.Precise {
		// Wheel distances reflect the platform and system scroll settings;
		// use a small volume step rather than a fraction of that distance.
		delta = math.Copysign(0.05, delta)
	}
	*volume = min(1, max(0, *volume+delta))
	pl.SetVolume(*volume)
	return true
}

// volumeIcon is the speaker for volume v: silent, or with one or two waves.
func volumeIcon(v float64) *ui.SVG {
	switch {
	case v == 0:
		return appui.IconMute
	case v < 0.4:
		return appui.IconVolumeLow
	default:
		return appui.IconVolume
	}
}

// muteButton is the speaker, which mutes and unmutes.
func muteButton(c *ui.Context, volume, mutedFrom *float64, pl *player.Player, color ui.Color, surface ...appui.ButtonSurface) {
	icon := appui.IconVolume
	if *volume == 0 {
		icon = appui.IconMute
	}
	if appui.IconButton(c.Key("mute"), icon, 16, color, "静音", surface...).Clicked() {
		if *volume > 0 {
			*mutedFrom, *volume = *volume, 0
		} else {
			*volume = max(*mutedFrom, 0.5)
		}
		pl.SetVolume(*volume)
	}
}

// VolumeSlider is a slim slider with the speaker before it, which mutes.
func VolumeSlider(c *ui.Context, volume, mutedFrom *float64, pl *player.Player, color, track ui.Color, width float32, surface ...appui.ButtonSurface) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		muteButton(c, volume, mutedFrom, pl, color, surface...)
		s := ui.SliderBase(c.Key("volume"), volume, 0, 1).Width(width).Height(16).FocusRing(false).Label("音量").Cursor(ui.CursorPointer)
		if s.Changed() {
			pl.SetVolume(*volume)
		}
		hot := s.Hovered() || s.Pressed() || s.FocusVisible()
		thick, knob, fill := float32(4), float32(0), color
		if hot {
			thick, knob, fill = 5, 1, color.Alpha(1)
		}
		motion := ui.Animation{Duration: 140 * time.Millisecond}
		thick = s.AnimateTo("thick", thick, motion)
		knob = s.AnimateTo("knob", knob, motion)
		fill = s.AnimateColor("fill", fill, motion)
		v := float32(*volume)
		s.Draw(func(p *ui.Painter, r ui.Rect) {
			y := r.Y + (r.H-thick)/2
			p.Fill(ui.Rect{X: r.X, Y: y, W: r.W, H: thick}, track, thick/2)
			p.Fill(ui.Rect{X: r.X, Y: y, W: max(thick, r.W*v), H: thick}, fill, thick/2)
			if knob > 0.01 {
				k := 10 * knob
				x := r.X + r.W*v
				p.Fill(ui.Rect{X: x - k/2, Y: r.Y + (r.H-k)/2, W: k, H: k}, fill, k/2)
			}
		})
	})
}

// Transport is the row of controls: shuffle, previous, play, next, repeat.
func Transport(c *ui.Context, pl *player.Player, color, dim, on ui.Color, size float32, surface ...appui.ButtonSurface) {
	ui.Row(c).Gap(size * 0.5).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		shuffle := dim
		if pl.Shuffle() {
			shuffle = on
		}
		if appui.IconButton(c.Key("shuffle"), appui.IconShuffle, size*0.62, shuffle, "随机播放", surface...).Clicked() {
			pl.ToggleShuffle()
		}
		if appui.IconButton(c.Key("prev"), appui.IconPrev, size*0.85, color, "上一首", surface...).Clicked() {
			pl.Prev()
		}
		if appui.PlayButton(c.Key("play"), pl.Playing(), size*1.25, color, surface...).Clicked() {
			pl.TogglePlay()
		}
		if appui.IconButton(c.Key("next"), appui.IconNext, size*0.85, color, "下一首", surface...).Clicked() {
			pl.Next()
		}
		repeat, ricon := dim, appui.IconRepeat
		switch pl.Repeat() {
		case player.RepeatAll:
			repeat = on
		case player.RepeatOne:
			repeat, ricon = on, appui.IconRepeat1
		}
		if appui.IconButton(c.Key("repeat"), ricon, size*0.62, repeat, "循环", surface...).Clicked() {
			pl.CycleRepeat()
		}
	})
}
