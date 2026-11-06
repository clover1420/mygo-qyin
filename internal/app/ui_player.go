package app

import (
	"math"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"lilt/internal/player"
)

// volumeButton is the speaker, which opens the volume in a popover above
// it, as an AppKit popover: a native slider standing up, the volume in
// percent over it. The wheel over the speaker turns it up and down
// without opening it.
func (a *app) volumeButton(c *ui.Context, color ui.Color) {
	surface := playerSurface
	if a.volumeOpen {
		surface = playerSelectedSurface
	}
	b := iconButton(c.Key("volume-button"), volumeIcon(a.volume), 16, color, "音量", surface)
	if b.Clicked() {
		a.volumeOpen = !a.volumeOpen
	}
	b.HandleInput(func(ev ui.InputEvent) bool { return a.scrollVolume(ev) })
	pop := popoverColors(a.pal.Dark)
	ui.PopoverBase(c.Key("volume-popover"), b, &a.volumeOpen, func(panel ui.Element) {
		const arrow = 7
		panel.AttachTo(b, ui.AnchorTop, ui.AnchorBottom).Margin(0, 0, arrow+3, 0).Width(52).Padding(11, 0, 12, 0).
			Radius(14).Background(pop.fill).Border(0.5, pop.edge).Shadow(0, 8, 24, -4, ink.Alpha(0.2)).
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
		ui.Textf(c, "%d%%", int(a.volume*100+0.5)).FontSize(11).FontWeight(500).FontFeatures("tnum").TextColor(a.pal.Muted)
		a.nativeVolumeSlider(c, pop)
	})
}

// popover holds the colors of a popover as AppKit draws them.
type popover struct {
	fill, edge, track, knob ui.Color
}

func popoverColors(dark bool) popover {
	return popover{fill: neutral(dark, 2), edge: neutral(dark, 6), track: neutral(dark, 5), knob: paper}
}

// nativeVolumeSlider is a vertical slider as AppKit draws one: a thin
// track, filled with the accent from the foot to a round knob.
func (a *app) nativeVolumeSlider(c *ui.Context, pop popover) {
	const knob, track = 20, 4
	s := ui.SliderBase(c.Key("volume"), &a.volume, 0, 1).Vertical().Size(knob+8, 124).PaddingY(knob / 2).
		Step(0.01).AutoFocus().FocusRing(false).Label("音量")
	if s.Changed() {
		a.pl.SetVolume(a.volume)
	}
	s.HandleInput(func(ev ui.InputEvent) bool { return a.scrollVolume(ev) })
	pressed := s.Pressed()
	v := float32(a.volume)
	accent := a.pal.Accent
	s.Draw(func(p *ui.Painter, r ui.Rect) {
		top, foot := r.Y+knob/2, r.Y+r.H-knob/2
		x := r.X + (r.W-track)/2
		y := foot - (foot-top)*v
		p.Fill(ui.Rect{X: x, Y: top - track/2, W: track, H: foot - top + track}, pop.track, track/2)
		p.Fill(ui.Rect{X: x, Y: y, W: track, H: foot - y + track/2}, accent, track/2)
		k := ui.Rect{X: r.X + (r.W-knob)/2, Y: y - knob/2, W: knob, H: knob}
		p.Shadow(k, knob/2, 0, 0.5, 1.5, 0, ink.Alpha(0.3))
		p.Shadow(k, knob/2, 0, 1, 4, 0, ink.Alpha(0.12))
		fill := pop.knob
		if pressed {
			fill = fill.Mix(ink, 0.06)
		}
		p.Fill(k, fill, knob/2)
		p.Stroke(k, ink.Alpha(0.08), knob/2, 0.5)
	})
}

// scrollVolume turns the volume up as the wheel or the fingers scroll up.
func (a *app) scrollVolume(ev ui.InputEvent) bool {
	if ev.Kind != ui.InputScroll || ev.DY == 0 {
		return false
	}
	delta := -float64(ev.DY) / 400
	if !ev.Precise {
		// Wheel distances reflect the platform and system scroll settings;
		// use a small volume step rather than a fraction of that distance.
		delta = math.Copysign(0.05, delta)
	}
	a.volume = min(1, max(0, a.volume+delta))
	a.pl.SetVolume(a.volume)
	return true
}

// volumeIcon is the speaker for volume v: silent, or with one or two waves.
func volumeIcon(v float64) *ui.SVG {
	switch {
	case v == 0:
		return iconMute
	case v < 0.4:
		return iconVolumeLow
	default:
		return iconVolume
	}
}

// muteButton is the speaker, which mutes and unmutes.
func (a *app) muteButton(c *ui.Context, color ui.Color, surface ...buttonSurface) {
	icon := iconVolume
	if a.volume == 0 {
		icon = iconMute
	}
	if iconButton(c.Key("mute"), icon, 16, color, "静音", surface...).Clicked() {
		if a.volume > 0 {
			a.mutedFrom, a.volume = a.volume, 0
		} else {
			a.volume = max(a.mutedFrom, 0.5)
		}
		a.pl.SetVolume(a.volume)
	}
}

// volumeSlider is a slim slider with the speaker before it, which mutes.
func (a *app) volumeSlider(c *ui.Context, color, track ui.Color, width float32, surface ...buttonSurface) {
	ui.Row(c).Gap(4).AlignItems(ui.Center).Children(func() {
		a.muteButton(c, color, surface...)
		s := ui.SliderBase(c.Key("volume"), &a.volume, 0, 1).Width(width).Height(16).FocusRing(false).Label("音量").Cursor(ui.CursorPointer)
		if s.Changed() {
			a.pl.SetVolume(a.volume)
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
		v := float32(a.volume)
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

// transport is the row of controls: shuffle, previous, play, next, repeat.
func (a *app) transport(c *ui.Context, color, dim, on ui.Color, size float32, surface ...buttonSurface) {
	pl := a.pl
	ui.Row(c).Gap(size * 0.5).AlignItems(ui.Center).Justify(ui.Center).Children(func() {
		shuffle := dim
		if pl.Shuffle() {
			shuffle = on
		}
		if iconButton(c.Key("shuffle"), iconShuffle, size*0.62, shuffle, "随机播放", surface...).Clicked() {
			pl.ToggleShuffle()
		}
		if iconButton(c.Key("prev"), iconPrev, size*0.85, color, "上一首", surface...).Clicked() {
			pl.Prev()
		}
		if playButton(c.Key("play"), pl.Playing(), size*1.25, color, surface...).Clicked() {
			pl.TogglePlay()
		}
		if iconButton(c.Key("next"), iconNext, size*0.85, color, "下一首", surface...).Clicked() {
			pl.Next()
		}
		repeat, ricon := dim, iconRepeat
		switch pl.Repeat() {
		case player.RepeatAll:
			repeat = on
		case player.RepeatOne:
			repeat, ricon = on, iconRepeat1
		}
		if iconButton(c.Key("repeat"), ricon, size*0.62, repeat, "循环", surface...).Clicked() {
			pl.CycleRepeat()
		}
	})
}

// playerBar runs along the foot of the content: what plays, its controls
// and its progress, and the volume.
func (a *app) playerBar(c *ui.Context) {
	p := a.pal
	pl := a.pl
	tr := pl.Current()
	a.tickTimes(c)
	if a.picked.id != 0 && (tr == nil || tr.ID != a.picked.id || pl.Playing()) {
		// Played, or gone: it says what it is again.
		a.picked = pickedUp{}
	}
	liquid := a.settings.LiquidGlass
	// Over the foot of the page: while content is under it, its edge frosts
	// it, with a hairline along the top. Glass floats clear of the foot
	// instead, a capsule over what the page scrolls under it.
	glassH := float32(playerBarHeight - glassBarFoot)
	if !liquid {
		edge := ui.Box(c.Key("bottom-edge")).Absolute().Left(0).Right(0).Bottom(0).Height(playerBarHeight).
			Material(glass.ScrollEdge{Hard: true, Bottom: true, Background: p.Bg})
		edge.Opacity(edgeShown(edge, a.underBottom))
	}
	// The controls keep the middle, between sides that share what is left:
	// narrower, the middle gives up to 140 DIPs, and what plays keeps room
	// for its cover, some of its title and its heart.
	controls := float32(2*32 + 4) // queue and the speaker
	if supportsDesktopLyrics() {
		controls += 32 + 4
	}
	bar := ui.Row(c).Absolute().Gap(20).AlignItems(ui.Center)
	if liquid {
		bar.Left(glassBarSide).Right(glassBarSide).Bottom(glassBarFoot).Height(glassH).PaddingX(20).
			Radius(glassH/2).Shadow(0, 6, 24, -6, ui.RGBA(0, 0, 0, 0.3)).
			Material(barGlass{dark: p.Dark, clear: float32(a.settings.GlassClear), bend: float32(a.settings.GlassBend)})
	} else {
		bar.Left(0).Right(0).Bottom(0).Height(playerBarHeight).PaddingX(20)
	}
	bar.Children(func() {
		// What plays.
		ui.Row(c).Grow(1).Basis(0).MinWidth(150).Children(func() {
			if tr == nil {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(48, 48).Radius(8).Background(p.Hover).Center().Children(func() {
						ui.Icon(c, iconNote).Size(18, 18).TextColor(p.Faint)
					})
					ui.Text(c, "未在播放").FontSize(13).TextColor(p.Muted)
				})
				return
			}
			info := ui.ButtonBase(c.Key("info")).Gap(12).AlignItems(ui.Center).FocusRing(false).Shrink(1).MinWidth(0).Label("打开正在播放")
			hovered := info.Hovered() || info.FocusVisible()
			motion := ui.Animation{Duration: 160 * time.Millisecond}
			reveal := info.AnimateTo("reveal", b2f(hovered), motion)
			info.Children(func() {
				cover := ui.Box(c).Size(48, 48).Shrink(0)
				cover.Translate(0, -reveal).Children(func() {
					a.cover(c, tr.Album.PicURL, 120, 48, 8).Shadow(0, 2, 8, 0, p.CardShade)
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(8).Center().PassThrough().
						Opacity(reveal).Background(shade(5)).TextColor(paper).Children(func() {
						ui.Icon(c, iconDown).Size(18, 18).Rotate(180).Translate(0, 3*(1-reveal))
					})
				})
				ui.Column(c).Shrink(1).MinWidth(0).Gap(3).Children(func() {
					ui.Text(c, tr.Name).FontSize(13.5).FontWeight(600).SingleLine()
					// What it is waiting on shows in place of the artists.
					var sub string
					color := p.Muted
					switch {
					case a.picked.id == tr.ID:
						sub = a.picked.heard
					case pl.Loading():
						sub = "正在缓冲…"
					case pl.Err() != "":
						sub, color = pl.Err(), p.Accent
					}
					if sub == "" {
						a.artistLinks(c, tr, p.Muted, p.Text).FontSize(12).SingleLine()
					} else {
						ui.Text(c, sub).FontSize(12).TextColor(color).SingleLine()
					}
				})
			})
			if info.Clicked() {
				a.nowPlaying = true
			}
			a.likeButton(c.Key("like"), tr, 16, p.Muted, playerSurface).Margin(0, 0, 0, 4).Shrink(0)
		})

		// Controls over the progress.
		ui.Column(c).Width(440).Shrink(1).MinWidth(300).Gap(0).AlignItems(ui.Stretch).Children(func() {
			a.transport(c, p.Text, p.Muted, p.Accent, 18, playerSurface)
			ui.Row(c).Gap(10).AlignItems(ui.Center).Margin(-4, 0, 0, 0).Children(func() {
				ui.Text(c, clock(pl.Position())).FontSize(10.5).FontFeatures("tnum").TextColor(p.Faint).Width(34).TextAlign(ui.End)
				a.scrubber(c.Key("bar-scrub"), &a.barScrub, p.Press, a.progressFill(p), 14, p.Accent).Grow(1)
				ui.Text(c, clock(pl.Duration())).FontSize(10.5).FontFeatures("tnum").TextColor(p.Faint).Width(34)
			})
		})

		// Queue, volume and the switch for the selected external lyrics mode.
		ui.Row(c).Grow(1).Basis(0).MinWidth(controls).Justify(ui.End).AlignItems(ui.Center).Gap(4).Children(func() {
			a.queueButton(c)
			a.volumeButton(c, p.Text.Alpha(0.75))
			if supportsDesktopLyrics() {
				a.externalLyricsPlayerButton(c, p)
			}
		})
	})
}

// Centered lyric lines identify the floating caption window. Open corners
// leave it free to move; closing the same frame marks its fixed position.
const desktopLyricsGlyph = `<path d="M7 10h10M9 14h6"/>`

var (
	iconDesktopLyrics       = strokeIcon(desktopLyricsGlyph + `<path d="M8 4H6a3 3 0 0 0-3 3v2M16 4h2a3 3 0 0 1 3 3v2M21 15v2a3 3 0 0 1-3 3h-2M8 20H6a3 3 0 0 1-3-3v-2"/>`)
	iconDesktopLyricsLocked = strokeIcon(desktopLyricsGlyph + `<rect x="3" y="4" width="18" height="16" rx="3"/>`)
	iconTaskbarLyrics       = strokeIcon(`<rect x="3" y="4" width="18" height="16" rx="3"/><path d="M3 14h18M7 17h10"/>`)
)

func (a *app) externalLyricsPlayerButton(c *ui.Context, p palette) {
	mode := lyricsMode(a.settings)
	label, icon := "桌面歌词", iconDesktopLyrics
	if mode == lyricsTaskbar {
		label, icon = "任务栏歌词", iconTaskbarLyrics
	}
	color, hint := p.Muted, "开启"+label
	if lyricsEnabled(a.settings) {
		color, hint = p.Accent, "关闭"+label
		if mode == lyricsDesktop {
			hint += " · 右键锁定"
			if a.settings.DesktopLyricsLocked {
				hint, icon = "关闭桌面歌词 · 右键解锁", iconDesktopLyricsLocked
			}
		}
	}
	button := iconButton(c.Key("external-lyrics"), icon, 18, color, label, playerSurface).Size(32, 32).Tooltip(hint)
	button.OnClick(func() { a.setLyricsEnabled(!lyricsEnabled(a.settings)) })
	if button.RightClicked() && mode == lyricsDesktop && a.settings.DesktopLyrics {
		a.setDesktopLyricsLocked(!a.settings.DesktopLyricsLocked)
	}
}
