package nowplaying

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/art"
	"lilt/internal/playerbar"
	"lilt/internal/qqmusic"
	"lilt/internal/transport"
)

// stageBar is the slim bar along the foot of the stage: what plays, where
// it is, and the controls.
func stageBar(c *ui.Context, d Deps, tr *qqmusic.Track, coverArt *art.Artwork, accent ui.Color) {
	pl := d.Player
	ui.Column(c).Padding(0, 48, 28, 48).Gap(10).Children(func() {
		transport.Scrubber(c.Key("stage-scrub"), pl, d.Settings.ProgressEffectMode(), d.NpScrub, white.Alpha(0.18), white.Alpha(0.9), 14, accent.Mix(white, 0.35))
		ui.Row(c).Gap(14).AlignItems(ui.Center).Children(func() {
			ui.Row(c).Grow(1).Basis(0).MinWidth(0).Gap(12).AlignItems(ui.Center).Children(func() {
				if bmp := appui.Bitmap(coverArt); bmp != nil {
					cov := ui.Box(c.Key(tr.ID)).Size(44, 44)
					in := appui.ArtIn(cov, bmp)
					cov.Children(func() { appui.ArtImage(c, bmp, 8, in) })
				}
				ui.Column(c).Shrink(1).MinWidth(0).Gap(2).Children(func() {
					ui.Text(c, tr.Name).FontSize(14).FontWeight(700).SingleLine()
					d.ArtistLinks(c, tr, white.Alpha(0.55), white).FontSize(12).SingleLine()
				})
				d.Like(c.Key("like"), tr, 17, white.Alpha(0.6), appui.MediaSurface)
				d.SongMenu(c.Key("more"), tr, 17, white.Alpha(0.6), appui.MediaSurface).Margin(0, 0, 0, -10)
			})
			playerbar.Transport(c, pl, white, white.Alpha(0.45), white, 22, appui.MediaSurface)
			ui.Row(c).Grow(1).Basis(0).MinWidth(0).Justify(ui.End).Children(func() {
				ui.Text(c, appui.Clock(pl.Position())+" / "+appui.Clock(pl.Duration())).FontSize(11.5).FontFeatures("tnum").TextColor(white.Alpha(0.5))
			})
		})
	})
}
