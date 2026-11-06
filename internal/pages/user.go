package pages

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/navigation"
	"lilt/internal/qqmusic"
	"lilt/internal/theme"
)

// PhotoSize is the width of a user's or artist's round photo on their page.
const PhotoSize = 212

// UserInfo is what the app has of a user: who they are, the playlists they
// made that anyone may see, and, of the listener's own page, what they
// follow.
type UserInfo struct {
	User          *qqmusic.User
	Playlists     []qqmusic.Playlist
	Loading       int
	Err           string
	Artists       []qqmusic.ArtistProfile
	Podcasts      []qqmusic.Podcast
	ArtistsShown  bool
	PodcastsShown bool
}

// UserDeps are what a user's page needs from the app it runs in.
type UserDeps struct {
	Palette theme.Palette
	Sidebar float32
	Own     bool

	Info          func() UserInfo
	Load          func(int64)
	Centered      func(*ui.Context, string, func())
	Skeleton      func(*ui.Context)
	ScrollPage    func(*ui.Context, func())
	Badge         func(*ui.Context)
	Empty         func(*ui.Context, string)
	CardsSkeleton func(*ui.Context)

	PlaylistCard func(*ui.Context, *qqmusic.Playlist)
	ArtistCard   func(*ui.Context, *qqmusic.ArtistProfile)
	PodcastRow   func(*ui.Context, *qqmusic.Podcast)
	Cover        func(*ui.Context, string, int, float32, float32) ui.Element
	CoverFill    func(*ui.Context, string, int, float32)
	Open         func(navigation.Route)
}

// UserPage shows a user: who they are, the playlists they made, and for the
// listener, what they follow.
func UserPage(c *ui.Context, d UserDeps, id int64) {
	p := d.Palette
	info := d.Info()
	if info.User == nil {
		if info.Loading == 0 {
			d.Centered(c, info.Err, func() { d.Load(id) })
			return
		}
		d.Skeleton(c)
		return
	}
	u := info.User
	d.ScrollPage(c, func() {
		ui.Row(c).Gap(28).AlignItems(ui.Center).Children(func() {
			ui.Box(c).Size(PhotoSize, PhotoSize).Radius(PhotoSize/2).Shadow(0, 18, 40, -10, p.CardShade).Children(func() {
				d.CoverFill(c, u.AvatarURL, 600, PhotoSize/2)
			})
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).AlignItems(ui.Start).Gap(8).Children(func() {
				ui.Text(c, u.Nickname).MinWidth(0).FontSize(30).FontWeight(700).LetterSpacing(-0.6).MaxLines(2).LineHeight(1.2)
				if d.Own {
					d.Badge(c)
				}
				if u.Signature != "" {
					ui.Text(c, u.Signature).FontSize(12.5).TextColor(p.Muted).MaxLines(2).LineHeight(1.45)
				}
			})
		})
		Section(c, "创建的歌单")
		switch {
		case info.Loading > 0:
			d.CardsSkeleton(c)
		case len(info.Playlists) == 0:
			d.Empty(c, "暂无公开歌单")
		default:
			ui.Grid(c).Columns(Columns(c, d.Sidebar, 168)).GapX(20).GapY(28).Children(func() {
				for i := range info.Playlists {
					d.PlaylistCard(c.Key(info.Playlists[i].ID), &info.Playlists[i])
				}
			})
		}
		if d.Own && info.PodcastsShown {
			Section(c, "订阅的播客")
			if len(info.Podcasts) == 0 {
				d.Empty(c, "还没有订阅播客")
			}
			ui.Grid(c).Columns(Columns(c, d.Sidebar, 340)).GapX(8).GapY(4).Margin(0, -12).Children(func() {
				for i := range info.Podcasts {
					d.PodcastRow(c.Key(info.Podcasts[i].ID), &info.Podcasts[i])
				}
			})
		}
		if d.Own && info.ArtistsShown {
			Section(c, "关注的歌手")
			if len(info.Artists) == 0 {
				d.Empty(c, "还没有关注歌手")
			}
			ui.Grid(c).Columns(Columns(c, d.Sidebar, 132)).GapX(20).GapY(28).Children(func() {
				for i := range info.Artists {
					d.ArtistCard(c.Key(info.Artists[i].ID), &info.Artists[i])
				}
			})
		}
	})
}

// UserRow is a user in a list, opening their page.
func UserRow(c *ui.Context, d UserDeps, u *qqmusic.User) {
	p := d.Palette
	row := ui.ButtonBase(c).Padding(10, 12).Gap(14).Radius(14).AlignItems(ui.Center).FocusRing(false).Label(u.Nickname)
	if row.Hovered() {
		row.Background(p.Hover)
	}
	row.Children(func() {
		d.Cover(c, u.AvatarURL, 120, 48, 24)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, u.Nickname).FontSize(14).FontWeight(600).SingleLine()
			if u.Signature != "" {
				ui.Text(c, u.Signature).FontSize(12.5).TextColor(p.Muted).SingleLine()
			}
		})
	})
	if row.Clicked() {
		d.Open(navigation.Route{Page: navigation.User, ID: u.UserID})
	}
}
