package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/media/lyrics"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/features/nowplaying"
)

// The backdrop and the quiet parts' hot comments live in
// lilt/internal/ui/features/nowplaying; the app keeps their state and these
// names while its other views still live here.
type backdropCache = nowplaying.BackdropCache
type quietState = nowplaying.QuietState

var (
	backdropColors = nowplaying.BackdropColors
	drawFog        = nowplaying.DrawFog
)

const backdropFrame = nowplaying.BackdropFrame
const backdropFade = nowplaying.BackdropFade

func (a *app) backdrop(c *ui.Context, artwork *artwork, tint ui.Color) {
	nowplaying.Backdrop(c, a.nowplayingDeps(), artwork, tint)
}

func (a *app) backdropOver(p *ui.Painter, win, band ui.Rect, alpha func(float32) float32) {
	nowplaying.BackdropOver(p, &a.backdropCache, win, band, alpha)
}

func (a *app) quietComment(tr *qqmusic.Track, lines []lyrics.Line, loaded bool, pos time.Duration) (*qqmusic.Comment, time.Duration) {
	return nowplaying.QuietComment(a.nowplayingDeps(), tr, lines, loaded, pos)
}

func (a *app) quietShown(c *ui.Context, el ui.Element, tr *qqmusic.Track, lines []lyrics.Line, loaded bool) (float32, bool) {
	return nowplaying.QuietShown(c, a.nowplayingDeps(), el, tr, lines, loaded)
}

func (a *app) quietCard(c *ui.Context, dim, top float32) {
	nowplaying.QuietCard(c, a.nowplayingDeps(), dim, top)
}

func (a *app) stageQuiet(c *ui.Context, dim, f, w float32) {
	nowplaying.StageQuiet(c, a.nowplayingDeps(), dim, f, w)
}
