package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/base/skeleton"
)

// The blocks that stand in for what is loading live in
// lilt/internal/ui/base/skeleton; the app keeps these names while its page
// layouts still live here.

type shimmer = skeleton.Shimmer

var boneWidth = skeleton.BoneWidth

// pageShimmer is the shimmer of the pages, on their background.
func (a *app) pageShimmer(c *ui.Context) shimmer {
	s := shimmer{Base: ink.Alpha(0.06), Sheen: ink.Alpha(0.022)}
	if a.pal.Dark {
		s.Base, s.Sheen = white.Alpha(0.07), white.Alpha(0.13)
	}
	return s.In(c)
}

// npShimmer is the shimmer over the now playing view's backdrop.
func npShimmer(c *ui.Context) shimmer {
	return shimmer{Base: white.Alpha(0.1), Sheen: white.Alpha(0.2)}.In(c)
}

// homeSkeleton is the home page while it loads.
func (a *app) homeSkeleton(c *ui.Context, s shimmer) {
	s.Bone(c.Key("daily"), dailyRadius).Height(dailyHeight)
	cards := a.columns(c, 168)
	if a.profile != nil {
		a.songShelfSkeleton(c.Key("song-shelves"), s)
		a.section(c, "今日为你推荐")
		a.playlistCardsSkeleton(c.Key("rec"), s, cards, 168, false)
	}
	if a.profile != nil {
		a.songShelfSkeleton(c.Key("more-song-shelves"), s)
	}
	a.section(c, "精选歌单")
	a.playlistCardsSkeleton(c.Key("featured"), s, cards*2, 168, false)
}

func (a *app) chartCardsSkeleton(c *ui.Context, s shimmer) {
	cols := min(4, max(2, a.columns(c, 260)))
	skeleton.ChartCards(c, s, cols, a.pal.Hover)
}

// playlistCardsSkeleton is n of playlistCard's cards, in columns at least
// min DIPs wide, with a line for how often they update if freq.
func (a *app) playlistCardsSkeleton(c *ui.Context, s shimmer, n int, min float32, freq bool) {
	skeleton.PlaylistCards(c, s, n, a.columns(c, min), freq)
}

// trackSkeleton is trackRows while its tracks load: rows of blocks under
// the header, which may hold blocks of its own.
func (a *app) trackSkeleton(c *ui.Context, s shimmer, covers bool, header func(c *ui.Context)) {
	top, bottom := a.insets(c)
	a.underTop, a.underBottom = false, true
	ui.Column(c.Key("skeleton")).Grow(1).Padding(top+pageGap, pagePad-12, bottom+32, pagePad-12).Children(func() {
		a.startMark(c)
		a.trackSkeletonRows(c, s, covers, header)
	})
}

func (a *app) trackSkeletonRows(c *ui.Context, s shimmer, covers bool, header func(c *ui.Context)) {
	w, _ := c.Size()
	skeleton.TrackRows(c, s, covers, w-sidebarWidth > 820, header)
}

// playlistHeaderSkeleton is playlistHeader while the playlist loads.
func (a *app) playlistHeaderSkeleton(c *ui.Context, s shimmer) {
	skeleton.PlaylistHeader(c, s)
}
