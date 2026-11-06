// Package appui holds the controls the app's views share, without the app's
// state: buttons, media feedback, segmented controls and formatting.
package appui

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/theme"
)

// IconButton is a borderless button showing an icon, which a soft circle
// marks under the pointer and which gives a little as it is pressed.
func IconButton(c *ui.Context, icon *ui.SVG, size float32, color ui.Color, label string, surface ...ButtonSurface) ui.Element {
	b := RoundButton(c, size, color, label, surface...)
	b.Children(func() { ui.Icon(c, icon).Size(size, size) })
	return b
}

// PlayButton is the icon button that plays and pauses, its icon turning
// from one into the other.
func PlayButton(c *ui.Context, playing bool, size float32, color ui.Color, surface ...ButtonSurface) ui.Element {
	b := RoundButton(c, size, color, "播放/暂停", surface...)
	b.Children(func() { PlayPause(c, b, playing, size) })
	return b
}

// PlayPause draws the icon of b, a button that plays and pauses: as it
// goes from one to the other, the icon going shrinks and fades as the one
// coming grows in, with a little spring.
func PlayPause(c *ui.Context, b ui.Element, playing bool, size float32) {
	to := float32(0)
	if playing {
		to = 1
	}
	v := b.AnimateTo("playing", to, ui.Animation{Spring: ui.SpringOptions{Duration: 340 * time.Millisecond, Bounce: 0.3}})
	ui.Box(c).Size(size, size).Children(func() {
		for _, icon := range []struct {
			svg   *ui.SVG
			shown float32
		}{{IconPlay, 1 - v}, {IconPause, v}} {
			if icon.shown <= 0.01 {
				continue
			}
			s := 0.4 + 0.6*icon.shown
			ui.Icon(c, icon.svg).Absolute().Top(0).Left(0).Size(size, size).Scale(s, s).Opacity(min(icon.shown, 1))
		}
	})
}

// ButtonSurface keeps media controls legible over artwork, independently
// of the appearance of the pages underneath them.
type ButtonSurface uint8

const (
	MediaSurface ButtonSurface = iota + 1
	PlayerSurface
	PlayerSelectedSurface
)

// MediaFeedback gives controls a translucent hover and press over artwork.
// Brightening alpha preserves the hue of liked hearts and active modes.
func MediaFeedback(b ui.Element, color ui.Color) ui.Element {
	bg, fg := ui.Transparent, color
	pose := ui.Pose{}
	if b.Hovered() || b.FocusVisible() {
		bg, fg = theme.Veil(2), color.Alpha(1)
	}
	if b.Pressed() {
		bg, fg, pose.Scale = theme.Veil(3), color.Alpha(1), 0.94
	}
	motion := ui.Animation{Duration: 140 * time.Millisecond}
	return b.Background(b.AnimateColor("bg", bg, motion)).
		TextColor(b.AnimateColor("fg", fg, motion)).AnimatePose(pose, motion)
}

// PlayerFeedback layers neutral gray over the frosted bar so its underlying
// tint stays visible. Playback state colors stay on their icons.
func PlayerFeedback(c *ui.Context, b ui.Element, color ui.Color, surface ButtonSurface) ui.Element {
	dark := c.Theme().Dark
	hoverFill, pressFill := theme.Ink.Alpha(0.07), theme.Ink.Alpha(0.11)
	if dark {
		hoverFill, pressFill = theme.Veil(2), theme.Veil(3)
	}
	strong := color.Alpha(1)
	if strong == theme.Neutral(dark, 11) {
		strong = theme.Neutral(dark, 12)
	}
	selected := surface == PlayerSelectedSurface
	bg, fg := ui.Transparent, color
	if selected {
		bg, fg = hoverFill, strong
	}
	if b.Hovered() || b.FocusVisible() {
		bg, fg = hoverFill, strong
	}
	pose := ui.Pose{}
	if b.Pressed() {
		bg, fg, pose.Scale = pressFill, strong, 0.97
	}
	motion := ui.Animation{Duration: 140 * time.Millisecond}
	return b.Background(b.AnimateColor("bg", bg, motion)).
		TextColor(b.AnimateColor("fg", fg, motion)).AnimatePose(pose, motion)
}

// RoundButton is the button of IconButton, without its icon.
func RoundButton(c *ui.Context, size float32, color ui.Color, label string, surface ...ButtonSurface) ui.Element {
	b := ui.ButtonBase(c).Size(size+16, size+16).Center().Radius(999).Label(label).TextColor(color)
	if len(surface) > 0 {
		switch surface[0] {
		case MediaSurface:
			return MediaFeedback(b, color)
		case PlayerSurface, PlayerSelectedSurface:
			b.Size(max(32, size+16), max(32, size+16)).Tooltip(label)
			return PlayerFeedback(c, b, color, surface[0])
		}
	}
	bg := ui.Transparent
	if b.Hovered() {
		bg = ui.RGBA(128, 128, 128, 0.14)
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	pose := ui.Pose{}
	if b.Pressed() {
		pose.Scale = 0.88
	}
	b.AnimatePose(pose, ui.Snappy)
	return b
}

// PillButton is a rounded button with an icon and a label.
func PillButton(c *ui.Context, icon *ui.SVG, label string, bg, fg ui.Color) ui.Element {
	b := ui.ButtonBase(c).Height(36).PaddingX(18).Gap(8).Radius(999).AlignItems(ui.Center).TextColor(fg)
	fill := bg
	if b.Hovered() {
		fill = bg.Mix(ui.RGB(255, 255, 255), 0.12)
	}
	b.Background(b.AnimateColor("bg", fill, ui.Animation{Duration: 120 * time.Millisecond}))
	pose := ui.Pose{}
	if b.Pressed() {
		pose.Scale = 0.96
	}
	b.AnimatePose(pose, ui.Snappy)
	b.Children(func() {
		if icon != nil {
			ui.Icon(c, icon).Size(14, 14)
		}
		ui.Text(c, label).FontSize(14).FontWeight(600).SingleLine()
	})
	return b
}

// Clock formats a duration as minutes and seconds.
func Clock(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	s := int(d.Round(time.Second) / time.Second)
	return fmt.Sprintf("%d:%02d", s/60, s%60)
}

// Plays formats a play count as people in China read it.
func Plays(n float64) string {
	switch {
	case n >= 1e8:
		return fmt.Sprintf("%.1f亿", n/1e8)
	case n >= 1e4:
		return fmt.Sprintf("%.0f万", n/1e4)
	default:
		return fmt.Sprintf("%.0f", n)
	}
}

// Greeting returns what to say at this hour.
func Greeting(now time.Time) string {
	switch h := now.Hour(); {
	case h < 5:
		return "夜深了"
	case h < 11:
		return "早上好"
	case h < 13:
		return "中午好"
	case h < 18:
		return "下午好"
	default:
		return "晚上好"
	}
}

// Segmented is a segmented control as AppKit draws one: segments of one
// width in a small gray track, divided by hairlines, the one chosen raised
// on a capsule that slides to the next as it changes. It reports whether
// the user chose another.
func Segmented(c *ui.Context, selected *int, labels ...string) bool {
	t := c.Theme()
	const height, inset, radius = 28, 2, 14
	width := float32(0)
	for _, l := range labels {
		w, _ := c.MeasureText(0, ui.Span{Text: l, Size: 13})
		width = max(width, w+26)
	}
	track, capsule, divider := theme.Neutral(t.Dark, 3), theme.Neutral(t.Dark, 1), theme.Neutral(t.Dark, 6)
	if t.Dark {
		capsule = theme.Neutral(true, 5)
	}
	before := *selected
	parts := ui.SegmentedBase(c, selected, len(labels))
	at := parts.Track.AnimateTo("at", float32(*selected), ui.Animation{Spring: ui.SpringOptions{Duration: 280 * time.Millisecond, Bounce: 0.08}})
	n := len(labels)
	dark := t.Dark
	parts.Track.Height(height).Padding(inset).Radius(radius).Background(track).Draw(func(p *ui.Painter, r ui.Rect) {
		// Hairlines between segments, but beside the capsule.
		for i := 1; i < n; i++ {
			if d := float32(i) - at; d > -1 && d < 2 {
				continue
			}
			x := r.X + inset + float32(i)*width
			p.Fill(ui.Rect{X: x - 0.5, Y: r.Y + 6, W: 1, H: r.H - 12}, divider, 0)
		}
		pill := ui.Rect{X: r.X + inset + at*width, Y: r.Y + inset, W: width, H: r.H - 2*inset}
		if !dark {
			p.Shadow(pill, radius-1, 0, 0.5, 1, 0, theme.Shade(4))
			p.Shadow(pill, radius-1, 0, 1, 3, 0, theme.Shade(1))
		}
		p.Fill(pill, capsule, radius-1)
	}).Children(func() {
		for i, l := range labels {
			seg := parts.Segment(i).Width(width).Height(height - 2*inset).Center().Radius(radius - 1)
			if seg.Pressed() && i != *selected {
				seg.Background(divider)
			}
			seg.Children(func() { ui.Text(c, l).FontSize(13).TextColor(t.Text).SingleLine() })
		}
	})
	return *selected != before
}
