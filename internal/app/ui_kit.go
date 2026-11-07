package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
)

// cover shows an artwork, or a placeholder of the same size while it loads.
func (a *app) cover(c *ui.Context, url string, px int, size, radius float32) ui.Element {
	p := a.pal
	box := ui.Box(c).Size(size, size).Radius(radius).Center()
	bmp := bitmap(a.art.Get(qqmusic.Sized(url, px), false))
	in := artIn(box, bmp)
	if in < 1 {
		box.Background(p.Hover)
	}
	box.Children(func() {
		if in < 1 {
			ui.Icon(c, iconNote).Size(size*0.3, size*0.3).TextColor(p.Faint).Opacity(1 - in)
		}
		artImage(c, bmp, radius, in)
	})
	return box
}

// paged is how a list loaded a page at a time is coming along.
type paged struct {
	loaded  bool // whether the first page came
	loading bool
	more    bool   // whether there is a page after those loaded
	err     string // why the last page didn't load
}

// pageFoot ends a list loaded a page at a time, asking next for the next
// page once it is within a screen of coming into view: where the last
// failed, it says so, with a button to try again.
func (a *app) pageFoot(c *ui.Context, pg *paged, next func()) {
	foot := ui.Row(c.Key("foot")).Height(56).Gap(12).Center()
	foot.Children(func() {
		switch {
		case pg.loading:
			ui.Text(c, "正在加载…").FontSize(12.5).TextColor(a.pal.Muted)
		case pg.err != "":
			ui.Text(c, pg.err).FontSize(12.5).TextColor(a.pal.Muted)
			if ui.Button(c.Key("retry"), "重试").Clicked() {
				next()
			}
		}
	})
	_, h := c.Size()
	if b := foot.Bounds(); pg.more && !pg.loading && pg.err == "" && b.W > 0 && b.Y < 2*h {
		next()
	}
}
