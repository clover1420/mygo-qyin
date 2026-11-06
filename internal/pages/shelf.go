package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/player"
	"lilt/internal/qqmusic"
	"lilt/internal/theme"
)

// ShelfPager pages a shelf's songs across its grid.
type ShelfPager interface {
	Pages() int
	Start() int
	Size() int
	Page() int
	Layout(total, size int)
	Move(delta int) bool
	Scroll(ui.InputEvent) bool
	Previous() *ui.Handle
	Next() *ui.Handle
}

// ShelfInfo is what the app has of the listener's recommended songs.
type ShelfInfo struct {
	SignedIn bool
	Shelves  []qqmusic.SongShelf
	Loading  bool
	Err      string
}

// ShelfDeps are what a shelf needs from the app it runs in.
type ShelfDeps struct {
	Palette theme.Palette
	Card    Deps
	Player  *player.Player

	Info    func() ShelfInfo
	Pager   func(*qqmusic.SongShelf) ShelfPager
	Columns func(*ui.Context) int

	CoverFill func(*ui.Context, string, int, float32)
	Like      func(*ui.Context, *qqmusic.Track, float32, ui.Color)
	SongItems func(*qqmusic.Track, int64) func(*ui.Menu)
	Bars      func(*ui.Context, ui.Color)
	Action    func(*ui.Context, string, string) ui.Element
	Login     func()
	Reload    func()
	Skeleton  func(*ui.Context)
	Tracks    func(*ui.Context, []qqmusic.Track, func(*ui.Context))
	PlayShelf func(*qqmusic.SongShelf)
	PlaySong  func(*qqmusic.SongShelf, int)
	Centered  func(*ui.Context, string, func())
}

// SongShelfSection shows shelves [start, end) of the recommendations, or
// the invitation to sign in.
func SongShelfSection(c *ui.Context, d ShelfDeps, start, end int) {
	info := d.Info()
	if !info.SignedIn {
		if start != 0 {
			return
		}
		ui.Row(c).Margin(28, 0, 0, 0).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "歌曲推荐").FontSize(19).FontWeight(700).Grow(1)
			d.Action(c, "登录查看推荐", "user").OnClick(d.Login)
		})
		return
	}
	if len(info.Shelves) == 0 && info.Loading {
		if start == 0 {
			d.Skeleton(c)
		}
		return
	}
	for i := start; i < min(end, len(info.Shelves)); i++ {
		shelf := &info.Shelves[i]
		if len(shelf.Tracks) == 0 {
			continue
		}
		SongShelfCard(c.Key(shelf), d, shelf)
	}
	if start == 0 && (info.Err != "" || len(info.Shelves) == 0) {
		ui.Row(c.Key("status")).Margin(24, 0, 0, 0).Gap(10).AlignItems(ui.Center).Children(func() {
			message := info.Err
			if message == "" {
				message = "暂无歌曲推荐"
			}
			ui.Text(c, message).FontSize(12).TextColor(d.Palette.Muted)
			d.Action(c, "重试", "reload").Label("重试歌曲推荐").Disabled(info.Loading).OnClick(d.Reload)
		})
	}
}

// SongShelfCard is a shelf's title over its grid of songs, paged across.
func SongShelfCard(c *ui.Context, d ShelfDeps, shelf *qqmusic.SongShelf) {
	pager := d.Pager(shelf)
	columns := d.Columns(c)
	resized := pager.Size() != columns*2
	pager.Layout(len(shelf.Tracks), columns*2)
	section := ui.Column(c).Margin(28, 0, 0, 0).Gap(10)
	section.HandleInput(pager.Scroll)
	if pager.Pages() > 1 {
		moveByKey := func(direction int) {
			if !pager.Move(direction) {
				return
			}
			// A song can leave the grid after a key turn. Keep keyboard focus
			// on an enabled, stable arrow so subsequent keys stay in this shelf.
			if pager.Page() == 0 || direction > 0 && pager.Page() < pager.Pages()-1 {
				pager.Next().Focus(c)
			} else {
				pager.Previous().Focus(c)
			}
		}
		section.OnShortcut(0, ui.KeyLeft, func() { moveByKey(-1) }).
			OnShortcut(0, ui.KeyRight, func() { moveByKey(1) })
	}
	section.Children(func() {
		ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
			SongShelfTitle(c, shelf.Title).FontSize(19).FontWeight(700).LetterSpacing(-0.3).MaxLines(2).Grow(1).Shrink(1).MinWidth(0)
			if pager.Pages() > 1 {
				shelfPageButton(c.Key("previous"), d.Palette, appui.IconBack, "上一组歌曲："+shelf.Title, "上一组", pager.Page() == 0).
					Bind(pager.Previous()).OnClick(func() { pager.Move(-1) })
				shelfPageButton(c.Key("next"), d.Palette, appui.IconForward, "下一组歌曲："+shelf.Title, "下一组", pager.Page() == pager.Pages()-1).
					Bind(pager.Next()).OnClick(func() { pager.Move(1) })
			}
		})
		ui.Box(c.Key("viewport")).Clip().Children(func() {
			grid := ui.Grid(c.Key("songs")).Columns(columns).GapX(16).GapY(4)
			start := pager.Start()
			if resized {
				grid.Jump("start", float32(start))
			}
			at := grid.AnimateTo("start", float32(start), ui.Animation{Duration: 200 * time.Millisecond})
			lag := min(float32(1), max(float32(-1), (float32(start)-at)/float32(pager.Size())))
			fade := lag
			if fade < 0 {
				fade = -fade
			}
			grid.Translate(lag*20, 0).Opacity(1 - 0.22*fade).Children(func() {
				for i := start; i < min(start+pager.Size(), len(shelf.Tracks)); i++ {
					SongShelfRow(c.Key(shelf.Tracks[i].ID), d, shelf, i)
				}
			})
		})
	})
}

func shelfPageButton(c *ui.Context, p theme.Palette, icon *ui.SVG, label, tooltip string, disabled bool) ui.Element {
	color := p.Muted
	if disabled {
		color = theme.Neutral(p.Dark, 8)
	}
	b := ui.ButtonBase(c).Size(32, 32).Center().Radius(999).Label(label).Tooltip(tooltip).Disabled(disabled).TextColor(color)
	bg := ui.Transparent
	if !disabled && (b.Hovered() || b.FocusVisible()) {
		bg = p.Hover
	}
	b.Background(b.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond}))
	pose := ui.Pose{}
	if !disabled && b.Pressed() {
		pose.Scale = 0.9
	}
	b.AnimatePose(pose, ui.Animation{Duration: 100 * time.Millisecond})
	b.Children(func() { ui.Icon(c, icon).Size(16, 16) })
	return b
}

// SongShelfTitle is a shelf's title, its corner quotes narrowed to avoid
// the full-width opening quote's leading blank, while its label keeps the
// original.
func SongShelfTitle(c *ui.Context, title string) ui.Element {
	if rest, ok := strings.CutPrefix(title, "「"); ok {
		return ui.Text(c, "｢"+strings.Replace(rest, "」", "｣", 1)).Label(title)
	}
	return ui.Text(c, title)
}

// ShelfHasPlayable reports whether a shelf has a song that may play.
func ShelfHasPlayable(shelf *qqmusic.SongShelf) bool {
	return len(shelf.Tracks) > 0 && shelf.Tracks[FirstPlayable(shelf.Tracks)].Playable()
}

// SongShelfRow is a song in a shelf's grid.
func SongShelfRow(c *ui.Context, d ShelfDeps, shelf *qqmusic.SongShelf, i int) {
	p := d.Palette
	tr := &shelf.Tracks[i]
	cur := d.Player.Current()
	current := cur != nil && cur.ID == tr.ID
	row := ui.Row(c).Height(76).Padding(6).Gap(3).Radius(12).AlignItems(ui.Center).ContextMenu(d.SongItems(tr, 0))
	row.Children(func() {
		b := ui.ButtonBase(c.Key("song")).Grow(1).Shrink(1).MinWidth(0).Gap(10).AlignItems(ui.Center).Label(tr.Name).
			Disabled(!tr.Playable()).OnClick(func() { d.PlaySong(shelf, i) })
		hovered := row.Hovered() || b.FocusVisible()
		bg := ui.Transparent
		if current {
			bg = theme.Accent(p.Dark, 3)
		} else if hovered {
			bg = p.Hover
		}
		row.Background(row.AnimateColor("bg", bg, ui.Animation{Duration: 100 * time.Millisecond}))
		if !tr.Playable() {
			b.Opacity(0.42)
		}
		b.Children(func() {
			ui.Box(c).Size(48, 48).Radius(8).Children(func() {
				d.CoverFill(c, tr.Album.PicURL, 120, 6)
				if hovered || current {
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(8).Center().Background(theme.Shade(5)).PassThrough().Children(func() {
						if current && d.Player.Playing() {
							d.Bars(c, theme.Paper)
						} else {
							ui.Icon(c, appui.IconPlay).Size(15, 15).TextColor(theme.Paper)
						}
					})
				}
			})
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
				name := p.Text
				if current {
					name = theme.Accent(p.Dark, 11)
				}
				ui.Text(c, tr.Name).FontSize(13.5).FontWeight(500).TextColor(name).SingleLine()
				ui.Text(c, tr.ArtistNames()).FontSize(12).TextColor(p.Muted).SingleLine()
			})
		})
		d.Like(c.Key("like"), tr, 14, p.Muted)
	})
}

// SongShelfPage lists a whole shelf's songs.
func SongShelfPage(c *ui.Context, d ShelfDeps, shelf *qqmusic.SongShelf) {
	info := d.Info()
	if shelf == nil || !info.SignedIn {
		d.Centered(c, "暂无歌曲推荐", nil)
		return
	}
	d.Tracks(c, shelf.Tracks, func(c *ui.Context) {
		ui.Column(c).Padding(0, 12, 26, 12).Gap(8).Children(func() {
			SongShelfTitle(c, shelf.Title).FontSize(28).FontWeight(700).LetterSpacing(-0.6).MaxLines(3).LineHeight(1.25)
			ui.Text(c, fmt.Sprintf("%d 首歌曲", len(shelf.Tracks))).FontSize(13).TextColor(d.Palette.Muted)
			appui.PillButton(c.Key("play-all"), appui.IconPlay, "播放全部", d.Palette.Accent, d.Palette.OnAccent).
				AlignSelf(ui.Start).Margin(8, 0, 0, 0).Label("播放全部：" + shelf.Title).Disabled(!ShelfHasPlayable(shelf)).
				OnClick(func() { d.PlayShelf(shelf) })
		})
	})
}
