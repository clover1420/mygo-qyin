package pages

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/navigation"
	"lilt/internal/player"
	"lilt/internal/qqmusic"
)

// LikedAlbumsInfo is what the app has loaded of the listener's saved albums.
type LikedAlbumsInfo struct {
	Albums  []qqmusic.AlbumProfile
	Loaded  bool
	Loading bool
	More    bool
	Err     string
}

// LikedScroll keeps each collection's place while the tabs switch.
type LikedScroll struct {
	Songs  *ListState
	Albums *ui.ListState
}

// LikedDeps are what the liked collection needs from the app it runs in.
type LikedDeps struct {
	Route  *navigation.Route
	Player *player.Player
	Card   Deps
	Album  AlbumDeps
	Mine   []qqmusic.Playlist

	Playlist func(id int64) (pl *qqmusic.Playlist, loading bool, err string)
	Albums   func() LikedAlbumsInfo
	State    func() LikedScroll

	ListPage      func(*ui.Context, any, *ui.ListState, func())
	TrackRows     func(*ui.Context, any, *ListState, []qqmusic.Track, func(*ui.Context))
	TrackSkeleton func(*ui.Context, func(*ui.Context))
	CardsSkeleton func(*ui.Context)
	UnderTop      *bool
	UnderBottom   *bool

	PageFoot     func(*ui.Context, func())
	Empty        func(*ui.Context, string)
	Centered     func(*ui.Context, string, func())
	Show         func(navigation.Route)
	LoadPlaylist func(int64)
	LoadAlbums   func(again bool)
	MoreAlbums   func()
}

// LikedPlaylistPage shows the listener's liked songs, or the albums they
// saved, as the tab chosen.
func LikedPlaylistPage(c *ui.Context, d LikedDeps, id int64) {
	content := navigation.Route{Page: navigation.Playlist, ID: id}
	header := func(c *ui.Context) {
		likedCollectionHeader(c, d, id)
		ui.Row(c).AlignItems(ui.Center).Gap(12).Margin(0, 0, 18, 0).Children(func() {
			tab := d.Route.Tab
			if appui.Segmented(c.Key("liked-kind"), &tab, "单曲", "专辑") {
				next := *d.Route
				next.Tab = tab
				d.Show(next)
			}
		})
	}
	if d.Route.Tab == 1 {
		info := d.Albums()
		state := d.State()
		d.ListPage(c, content, state.Albums, func() {
			header(c)
			AlbumGrid(c, d.Album, info.Albums)
			if !info.Loaded && info.Loading {
				d.CardsSkeleton(c)
				return
			}
			d.PageFoot(c, d.MoreAlbums)
			if info.Loaded && len(info.Albums) == 0 {
				d.Empty(c, "暂无收藏专辑")
			}
		})
		return
	}
	pl, loading, err := d.Playlist(id)
	state := d.State()
	if pl == nil {
		if loading {
			d.ListPage(c, content, &state.Songs.ListState, func() {
				*d.UnderTop, *d.UnderBottom = false, true
				d.TrackSkeleton(c, header)
			})
		} else {
			d.ListPage(c, content, &state.Songs.ListState, func() {
				header(c)
				d.Centered(c, err, func() { d.LoadPlaylist(id) })
			})
		}
		return
	}
	d.TrackRows(c, content, state.Songs, pl.Tracks, header)
}

// likedCollectionHeader is the liked collection's cover, count and actions.
func likedCollectionHeader(c *ui.Context, d LikedDeps, id int64) {
	p := d.Card.Palette
	pl, _, _ := d.Playlist(id)
	loaded := pl != nil
	if pl == nil {
		for i := range d.Mine {
			if d.Mine[i].ID == id {
				pl = &d.Mine[i]
				break
			}
		}
	}
	if pl == nil {
		pl = &qqmusic.Playlist{ID: id, Name: "我喜欢"}
	}
	count := pl.TrackCount
	if loaded {
		count = len(pl.Tracks)
	}
	meta := fmt.Sprintf("%d 首单曲", count)
	if d.Route.Tab == 1 {
		meta = "正在加载专辑"
		info := d.Albums()
		switch {
		case info.Loaded:
			more := ""
			if info.More {
				more = "+"
			}
			meta = fmt.Sprintf("%d%s 张专辑", len(info.Albums), more)
		case info.Err != "":
			meta = "专辑暂不可用"
		}
	}
	CollectionHeader(c, d.Card, pl, "收藏", meta, func() {
		if d.Route.Tab == 1 {
			info := d.Albums()
			appui.PillButton(c.Key("refresh-liked-albums"), appui.IconReload, "刷新", p.Hover.Over(p.Bg), p.Text).
				Label("刷新收藏专辑").Disabled(info.Loading).
				OnClick(func() { d.LoadAlbums(true) })
			return
		}
		appui.PillButton(c.Key("play"), appui.IconPlay, "播放", p.Accent, p.OnAccent).Disabled(!loaded || len(pl.Tracks) == 0).
			OnClick(func() {
				if d.Player.Shuffle() {
					d.Player.ToggleShuffle()
				}
				d.Player.PlayQueue(pl.Tracks, FirstPlayable(pl.Tracks))
			})
		appui.PillButton(c.Key("shuffle"), appui.IconShuffle, "随机播放", p.Hover.Over(p.Bg), p.Text).Disabled(!loaded || len(pl.Tracks) == 0).
			OnClick(func() {
				if !d.Player.Shuffle() {
					d.Player.ToggleShuffle()
				}
				d.Player.PlayQueue(pl.Tracks, FirstPlayable(pl.Tracks))
			})
	})
}
