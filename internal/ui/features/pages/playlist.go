package pages

import (
	"fmt"
	"strings"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/navigation"
	"lilt/internal/media/player"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
)

// PlaylistInfo is what the app has of a playlist.
type PlaylistInfo struct {
	Playlist *qqmusic.Playlist
	Loading  bool
	Err      string
}

// PlaylistDeps are what a playlist's page needs from the app it runs in.
type PlaylistDeps struct {
	Palette theme.Palette
	Route   *navigation.Route
	Player  *player.Player
	Card    Deps

	Liked     func(int64) bool
	LikedPage func(*ui.Context, int64)
	Info      func(int64) PlaylistInfo
	Load      func(int64)
	Centered  func(*ui.Context, string, func())
	Skeleton  func(*ui.Context)
	Tracks    func(*ui.Context, []qqmusic.Track, func(*ui.Context))

	Editable   func(int64) bool
	Rename     func(int64)
	Delete     func(int64)
	Saved      func(int64) bool
	Unfavorite func(int64)
}

// PlaylistPage shows a playlist: its header, then its songs.
func PlaylistPage(c *ui.Context, d PlaylistDeps, id int64) {
	if d.Liked(id) {
		d.LikedPage(c, id)
		return
	}
	info := d.Info(id)
	if info.Playlist == nil {
		if !info.Loading {
			d.Centered(c, info.Err, func() { d.Load(id) })
			return
		}
		d.Skeleton(c)
		return
	}
	d.Tracks(c, info.Playlist.Tracks, func(c *ui.Context) { PlaylistHeader(c, d, info.Playlist) })
}

// PlaylistHeader is a playlist's cover, name, metadata and actions.
func PlaylistHeader(c *ui.Context, d PlaylistDeps, pl *qqmusic.Playlist) {
	p := d.Palette
	kind := "歌单"
	if pl.UpdateFreq != "" {
		kind = "榜单 · " + pl.UpdateFreq
	}
	var meta []string
	if pl.Creator.Nickname != "" {
		meta = append(meta, pl.Creator.Nickname)
	}
	meta = append(meta, fmt.Sprintf("%d 首", len(pl.Tracks)))
	if pl.PlayCount > 0 {
		meta = append(meta, appui.Plays(pl.PlayCount)+" 次播放")
	}
	CollectionHeader(c, d.Card, pl, kind, strings.Join(meta, " · "), func() {
		if appui.PillButton(c.Key("play"), appui.IconPlay, "播放", p.Accent, p.OnAccent).Clicked() {
			if d.Player.Shuffle() {
				d.Player.ToggleShuffle()
			}
			d.Player.PlayQueue(pl.Tracks, FirstPlayable(pl.Tracks))
		}
		if appui.PillButton(c.Key("shuffle"), appui.IconShuffle, "随机播放", p.Hover.Over(p.Bg), p.Text).Clicked() {
			if !d.Player.Shuffle() {
				d.Player.ToggleShuffle()
			}
			d.Player.PlayQueue(pl.Tracks, FirstPlayable(pl.Tracks))
		}
		if d.Editable(pl.ID) {
			more := appui.IconButton(c.Key("more"), appui.IconMore, 18, p.Muted, "更多")
			more.Menu(func(m *ui.Menu) {
				if m.Item("重命名…").Chosen() {
					d.Rename(pl.ID)
				}
				m.Separator()
				if m.Item("删除歌单…").Chosen() {
					d.Delete(pl.ID)
				}
			})
		} else if d.Saved(pl.ID) {
			more := appui.IconButton(c.Key("more"), appui.IconMore, 18, p.Muted, "更多")
			more.Menu(func(m *ui.Menu) {
				if m.Item("取消收藏…").Chosen() {
					d.Unfavorite(pl.ID)
				}
			})
		}
	})
}
