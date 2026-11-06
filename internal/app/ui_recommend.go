package app

import (
	"slices"

	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
)

var readSongShelves = qqmusic.RecommendedSongShelves

// The request and its old shelves share an identity until the listener changes.
type songShelfData struct {
	shelves []qqmusic.SongShelf
	loading bool
	err     string
}

func (a *app) loadSongShelves() {
	if a.profile == nil || a.songShelves != nil && a.songShelves.loading {
		return
	}
	if a.songShelves == nil {
		a.songShelves = &songShelfData{}
	}
	d := a.songShelves
	d.loading, d.err = true, ""
	fetch := readSongShelves
	go func() {
		shelves, err := fetch()
		a.post(func() { a.songShelvesAnswered(d, shelves, err) })
	}()
}

func (a *app) songShelvesAnswered(d *songShelfData, shelves []qqmusic.SongShelf, err error) {
	if a.songShelves != d {
		return
	}
	d.loading = false
	if err != nil {
		d.err = "歌曲推荐加载失败"
		return
	}
	d.shelves, d.err = shelves, ""
	a.shelfPagers = nil
}

// Account changes discard personal snapshots, including those in Back and Forward.
func (a *app) clearSongShelves() {
	a.songShelves = nil
	a.shelfPagers = nil
	if a.route.Page == pageSongShelf {
		a.route = route{Page: pageHome}
	}
	isShelf := func(r route) bool { return r.Page == pageSongShelf }
	a.history = slices.DeleteFunc(a.history, isShelf)
	a.forward = slices.DeleteFunc(a.forward, isShelf)
	for r := range a.pageStates {
		if isShelf(r) {
			delete(a.pageStates, r)
		}
	}
}

func (a *app) songShelfColumns(c *ui.Context) int {
	w, _ := c.Size()
	return min(3, max(1, int((w-sidebarWidth-2*pagePad+16)/240)))
}

func (a *app) shelfAction(c *ui.Context, icon *ui.SVG, label string) ui.Element {
	b := ui.ButtonBase(c).Height(32).PaddingX(8).Gap(5).Radius(8).AlignItems(ui.Center).Label(label)
	color := a.pal.Muted
	if b.Hovered() || b.FocusVisible() {
		b.Background(a.pal.Hover)
		color = a.pal.Text
	}
	b.TextColor(color).Children(func() {
		if icon != nil {
			ui.Icon(c, icon).Size(14, 14)
		}
		ui.Text(c, label).FontSize(12).SingleLine()
	})
	return b
}
func (a *app) playSongShelf(shelf *qqmusic.SongShelf) {
	if !shelfHasPlayable(shelf) {
		return
	}
	if a.pl.Shuffle() {
		a.pl.ToggleShuffle()
	}
	a.pl.PlayQueue(shelf.Tracks, firstPlayable(shelf.Tracks))
}

func (a *app) playShelfSong(shelf *qqmusic.SongShelf, i int) {
	tr := &shelf.Tracks[i]
	if !tr.Playable() {
		return
	}
	if cur := a.pl.Current(); cur != nil && cur.ID == tr.ID {
		a.pl.TogglePlay()
		return
	}
	a.pl.PlayQueue(shelf.Tracks, i)
}
func (a *app) songShelfSkeleton(c *ui.Context, s shimmer) {
	ui.Column(c).Margin(28, 0, 0, 0).Gap(10).Children(func() {
		s.Text(c, 19, 1.4).WidthPercent(40)
		ui.Grid(c).Columns(a.songShelfColumns(c)).GapX(16).GapY(4).Children(func() {
			for i := range a.songShelfColumns(c) * 2 {
				ui.Row(c.Key(i)).Height(76).Padding(6).Gap(10).AlignItems(ui.Center).Children(func() {
					s.Bone(c, 6).Size(48, 48)
					ui.Column(c).Grow(1).MinWidth(0).Gap(5).Children(func() {
						s.Text(c, 13.5, 1.3).WidthPercent(boneWidth(i))
						s.Text(c, 12, 1.3).WidthPercent(45)
					})
				})
			}
		})
	})
}
