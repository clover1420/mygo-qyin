// Package pages owns the shapes the app's pages share: their headings and
// sections, the cards of playlists and charts, the header of a collection,
// and the covers they draw.
package pages

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/art"
	"lilt/internal/navigation"
	"lilt/internal/qqmusic"
	"lilt/internal/theme"
)

// Pad is the room between a page's edge and its content.
const Pad = 40

// Deps are what the shared page shapes need from the app: the palette, the
// covers, and what opens a page.
type Deps struct {
	Palette     theme.Palette
	Art         *art.Cache
	Open        func(navigation.Route)
	Load        func(id int64, play bool)
	Description func(*ui.Context, string)
}

// Columns returns how many cards of at least min DIPs fit across the page.
func Columns(c *ui.Context, sidebar, min float32) int {
	w, _ := c.Size()
	inner := w - sidebar - 2*Pad
	return max(2, int((inner+20)/(min+20)))
}

// Heading is a page's title, over its subtitle when it has one.
func Heading(c *ui.Context, p theme.Palette, title, sub string) {
	ui.Column(c).Gap(4).Margin(0, 0, 18, 0).Children(func() {
		if sub != "" {
			ui.Text(c, sub).FontSize(12).FontWeight(600).TextColor(p.Muted).LetterSpacing(0.4)
		}
		ui.Text(c, title).FontSize(30).FontWeight(700).LetterSpacing(-0.6)
	})
}

// Section names a part of a page.
func Section(c *ui.Context, title string) ui.Element {
	return ui.Text(c, title).FontSize(19).FontWeight(700).LetterSpacing(-0.3).Margin(34, 0, 14, 0)
}

// Weekday names the day t falls on.
func Weekday(t time.Time) string {
	return [...]string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}[t.Weekday()]
}

// FirstPlayable is where a list starts playing: its first song that may play.
func FirstPlayable(tracks []qqmusic.Track) int {
	for i := range tracks {
		if tracks[i].Playable() {
			return i
		}
	}
	return 0
}

// PlaylistCard is a square cover with the playlist's name below: it lifts
// under the pointer, where a play button shows on it.
func PlaylistCard(c *ui.Context, d Deps, pl *qqmusic.Playlist) {
	if CoverCard(c, d, pl.Name, pl.Cover(), pl.PlayCount, pl.UpdateFreq, func() { d.Load(pl.ID, true) }) {
		d.Open(navigation.Route{Page: navigation.Playlist, ID: pl.ID})
	}
}

// ArtistCard is an artist's round photo over their name, opening their page.
func ArtistCard(c *ui.Context, d Deps, ar *qqmusic.ArtistProfile) {
	p := d.Palette
	card := ui.ButtonBase(c).Column().Gap(10).AlignItems(ui.Stretch).FocusRing(false).Label(ar.Name)
	hovered := card.Hovered()
	card.Children(func() {
		photo := ar.Avatar
		if photo == "" {
			photo = ar.PicURL
		}
		art := ui.Box(c).AspectRatio(1).Radius(999)
		if hovered {
			art.Shadow(0, 12, 28, -6, p.CardShade)
		} else {
			art.Shadow(0, 2, 8, -2, p.CardShade)
		}
		art.Children(func() { CoverFill(c, d.Art, p, photo, 400, 999) })
		ui.Text(c, ar.Name).FontSize(13).FontWeight(500).SingleLine().AlignSelf(ui.Center)
	})
	if card.Clicked() {
		d.Open(navigation.Route{Page: navigation.Artist, ID: ar.ID})
	}
}

// CoverCard is a playlist's card, or a podcast's: its cover, with how many
// times it was played on it, over its name and a line under that. The
// button on the cover calls play; it reports whether the rest was clicked.
func CoverCard(c *ui.Context, d Deps, name, cover string, playCount float64, sub string, play func()) bool {
	p := d.Palette
	// From the top: a grid stretches the cards of a row as tall as the
	// tallest, whose name or line under it takes two lines.
	card := ui.ButtonBase(c).Column().Gap(10).AlignItems(ui.Stretch).Justify(ui.Start).FocusRing(false).Label(name)
	hovered := card.Hovered()
	card.Children(func() {
		box := ui.Box(c).AspectRatio(1).Radius(14)
		lift := ui.Pose{}
		if hovered {
			lift = ui.Pose{Y: -3}
			box.Shadow(0, 12, 28, -6, p.CardShade)
		} else {
			box.Shadow(0, 2, 8, -2, p.CardShade)
		}
		box.AnimatePose(lift, ui.Snappy)
		box.Children(func() {
			CoverFill(c, d.Art, p, cover, 400, 10)
			if playCount > 0 {
				ui.Row(c).Absolute().Top(8).Right(8).Padding(2, 7).Radius(999).Gap(3).AlignItems(ui.Center).
					Background(theme.Ink.Alpha(0.35)).TextColor(theme.Paper).Children(func() {
					ui.Icon(c, appui.IconHeadset).Size(10, 10)
					ui.Text(c, appui.Plays(playCount)).FontSize(10.5).FontWeight(600)
				})
			}
			CoverPlayButton(c, p, hovered, play)
		})
		ui.Text(c, name).FontSize(13).FontWeight(500).MaxLines(2).LineHeight(1.35)
		if sub != "" {
			ui.Text(c, sub).FontSize(12).TextColor(p.Muted).SingleLine().Margin(-6, 0, 0, 0)
		}
	})
	return card.Clicked()
}

// CoverFill fills its parent with a cover, its corners rounded by radius.
// The cover rounds itself rather than its parent clipping it: a clip moved
// by a transform, as a card lifting under the pointer, is drawn on the CPU.
func CoverFill(c *ui.Context, covers *art.Cache, p theme.Palette, url string, px int, radius float32) {
	bmp := appui.Bitmap(covers.Get(qqmusic.Sized(url, px), false))
	box := ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(radius)
	in := appui.ArtIn(box, bmp)
	if in < 1 {
		box.Background(p.Hover)
	}
	box.Children(func() { appui.ArtImage(c, bmp, radius, in) })
}

// CoverPlayButton plays a playlist or a podcast from its card, without
// opening it.
func CoverPlayButton(c *ui.Context, p theme.Palette, shown bool, play func()) {
	b := ui.ButtonBase(c.Key("play")).Absolute().Bottom(10).Right(10).Size(38, 38).Radius(999).Center().
		Background(p.Accent).TextColor(theme.Paper).Label("播放").FocusRing(false).
		Shadow(0, 4, 12, 0, theme.Ink.Alpha(0.3))
	pose := ui.Pose{Fade: 1, Scale: 0.8, Y: 6}
	if shown || b.Hovered() {
		pose = ui.Pose{Scale: 1}
		if b.Hovered() {
			pose.Scale = 1.08
		}
	}
	b.AnimatePose(pose, ui.Snappy)
	b.Children(func() { ui.Icon(c, appui.IconPlay).Size(15, 15).Margin(0, 0, 0, 2) })
	if b.Clicked() {
		play()
	}
}

// ChartCard shows a chart with its top three songs beside its cover.
func ChartCard(c *ui.Context, d Deps, pl *qqmusic.Playlist) {
	p := d.Palette
	card := ui.ButtonBase(c).Gap(14).Padding(12).Radius(14).AlignItems(ui.Center).FocusRing(false).Label(pl.Name)
	hovered := card.Hovered()
	bg := p.Hover
	if hovered {
		bg = p.Press
	}
	card.Background(card.AnimateColor("bg", bg, ui.Animation{Duration: 140 * time.Millisecond}))
	card.Children(func() {
		ui.Box(c).Size(84, 84).Radius(8).Shadow(0, 3, 10, -2, p.CardShade).Children(func() {
			CoverFill(c, d.Art, p, pl.Cover(), 200, 9)
		})
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(5).Children(func() {
			ui.Row(c).Gap(6).AlignItems(ui.Baseline).Children(func() {
				ui.Text(c, pl.Name).FontSize(14).FontWeight(700).SingleLine()
				ui.Text(c, pl.UpdateFreq).FontSize(11).TextColor(p.Faint).SingleLine()
			})
			for i, line := range pl.Preview {
				ui.Row(c).Gap(7).Children(func() {
					ui.Textf(c, "%d", i+1).FontSize(12).FontWeight(700).Width(10).FontFeatures("tnum").TextColor(p.Muted)
					ui.RichText(c,
						ui.Span{Text: line.First, Weight: 500},
						ui.Span{Text: "  " + line.Second, Color: p.Muted},
					).FontSize(12).SingleLine().Shrink(1).MinWidth(0)
				})
			}
		})
	})
	if card.Clicked() {
		d.Open(navigation.Route{Page: navigation.Playlist, ID: pl.ID})
	}
}

// CollectionHeader keeps collection artwork, metadata and actions aligned
// across content types.
func CollectionHeader(c *ui.Context, d Deps, pl *qqmusic.Playlist, kind, meta string, actions func()) {
	p := d.Palette
	ui.Row(c).Gap(28).AlignItems(ui.End).Padding(0, 0, 26, 0).Children(func() {
		ui.Box(c).Size(212, 212).Radius(14).Shadow(0, 18, 40, -10, p.CardShade).Children(func() {
			CoverFill(c, d.Art, p, pl.Cover(), 600, 14)
		})
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
			ui.Text(c, kind).FontSize(12).FontWeight(600).TextColor(p.Accent)
			ui.Text(c, pl.Name).FontSize(30).FontWeight(800).LetterSpacing(-0.6).MaxLines(2).LineHeight(1.2)
			ui.Text(c, meta).FontSize(13).TextColor(p.Muted).SingleLine()
			d.Description(c, pl.Description)
			ui.Row(c).Height(36).Gap(10).Margin(10, 0, 0, 0).Children(actions)
		})
	})
}
