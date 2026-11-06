package pages

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/navigation"
	"lilt/internal/player"
	"lilt/internal/qqmusic"
	"lilt/internal/theme"
	"lilt/internal/transport"
)

const (
	// TrackRowHeight is how tall a song's row is.
	TrackRowHeight = 54
	// PageGap is the room between the bar over a page and its content.
	PageGap = 12
)

// ListState keeps the list's row anchor and its animated offset together when
// a page is hidden or the listener switches between collection tabs.
type ListState struct {
	ui.ListState
	Scroll   ui.ScrollState
	locating int64
	target   float32
}

// ListDeps are what the list of songs needs from the app it runs in.
type ListDeps struct {
	Player      *player.Player
	Palette     theme.Palette
	Route       *navigation.Route
	Liked       map[int64]bool
	UnderTop    *bool
	UnderBottom *bool
	Bars        *transport.Bar
	Sidebar     float32

	Insets      func(*ui.Context) (top, bottom float32)
	StartMark   func(*ui.Context)
	EndMark     func(*ui.Context)
	SongItems   func(*qqmusic.Track, int64) func(*ui.Menu)
	Cover       func(*ui.Context, string, int, float32, float32) ui.Element
	ArtistLinks func(*ui.Context, *qqmusic.Track, ui.Color, ui.Color, ...bool) ui.Element
	Link        func(*ui.Context, string, ui.Color, ui.Color, navigation.Route)
	Like        func(*ui.Context, *qqmusic.Track, float32, ui.Color, ...appui.ButtonSurface) ui.Element
	PlayQueue   func(tracks []qqmusic.Track, at int)
}

// TrackRows lists tracks under a header, which scrolls with them; a click
// plays the list from the track clicked.
func TrackRows(c *ui.Context, d ListDeps, key any, state *ListState, tracks []qqmusic.Track, covers bool, header func(*ui.Context), footer ...func(*ui.Context)) {
	p := d.Palette
	cur := d.Player.Current()
	current := -1
	if cur != nil {
		for i := range tracks {
			if tracks[i].ID == cur.ID {
				current = i
				break
			}
		}
	}
	// Geometry and offset come from the same committed frame. Any measured
	// song can locate another: song rows share a height, unlike the header.
	scrollY := state.Scroll.Y
	currentTop, measured := float32(0), false
	// A song's menu can take it out of the playlist it is listed in.
	var pid int64
	if d.Route.Page == navigation.Playlist {
		pid = d.Route.ID
	}
	w, _ := c.Size()
	// An album's page has no column of the album its songs are on.
	wide := w-d.Sidebar > 820 && d.Route.Page != navigation.Album
	top, bottom := d.Insets(c)
	// Rows out of view are not built, their marks with them: a list
	// scrolled away from its start or end has content under the bars.
	*d.UnderTop, *d.UnderBottom = true, true
	rows := len(tracks) + 1 + len(footer)
	ui.List(c.Key(key), &state.ListState, rows).TrackScroll(&state.Scroll).Grow(1).
		Padding(top+PageGap, Pad-12, bottom+32, Pad-12).ScrollbarInsets(top, 0, bottom, 0).
		Rows(func(item ui.ListRow) {
			c := item.Context
			i := item.Index
			if i == 0 {
				heading := ui.Column(c)
				if r := heading.Bounds(); current >= 0 && r.H > 0 {
					currentTop, measured = r.Y+r.H+float32(current)*TrackRowHeight, true
				}
				heading.Children(func() {
					d.StartMark(c)
					header(c)
				})
				return
			}
			if i == rows-1 {
				defer d.EndMark(c)
			}
			if i > len(tracks) {
				footer[i-len(tracks)-1](c)
				return
			}
			i--
			tr := &tracks[i]
			playing := cur != nil && cur.ID == tr.ID
			row := ui.ButtonBase(c.Key(tr.ID)).Height(TrackRowHeight).PaddingX(12).Gap(14).Radius(14).AlignItems(ui.Center).FocusRing(false).Label(tr.Name).
				ContextMenu(d.SongItems(tr, pid))
			if r := row.Bounds(); current >= 0 && !measured && r.H > 0 {
				currentTop, measured = r.Y+float32(current-i)*TrackRowHeight, true
			}
			hovered := row.Hovered()
			bg := ui.Transparent
			switch {
			case playing && row.Pressed():
				bg = theme.Accent(p.Dark, 5)
			case playing && hovered:
				bg = theme.Accent(p.Dark, 4)
			case playing:
				bg = theme.Accent(p.Dark, 3)
			case hovered:
				bg = p.Hover
			}
			row.Background(row.AnimateColor("bg", bg, ui.Animation{Duration: 100 * time.Millisecond}))
			if !tr.Playable() {
				row.Opacity(0.42)
			}
			row.Children(func() {
				ui.Box(c).Width(26).Center().Children(func() {
					switch {
					case playing:
						NowPlayingBars(c, d, p.Accent)
					case hovered && tr.Playable():
						ui.Icon(c, appui.IconPlay).Size(13, 13).TextColor(p.Text)
					default:
						ui.Textf(c, "%d", i+1).FontSize(13).FontFeatures("tnum").TextColor(p.Faint)
					}
				})
				if covers {
					d.Cover(c, tr.Album.PicURL, 96, 38, 6)
				}
				ui.Column(c).Grow(3).Basis(0).MinWidth(0).Gap(3).Children(func() {
					name := p.Text
					if playing {
						name = p.Accent
					}
					spans := []ui.Span{{Text: tr.Name, Weight: 500, Color: name}}
					if len(tr.Alias) > 0 {
						spans = append(spans, ui.Span{Text: "（" + tr.Alias[0] + "）", Color: p.Faint})
					}
					ui.RichText(c, spans...).FontSize(14).SingleLine()
					ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
						if !tr.Playable() {
							ui.Text(c, "VIP").FontSize(9).FontWeight(800).TextColor(p.Accent).Padding(0, 4).Radius(3).Border(1, p.Accent.Alpha(0.6))
						}
						d.ArtistLinks(c, tr, p.Muted, p.Text).FontSize(12.5).SingleLine().Shrink(1).MinWidth(0)
						if tr.Reason != "" {
							ui.Text(c, tr.Reason).FontSize(10.5).FontWeight(600).TextColor(p.Accent).Padding(1, 6).Radius(999).Background(p.Accent.Alpha(0.1)).SingleLine()
						}
					})
				})
				if wide {
					album := ui.RichText(c).Grow(2).Basis(0).MinWidth(0).FontSize(12.5).TextColor(p.Muted).SingleLine()
					album.Children(func() {
						if tr.Album.ID == 0 {
							ui.Text(c, tr.Album.Name)
						} else {
							d.Link(c.Key("album"), tr.Album.Name, p.Muted, p.Text, navigation.Route{Page: navigation.Album, ID: tr.Album.ID})
						}
					})
				}
				// A liked song's heart shows; the others' as the pointer is on them.
				ui.Box(c).Size(32, 32).Center().Children(func() {
					if d.Liked[tr.ID] || hovered {
						d.Like(c.Key("like"), tr, 15, p.Muted)
					}
				})
				ui.Text(c, appui.Clock(tr.Length())).FontSize(12.5).FontFeatures("tnum").TextColor(p.Faint).Width(44).TextAlign(ui.End)
			})
			if row.Clicked() && tr.Playable() {
				if playing {
					d.Player.TogglePlay()
				} else {
					d.PlayQueue(tracks, i)
				}
			}
		})
	LocateTrackButton(c, d, state, cur, current, currentTop, scrollY, measured)
}

// NowPlayingBars dance in the row of the track playing.
func NowPlayingBars(c *ui.Context, d ListDeps, color ui.Color) {
	ui.Box(c).Size(16, 14).Draw(func(p *ui.Painter, r ui.Rect) {
		lv, live := d.Player.Levels()
		d.Bars.Step(p.Now(), lv, live)
		transport.DrawBars(p, r, d.Bars, 4, color, 2.5)
		if live || d.Bars.Group(0, player.Bands) > 0.01 {
			p.AnimationFrame()
		}
	})
}

// LocateTrackButton floats the button that brings the song playing into view.
func LocateTrackButton(c *ui.Context, d ListDeps, state *ListState, current *qqmusic.Track, currentIndex int, currentTop, scrollY float32, measured bool) {
	// Visible asks for another build when layout changes the visible rows.
	// Geometry below also excludes rows covered by the floating page bars.
	state.Visible()
	bounds := state.Bounds(c)
	if current == nil || !measured || bounds.H == 0 {
		state.locating = 0
		return
	}
	top, bottom := d.Insets(c)
	viewTop, viewBottom := bounds.Y+top, bounds.Y+bounds.H-bottom
	target := scrollY + currentTop - (viewTop+viewBottom-TrackRowHeight)/2
	if state.locating != 0 {
		if state.locating != current.ID || !state.Scroll.Scrolling() {
			state.locating = 0
		} else if delta := target - state.target; delta < -1 || delta > 1 {
			// Unvisited rows initially use estimated heights. As the glide
			// measures them, keep its destination on the actual song.
			state.target = target
			state.Scroll.ScrollTo(0, target, ui.Smooth)
		}
	}
	if currentTop+TrackRowHeight > viewTop && currentTop < viewBottom {
		return
	}

	p := d.Palette
	b := ui.ButtonBase(c.Key("locate-track")).Absolute().Right(20).Bottom(bottom+20).
		Size(32, 32).Radius(999).Center().Label("定位当前歌曲").Tooltip("定位当前歌曲").
		TextColor(theme.Accent(p.Dark, 11)).Shadow(0, 3, 12, 0, p.CardShade.Alpha(0.65))
	bg := theme.Neutral(p.Dark, 1).Alpha(0.88)
	border := theme.Accent(p.Dark, 6).Alpha(0.55)
	if b.Pressed() {
		bg = theme.Accent(p.Dark, 4)
		border = theme.Accent(p.Dark, 6)
	} else if b.Hovered() {
		bg = theme.Accent(p.Dark, 3).Alpha(0.96)
		border = theme.Accent(p.Dark, 6).Alpha(0.85)
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	b.Border(1, b.AnimateColor("border", border, ui.Animation{Duration: 120 * time.Millisecond}))
	pose := ui.Pose{}
	if b.Pressed() {
		pose.Scale = 0.94
	}
	b.AnimatePose(pose, ui.Snappy)
	b.Children(func() { ui.Icon(c, appui.IconLocate).Size(16, 16) })
	if b.Clicked() {
		if c.Preferences().ReduceMotion {
			// No glide will measure the intervening rows: ask the virtual
			// list to measure and place the destination directly.
			state.ScrollTo(currentIndex+1, ui.Center)
			state.locating = 0
			return
		}
		state.locating, state.target = current.ID, target
		state.Scroll.ScrollTo(0, target, ui.Smooth)
	}
}
