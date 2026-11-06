package likes

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
)

// SongMenu is three dots that open the menu of tr.
func SongMenu(c *ui.Context, d Deps, tr *qqmusic.Track, size float32, color ui.Color, surface ...appui.ButtonSurface) ui.Element {
	b := appui.IconButton(c, appui.IconMore, size, color, "更多", surface...)
	b.Menu(SongItems(d, tr, 0))
	return b
}

// SongItems builds the menu of what can be done with tr, behind its dots or
// a right click on its row: playing it next, copying its title, its title
// with its artists, or the link to its page, opening its artists' or album's,
// and, signed in, adding it to a playlist or taking it out of pid, the
// playlist it is listed in, if set.
func SongItems(d Deps, tr *qqmusic.Track, pid int64) func(m *ui.Menu) {
	return func(m *ui.Menu) {
		if tr.ProgramID != 0 {
			programItems(d, m, tr)
			return
		}
		if m.Item("下一首播放").Disabled(!tr.Playable()).Chosen() {
			d.Player.PlayNext(*tr)
			*d.Toast = "已设为下一首播放"
		}
		m.Separator()
		if m.Item("复制歌名").Chosen() {
			d.Copy(tr.Name)
		}
		if m.Item("复制歌名和歌手").Chosen() {
			d.Copy(tr.Name + " - " + tr.ArtistNames())
		}
		m.Separator()
		if m.Item("复制链接").Chosen() {
			d.Copy(tr.URL())
		}
		// Its artists' and album's pages, those NetEase has.
		linked := tr.Album.ID != 0
		for _, ar := range tr.Artists {
			linked = linked || ar.ID != 0
		}
		if linked {
			m.Separator()
		}
		for _, ar := range tr.Artists {
			if ar.ID != 0 && m.Item("查看歌手："+ar.Name).Chosen() {
				d.OpenArtist(ar.ID)
			}
		}
		if tr.Album.ID != 0 && m.Item("查看专辑："+tr.Album.Name).Chosen() {
			d.OpenAlbum(tr.Album.ID)
		}
		d.PlaylistItems(m, tr, pid)
	}
}

// programItems builds the menu of an episode: copying its title or the link
// to its page, or opening its podcast's.
func programItems(d Deps, m *ui.Menu, tr *qqmusic.Track) {
	if m.Item("复制节目名").Chosen() {
		d.Copy(tr.Name)
	}
	if m.Item("复制链接").Chosen() {
		d.Copy(tr.URL())
	}
	m.Separator()
	if m.Item("查看播客").Chosen() {
		d.OpenPodcast(tr.PodcastID)
	}
}
