// Package playerbar owns the bar along the foot of the pages: what plays,
// its controls and its progress, the volume beside them, the queue that
// floats above it, and the glass it sits on.
package playerbar

import (
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/lyricsui"
	"lilt/internal/player"
	"lilt/internal/qqmusic"
	"lilt/internal/settings"
	"lilt/internal/theme"
	"lilt/internal/transport"
)

// Height is the height of the player along the foot of the pages.
const Height = 84

// Picked is the song a login picked up, while it waits to play.
type Picked struct {
	ID    int64
	Heard string // where and when, as "12分钟前在手机上听过"
}

// Ink holds the colors of what is on a pane, as the queue uses them.
type Ink struct {
	Text, Muted ui.Color
}

// Deps are what the bar needs from the app it runs in: the player and the
// settings it reads directly, the listener's state by pointer, and the
// views that live elsewhere by callback.
type Deps struct {
	Player   *player.Player
	Settings *settings.Settings
	Palette  theme.Palette

	Volume       *float64
	VolumeOpen   *bool
	MutedFrom    *float64
	Picked       *Picked
	BarScrub     *transport.State
	QueueOpen    *bool
	QueuePopup   *ui.ListState
	QueuePopupAt *int

	UnderBottom bool

	Cover        func(c *ui.Context, url string, px int, size, radius float32) ui.Element
	Like         func(c *ui.Context, tr *qqmusic.Track, size float32, color ui.Color, surface ...appui.ButtonSurface) ui.Element
	ArtistLinks  func(c *ui.Context, tr *qqmusic.Track, color, hover ui.Color, album ...bool) ui.Element
	ProgressFill func(p theme.Palette) ui.Color
	Pane         func(e ui.Element, radius float32, onDark bool) ui.Element
	Ink          func(onDark bool) Ink
	QueueTracks  func(c *ui.Context, list *ui.ListState, lastAt *int, p theme.Palette) ui.Element
	TopBar       func(c *ui.Context) float32

	OpenNowPlaying         func()
	SetLyricsEnabled       func(bool)
	SetDesktopLyricsLocked func(bool)
}

// Bar runs along the foot of the content: what plays, its controls and its
// progress, and the volume.
func Bar(c *ui.Context, d Deps) {
	p := d.Palette
	pl := d.Player
	tr := pl.Current()
	transport.TickTimes(c, pl)
	if d.Picked.ID != 0 && (tr == nil || tr.ID != d.Picked.ID || pl.Playing()) {
		// Played, or gone: it says what it is again.
		*d.Picked = Picked{}
	}
	liquid := d.Settings.LiquidGlass
	// Over the foot of the page: while content is under it, its edge frosts
	// it, with a hairline along the top. Glass floats clear of the foot
	// instead, a capsule over what the page scrolls under it.
	glassH := float32(Height - glassBarFoot)
	if !liquid {
		edge := ui.Box(c.Key("bottom-edge")).Absolute().Left(0).Right(0).Bottom(0).Height(Height).
			Material(glass.ScrollEdge{Hard: true, Bottom: true, Background: p.Bg})
		edge.Opacity(appui.EdgeShown(edge, d.UnderBottom))
	}
	// The controls keep the middle, between sides that share what is left:
	// narrower, the middle gives up to 140 DIPs, and what plays keeps room
	// for its cover, some of its title and its heart.
	controls := float32(2*32 + 4) // queue and the speaker
	if lyricsui.SupportsDesktopLyrics() {
		controls += 32 + 4
	}
	bar := ui.Row(c).Absolute().Gap(20).AlignItems(ui.Center)
	if liquid {
		bar.Left(glassBarSide).Right(glassBarSide).Bottom(glassBarFoot).Height(glassH).PaddingX(20).
			Radius(glassH/2).Shadow(0, 6, 24, -6, ui.RGBA(0, 0, 0, 0.3)).
			Material(barGlass{dark: p.Dark, clear: float32(d.Settings.GlassClear), bend: float32(d.Settings.GlassBend)})
	} else {
		bar.Left(0).Right(0).Bottom(0).Height(Height).PaddingX(20)
	}
	bar.Children(func() {
		// What plays.
		ui.Row(c).Grow(1).Basis(0).MinWidth(150).Children(func() {
			if tr == nil {
				ui.Row(c).Gap(12).AlignItems(ui.Center).Children(func() {
					ui.Box(c).Size(48, 48).Radius(8).Background(p.Hover).Center().Children(func() {
						ui.Icon(c, appui.IconNote).Size(18, 18).TextColor(p.Faint)
					})
					ui.Text(c, "未在播放").FontSize(13).TextColor(p.Muted)
				})
				return
			}
			info := ui.ButtonBase(c.Key("info")).Gap(12).AlignItems(ui.Center).FocusRing(false).Shrink(1).MinWidth(0).Label("打开正在播放")
			hovered := info.Hovered() || info.FocusVisible()
			motion := ui.Animation{Duration: 160 * time.Millisecond}
			reveal := info.AnimateTo("reveal", boolFloat(hovered), motion)
			info.Children(func() {
				cover := ui.Box(c).Size(48, 48).Shrink(0)
				cover.Translate(0, -reveal).Children(func() {
					d.Cover(c, tr.Album.PicURL, 120, 48, 8).Shadow(0, 2, 8, 0, p.CardShade)
					ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(8).Center().PassThrough().
						Opacity(reveal).Background(theme.Shade(5)).TextColor(theme.Paper).Children(func() {
						ui.Icon(c, appui.IconDown).Size(18, 18).Rotate(180).Translate(0, 3*(1-reveal))
					})
				})
				ui.Column(c).Shrink(1).MinWidth(0).Gap(3).Children(func() {
					ui.Text(c, tr.Name).FontSize(13.5).FontWeight(600).SingleLine()
					// What it is waiting on shows in place of the artists.
					var sub string
					color := p.Muted
					switch {
					case d.Picked.ID == tr.ID:
						sub = d.Picked.Heard
					case pl.Loading():
						sub = "正在缓冲…"
					case pl.Err() != "":
						sub, color = pl.Err(), p.Accent
					}
					if sub == "" {
						d.ArtistLinks(c, tr, p.Muted, p.Text).FontSize(12).SingleLine()
					} else {
						ui.Text(c, sub).FontSize(12).TextColor(color).SingleLine()
					}
				})
			})
			if info.Clicked() {
				d.OpenNowPlaying()
			}
			d.Like(c.Key("like"), tr, 16, p.Muted, appui.PlayerSurface).Margin(0, 0, 0, 4).Shrink(0)
		})

		// Controls over the progress.
		ui.Column(c).Width(440).Shrink(1).MinWidth(300).Gap(0).AlignItems(ui.Stretch).Children(func() {
			Transport(c, pl, p.Text, p.Muted, p.Accent, 18, appui.PlayerSurface)
			ui.Row(c).Gap(10).AlignItems(ui.Center).Margin(-4, 0, 0, 0).Children(func() {
				ui.Text(c, appui.Clock(pl.Position())).FontSize(10.5).FontFeatures("tnum").TextColor(p.Faint).Width(34).TextAlign(ui.End)
				transport.Scrubber(c.Key("bar-scrub"), pl, d.Settings.ProgressEffectMode(), d.BarScrub, p.Press, d.ProgressFill(p), 14, p.Accent).Grow(1)
				ui.Text(c, appui.Clock(pl.Duration())).FontSize(10.5).FontFeatures("tnum").TextColor(p.Faint).Width(34)
			})
		})

		// Queue, volume and the switch for the selected external lyrics mode.
		ui.Row(c).Grow(1).Basis(0).MinWidth(controls).Justify(ui.End).AlignItems(ui.Center).Gap(4).Children(func() {
			queueButton(c, d)
			volumeButton(c, d, p.Text.Alpha(0.75))
			if lyricsui.SupportsDesktopLyrics() {
				externalLyricsPlayerButton(c, d, p)
			}
		})
	})
}

// Centered lyric lines identify the floating caption window. Open corners
// leave it free to move; closing the same frame marks its fixed position.
const desktopLyricsGlyph = `<path d="M7 10h10M9 14h6"/>`

var (
	iconDesktopLyrics       = appui.StrokeIcon(desktopLyricsGlyph + `<path d="M8 4H6a3 3 0 0 0-3 3v2M16 4h2a3 3 0 0 1 3 3v2M21 15v2a3 3 0 0 1-3 3h-2M8 20H6a3 3 0 0 1-3-3v-2"/>`)
	iconDesktopLyricsLocked = appui.StrokeIcon(desktopLyricsGlyph + `<rect x="3" y="4" width="18" height="16" rx="3"/>`)
	iconTaskbarLyrics       = appui.StrokeIcon(`<rect x="3" y="4" width="18" height="16" rx="3"/><path d="M3 14h18M7 17h10"/>`)
)

func externalLyricsPlayerButton(c *ui.Context, d Deps, p theme.Palette) {
	mode := lyricsui.Mode(*d.Settings)
	label, icon := "桌面歌词", iconDesktopLyrics
	if mode == lyricsui.TaskbarMode {
		label, icon = "任务栏歌词", iconTaskbarLyrics
	}
	color, hint := p.Muted, "开启"+label
	if lyricsui.Enabled(*d.Settings) {
		color, hint = p.Accent, "关闭"+label
		if mode == lyricsui.DesktopMode {
			hint += " · 右键锁定"
			if d.Settings.DesktopLyricsLocked {
				hint, icon = "关闭桌面歌词 · 右键解锁", iconDesktopLyricsLocked
			}
		}
	}
	button := appui.IconButton(c.Key("external-lyrics"), icon, 18, color, label, appui.PlayerSurface).Size(32, 32).Tooltip(hint)
	button.OnClick(func() { d.SetLyricsEnabled(!lyricsui.Enabled(*d.Settings)) })
	if button.RightClicked() && mode == lyricsui.DesktopMode && d.Settings.DesktopLyrics {
		d.SetDesktopLyricsLocked(!d.Settings.DesktopLyricsLocked)
	}
}

// boolFloat is 0 for false and 1 for true.
func boolFloat(b bool) float32 {
	if b {
		return 1
	}
	return 0
}
