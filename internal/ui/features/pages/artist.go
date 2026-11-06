package pages

import (
	"fmt"
	"slices"
	"strings"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/navigation"
	"lilt/internal/media/player"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
)

// What an artist's page lists: their songs, the most popular or the
// newest, their albums, or their events.
const (
	ArtistHot = iota
	ArtistLatest
	ArtistAlbums
	ArtistEvents
)

var artistTabTitles = [...]string{"歌曲", "歌曲", "专辑", "动态"}

// ArtistInfo is what the app has of an artist.
type ArtistInfo struct {
	Profile       *qqmusic.ArtistProfile
	Loading       bool
	Err           string
	Latest        []qqmusic.Track
	LatestLoading bool
	LatestErr     string
	Albums        []qqmusic.AlbumProfile
	AlbumsLoaded  bool
	AlbumsLoading bool
	AlbumsMore    bool
	AlbumsErr     string
}

// ArtistDeps are what an artist's page needs from the app it runs in.
type ArtistDeps struct {
	Palette theme.Palette
	Route   *navigation.Route
	Player  *player.Player
	Tab     *int

	Info       func() ArtistInfo
	Load       func(int64)
	LoadLatest func(int64)
	LoadAlbums func(int64)
	Centered   func(*ui.Context, string, func())
	Skeleton   func(*ui.Context)
	Tracks     func(*ui.Context, []qqmusic.Track, func(*ui.Context))
	ScrollPage func(*ui.Context, func())
	PageFoot   func(*ui.Context, func())
	Empty      func(*ui.Context, string)
	AlbumGrid  func(*ui.Context, []qqmusic.AlbumProfile)
	Events     func(*ui.Context)
	CoverFill  func(*ui.Context, string, int, float32)
}

// ArtistPage shows who an artist is over their songs, the most popular or
// the newest, their albums or their events.
func ArtistPage(c *ui.Context, d ArtistDeps, id int64) {
	info := d.Info()
	if info.Profile == nil {
		if !info.Loading {
			d.Centered(c, info.Err, func() { d.Load(id) })
			return
		}
		d.Skeleton(c)
		return
	}
	header := func(c *ui.Context) { ArtistHeader(c, d, info) }
	switch *d.Tab {
	case ArtistAlbums:
		d.ScrollPage(c, func() {
			header(c)
			d.AlbumGrid(c, info.Albums)
			d.PageFoot(c, func() { d.LoadAlbums(id) })
			if info.AlbumsLoaded && len(info.Albums) == 0 {
				d.Empty(c, "暂无专辑")
			}
		})
		return
	case ArtistEvents:
		d.ScrollPage(c, func() {
			header(c)
			d.Events(c)
		})
		return
	}
	songs := info.Profile.Songs
	if *d.Tab == ArtistLatest {
		switch {
		case info.Latest != nil:
			songs = info.Latest
		case info.LatestLoading:
			d.Skeleton(c)
			return
		default:
			songs = nil
			header = func(c *ui.Context) {
				ArtistHeader(c, d, info)
				ui.Column(c).Padding(40, 0).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Text(c, info.LatestErr).TextColor(d.Palette.Muted).FontSize(13)
					if ui.Button(c.Key("retry"), "重试").Clicked() {
						d.LoadLatest(info.Profile.ID)
					}
				})
			}
		}
	}
	d.Tracks(c, songs, header)
}

// ArtistHeader is an artist's photo, name, aliases, counts and actions.
func ArtistHeader(c *ui.Context, d ArtistDeps, info ArtistInfo) {
	p := d.Palette
	ar := info.Profile
	// What the buttons play: the songs listed, or else the most popular.
	songs := ar.Songs
	if *d.Tab == ArtistLatest && info.Latest != nil {
		songs = info.Latest
	}
	ui.Row(c).Gap(28).AlignItems(ui.End).Padding(0, 0, 8, 0).Children(func() {
		photo := ar.Avatar
		if photo == "" {
			photo = ar.PicURL
		}
		ui.Box(c).Size(PhotoSize, PhotoSize).Radius(PhotoSize/2).Shadow(0, 18, 40, -10, p.CardShade).Children(func() {
			d.CoverFill(c, photo, 600, PhotoSize/2)
		})
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
			ui.Text(c, "歌手").FontSize(12).FontWeight(600).TextColor(p.Accent)
			ui.Text(c, ar.Name).FontSize(30).FontWeight(800).LetterSpacing(-0.6).MaxLines(2).LineHeight(1.2)
			var meta []string
			if aka := distinct(append([]string{ar.Trans}, ar.Alias...)); len(aka) > 0 {
				meta = append(meta, strings.Join(aka, " / "))
			}
			if ar.MusicSize > 0 {
				meta = append(meta, fmt.Sprintf("%d 首歌曲", ar.MusicSize))
			}
			if ar.AlbumSize > 0 {
				meta = append(meta, fmt.Sprintf("%d 张专辑", ar.AlbumSize))
			}
			ui.Text(c, strings.Join(meta, " · ")).FontSize(13).TextColor(p.Muted)
			if desc := strings.TrimSpace(ar.BriefDesc); desc != "" {
				ui.Text(c, strings.ReplaceAll(desc, "\n", " ")).FontSize(12.5).TextColor(p.Muted).MaxLines(2).LineHeight(1.45).Margin(2, 0, 0, 0)
			}
			ui.Row(c).Gap(10).Margin(10, 0, 0, 0).Children(func() {
				if appui.PillButton(c.Key("play"), appui.IconPlay, "播放", p.Accent, p.OnAccent).Clicked() {
					if d.Player.Shuffle() {
						d.Player.ToggleShuffle()
					}
					d.Player.PlayQueue(songs, FirstPlayable(songs))
				}
				if appui.PillButton(c.Key("shuffle"), appui.IconShuffle, "随机播放", p.Hover.Over(p.Bg), p.Text).Clicked() {
					if !d.Player.Shuffle() {
						d.Player.ToggleShuffle()
					}
					d.Player.PlayQueue(songs, FirstPlayable(songs))
				}
			})
		})
	})
	ui.Row(c).AlignItems(ui.Center).Margin(34, 0, 14, 0).Children(func() {
		ui.Text(c, artistTabTitles[*d.Tab]).FontSize(19).FontWeight(700).LetterSpacing(-0.3)
		ui.Spacer(c)
		if appui.Segmented(c.Key("tab"), d.Tab, "热门", "最新", "专辑") {
			d.Load(ar.ID)
		}
	})
}

// distinct returns ss without empty strings, nor any after its first: an
// artist's translated name is often among their aliases too.
func distinct(ss []string) []string {
	var out []string
	for _, s := range ss {
		if s != "" && !slices.Contains(out, s) {
			out = append(out, s)
		}
	}
	return out
}
