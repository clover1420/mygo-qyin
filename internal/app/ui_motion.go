package app

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// The way the app moves: what changes the view, and what grows over it.
var (
	motionEase = ui.CubicBezier(0.22, 0.8, 0.24, 1)
	motionPage = ui.Animation{Duration: 340 * time.Millisecond, Ease: motionEase}
)

// pageKey names a page for its identity, which a tab of it leaves alone: the
// subtree of a page keeps its state as what it lists changes, and only
// another page takes it down.
type pageKey struct {
	page  pageKind
	id    int64
	query string
}

func keyOf(r route) pageKey { return pageKey{page: r.Page, id: r.ID, query: r.Query} }

func clamp01(v float32) float32 { return min(1, max(0, v)) }

func lerp(a, b, t float32) float32 { return a + (b-a)*t }

// pageLevel is how deep a page is: those of the sidebar, and what opens from
// them.
func pageLevel(r route) int {
	switch r.Page {
	case pagePlaylist, pageArtist, pageAlbum, pageUser, pagePodcast, pageSongShelf:
		return 1
	}
	return 0
}

// pageMotion moves a page as it comes: in from the right on the way deeper,
// from the left on the way back, and across with a small rise between two
// pages of the same depth. With no exit: two pages showing through each
// other read as neither.
func (a *app) pageMotion(c *ui.Context, page ui.Element) {
	if !a.pageIn || c.Preferences().ReduceMotion {
		return
	}
	in := &ui.Motion{}
	switch {
	case a.pageDir > 0:
		in.X = 34
	case a.pageDir < 0:
		in.X = -24
	default:
		in.Y = 10
	}
	page.Transition(ui.ElementTransition{Animation: &motionPage, Enter: in})
}
