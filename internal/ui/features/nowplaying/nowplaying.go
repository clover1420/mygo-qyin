// Package nowplaying owns the view that covers the window with what plays:
// its cover over a backdrop of its own colors, the controls, the tabs, and
// the panes beside the cover. The panes' contents live elsewhere, and are
// called through Deps.
package nowplaying

import (
	"math"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/settings"
	"lilt/internal/media/art"
	"lilt/internal/media/captions"
	"lilt/internal/media/lyrics"
	"lilt/internal/media/player"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
	"lilt/internal/ui/base/transport"
	"lilt/internal/ui/features/playerbar"
)

var white = theme.Paper

var npSpring = ui.Animation{Spring: ui.SpringOptions{Duration: 450 * time.Millisecond, Bounce: 0.08}}

// npSettle is how long Now Playing takes to come or go, its spring at rest.
const npSettle = 600 * time.Millisecond

// The panes of Now Playing.
const (
	PaneLyrics = iota
	PaneQueue
	PaneStage
	PaneComments
	PaneProgram
	PaneCaptions
	PaneWall
)

// Deps are what Now Playing needs from the app it runs in: the player and
// the art it reads directly, the listener's state by pointer, and the panes
// that live elsewhere by callback.
type Deps struct {
	Player   *player.Player
	Settings *settings.Settings
	Art      *art.Cache
	NpScrub  *transport.State

	NowPlaying    *bool
	Open          *bool
	ChangedAt     *time.Time
	Pane          *int
	PaneTop       *float32
	NPArt         **art.Artwork
	BackdropCache *BackdropCache
	GlowColor     *ui.Color
	CoverState    *CoverState
	Volume        *float64
	MutedFrom     *float64
	QueueList     *ui.ListState
	QueueAt       *int
	LyricScroll   *ui.ScrollState
	LyricGlide    *Glide
	LyricAt       *int
	LyricTrack    *int64

	TopBar func(*ui.Context) float32

	SetButtonsTheme func(dark bool)
	ClearBackdrop   func()
	ResetPanes      func()
	ResetQueue      func()
	ResetWall       func()

	Backdrop         func(*ui.Context, *art.Artwork, ui.Color)
	CommentsFor      func(*qqmusic.Track)
	CaptionsOn       func() bool
	CaptionsShown    func() bool
	LyricsPaneChosen func() int

	PaneTabs func(*ui.Context)

	RadioChip   func(*ui.Context, string)
	ArtistLinks func(*ui.Context, *qqmusic.Track, ui.Color, ui.Color, ...bool) ui.Element
	Like        func(*ui.Context, *qqmusic.Track, float32, ui.Color, ...appui.ButtonSurface) ui.Element
	SongMenu    func(*ui.Context, *qqmusic.Track, float32, ui.Color, ...appui.ButtonSurface) ui.Element
	Cover       func(*ui.Context, string, int, float32, float32) ui.Element
	SongItems   func(*qqmusic.Track, int64) func(*ui.Menu)
	Avatar      func(*ui.Context, string, float32, ui.Color)

	Quiet         *QuietState
	QuietShown    func(*ui.Context, ui.Element, *qqmusic.Track, []lyrics.Line, bool) (float32, bool)
	QuietCard     func(*ui.Context, float32, float32)
	OpenComments  func()
	Skeleton      func(*ui.Context)
	StageQuiet    func(*ui.Context, float32, float32, float32)
	StageSkeleton func(*ui.Context, float32, float32)
	BackdropOver  func(*ui.Painter, ui.Rect, ui.Rect, func(float32) float32)

	Wall *WallState

	Comments          CommentFeed
	LoadComments      func()
	ToggleCommentLike func(int64)
	CommentSkeleton   func(*ui.Context, int)
	MoreSkeleton      func(*ui.Context)
	Accent            ui.Color

	Captions            *captions.Model
	SyncCaptions        func()
	StartModelDownload  func()
	CancelModelDownload func()
}

// State tracks Now Playing opening and closing, and reports whether it is
// open and whether it has come to rest, open or closed. Call it once a
// build, before the window's content.
func State(c *ui.Context, d Deps) (open, settled bool) {
	open = *d.NowPlaying && d.Player.Current() != nil
	now := c.Now()
	if open != *d.Open {
		*d.Open, *d.ChangedAt = open, now
		// A build once it is at rest: to stop painting what it covers, or
		// itself once gone.
		c.After(npSettle)
		// The window buttons light over its dark backdrop, and as the
		// theme over the pages (Windows).
		d.SetButtonsTheme(open)
	}
	return open, now.Sub(*d.ChangedAt) >= npSettle
}

// Layer holds Now Playing over the whole window, rising into view as it
// opens and sinking away as it closes. The layer moves itself, staying on
// top all the way, rather than leaving as an exit transition would, under
// the rest of the window: that would show it going over the sidebar's
// material while the page above it hid it at once. Its pose moves it in
// frames that only paint, building nothing.
func Layer(c *ui.Context, d Deps, open, settled bool) {
	layer := ui.Box(c.Key("now-playing-layer")).Absolute().Top(0).Left(0).Right(0).Bottom(0)
	pose := ui.Pose{Y: 60, Fade: 1}
	if open {
		pose = ui.Pose{}
	}
	layer.AnimatePose(pose, npSpring)
	if !open && settled || d.Player.Current() == nil {
		// Closed: nothing to paint or click.
		// Its backdrop goes with it: the canvas, and the GPU
		// texture it holds, would otherwise stay for the rest of
		// the session, invisible.
		d.ClearBackdrop()
		layer.Invisible()
		d.ResetPanes()
		return
	}
	if !open {
		// On its way out, the window under it takes the clicks.
		layer.PassThrough()
	}
	layer.Children(func() { View(c, d) })
}

// View fills the window with what plays: its cover over a slow drift of its
// own colors, the controls, and lyrics that follow the song.
func View(c *ui.Context, d Deps) {
	pl := d.Player
	tr := pl.Current()
	w, h := c.Size()
	art := d.Art.Get(qqmusic.Sized(tr.Album.PicURL, 600), true)
	// The colors stay the last song's while its cover loads, rather than
	// going gray between the two.
	shown := art
	if art == nil && tr.Album.PicURL != "" {
		shown = *d.NPArt
	}
	*d.NPArt = shown
	tint := theme.Neutral(true, 9)
	if shown != nil && !shown.Failed {
		tint = shown.Tint
	}
	accent := tint.Mix(white, 0.45)
	// The light of the line sung takes the cover's own color, lightened as
	// the lyrics beside it are.
	*d.GlowColor = accent

	root := ui.Box(c.Key("now-playing")).Absolute().Top(0).Left(0).Right(0).Bottom(0).
		Background(theme.Neutral(true, 1)).TextColor(white)
	// The pointer is over the view whenever it is in the window.
	pointerIn := root.Hovered()
	root.Children(func() {
		d.Backdrop(c, shown, tint)
		// Over the backdrop, which is positioned too: positioned elements
		// paint over those in the flow.
		ui.Column(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Children(func() { body(c, d, tr, art, accent, w, h, pointerIn) })
	})
}

func body(c *ui.Context, d Deps, tr *qqmusic.Track, art *art.Artwork, accent ui.Color, w, h float32, pointerIn bool) {
	pl := d.Player
	d.CommentsFor(tr)
	// An episode has what it is about and its captions in place of lyrics;
	// a song, lyrics in place of those.
	switch program := tr.ProgramID != 0; {
	case program && *d.Pane == PaneWall:
		*d.Pane = PaneQueue
	case program && (*d.Pane == PaneLyrics || *d.Pane == PaneStage):
		*d.Pane = PaneProgram
		if d.CaptionsOn() {
			*d.Pane = PaneCaptions
		}
	case program && *d.Pane == PaneCaptions && !d.CaptionsShown():
		*d.Pane = PaneProgram
	case !program && (*d.Pane == PaneProgram || *d.Pane == PaneCaptions):
		*d.Pane = d.LyricsPaneChosen()
	}
	if *d.Pane != PaneQueue {
		d.ResetQueue()
	}
	if *d.Pane != PaneWall {
		d.ResetWall()
	}
	{
		// The top strip drags the window; the chevron closes the view,
		// showing only while the pointer is in the window, as the pane tabs do.
		ui.Row(c).Height(d.TopBar(c)).Justify(ui.Center).AlignItems(ui.Center).DragWindow().Children(func() {
			Reveal(ui.Box(c.Key("close-reveal")), pointerIn).Children(func() {
				if collapseButton(c.Key("close")).Clicked() {
					*d.NowPlaying = false
				}
			})
		})
		if *d.Pane == PaneStage {
			// The stage takes the whole view: the lyrics are the show, the
			// controls a slim bar at the foot.
			Reveal(TabsRow(c, w, h, d.TopBar(c)), pointerIn).Children(func() {
				d.PaneTabs(c)
				stageStyleButton(c, d)
			})
			Stage(c, d, tr, accent)
			stageBar(c, d, tr, art, accent)
			return
		}
		if *d.Pane == PaneWall {
			// The wall takes the whole view too, under the song playing.
			WallPane(c, d, tr, art, accent, w, h, pointerIn)
			return
		}

		coverSize := npCoverSize(w, h)
		// The cover and what plays start at one edge, with the column
		// wide enough for the controls even when the cover is small.
		side := max(coverSize, 280)
		padX := npSideX(w)
		Reveal(TabsRow(c, w, h, d.TopBar(c)), pointerIn).Children(func() { d.PaneTabs(c) })
		ui.Row(c).Grow(1).Padding(0, padX, 44, padX).Gap(72).AlignItems(ui.Stretch).Children(func() {
			var coverTop float32
			ui.Column(c).Width(side).Justify(ui.Center).Gap(0).Children(func() {
				coverTop = ui.Box(c.Key("cover-top")).Height(0).LayoutBounds().Y
				Cover(c, pl, d.CoverState, tr, art, coverSize)

				ui.Row(c).Gap(8).AlignItems(ui.Center).Margin(36, 0, 18, 0).Children(func() {
					ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(4).Children(func() {
						if r := pl.Radio(); r != "" {
							d.RadioChip(c, r)
						}
						ui.Text(c, tr.Name).FontSize(22).FontWeight(700).LetterSpacing(-0.3).SingleLine()
						d.ArtistLinks(c, tr, white.Alpha(0.6), white, true).FontSize(15).SingleLine()
					})
					d.Like(c.Key("like"), tr, 22, white.Alpha(0.7), appui.MediaSurface)
					d.SongMenu(c.Key("more"), tr, 22, white.Alpha(0.7), appui.MediaSurface).Margin(0, -8, 0, 0)
				})
				transport.Scrubber(c.Key("np-scrub"), pl, d.Settings.ProgressEffectMode(), d.NpScrub, white.Alpha(0.18), white.Alpha(0.9), 16, accent.Mix(white, 0.35))
				// As tall with the quality's badge as without it: the badge goes
				// while the next song loads, and the column would move with it.
				ui.Row(c).Height(16).AlignItems(ui.Center).Children(func() {
					ui.Text(c, appui.Clock(pl.Position())).FontSize(11).FontFeatures("tnum").TextColor(white.Alpha(0.5))
					ui.Spacer(c)
					qualityBadge(c, pl.Stream())
					ui.Spacer(c)
					status := "-" + appui.Clock(pl.Duration()-pl.Position())
					if pl.Loading() {
						status = "正在缓冲…"
					}
					ui.Text(c, status).FontSize(11).FontFeatures("tnum").TextColor(white.Alpha(0.5))
				})
				ui.Box(c).Margin(14, 0, 18, 0).Children(func() {
					playerbar.Transport(c, pl, white, white.Alpha(0.45), accent, 28, appui.MediaSurface)
				})
				ui.Row(c).Justify(ui.Center).Children(func() {
					playerbar.VolumeSlider(c, d.Volume, d.MutedFrom, pl, white.Alpha(0.7), white.Alpha(0.18), 180, appui.MediaSurface)
				})
			})

			pane := ui.Column(c.Key("pane")).Grow(1).Basis(0).MinWidth(0)
			pane.Children(func() {
				// The pane starts level with the cover beside it, but below the
				// tabs, which start level with the pane's top: they come and go
				// with the pointer, so they sit out of the flow, never over it.
				b := pane.LayoutBounds()
				if b.H > 0 && coverTop > 0 {
					if top := max(coverTop-b.Y, TabsHeight+8); top != *d.PaneTop {
						*d.PaneTop = top
						c.AnimationFrame()
					}
				} else {
					c.AnimationFrame() // not laid out yet: the next frame places it
				}
				ui.Box(c).Height(*d.PaneTop)
				switch *d.Pane {
				case PaneLyrics:
					LyricsPane(c, d, tr, h)
				case PaneComments:
					CommentsPane(c, d)
				case PaneProgram:
					ProgramPane(c, tr)
				case PaneCaptions:
					CaptionsPane(c, d, tr, h)
				default:
					QueuePane(c, d)
				}
			})
		})
	}
}

// npMaxWidth is how wide the cover and the pane beside it go, at most.
const npMaxWidth = 1280

// npCoverSize leaves room for the controls and the strip above the cover.
func npCoverSize(w, h float32) float32 {
	return min(float32(math.Min(float64(w)*0.36, 440)), h-436)
}

// npSideX is how far in from the sides of a view w wide the cover and the
// pane beside it are: in a wide window they keep together in the middle,
// rather than the pane taking all the width the cover leaves.
func npSideX(w float32) float32 {
	return max(72, (w-npMaxWidth)/2)
}

// TabsHeight is how tall the row of Now Playing's tabs is, for what is
// under it to start below it.
const TabsHeight = 28

// collapseButton uses translucent media controls over the cover's dark
// backdrop, independently of the appearance of the pages underneath it.
func collapseButton(c *ui.Context) ui.Element {
	b := ui.ButtonBase(c).Size(40, 32).Center().Radius(12).Label("收起播放页").Tooltip("收起播放页")
	appui.MediaFeedback(b, theme.Veil(10))
	b.Children(func() { ui.Icon(c, appui.IconDown).Size(20, 20) })
	return b
}

// TabsRow is the row of Now Playing's tabs, in the same place for every
// pane, above the right-hand content. Anchoring their start keeps them
// still as the comments' count changes or the stage's style button appears.
func TabsRow(c *ui.Context, w, h float32, top float32) ui.Element {
	left := npSideX(w) + max(npCoverSize(w, h), 280) + 72
	return ui.Row(c.Key("tabs")).Absolute().Top(top).Left(left).Height(TabsHeight).Gap(6).AlignItems(ui.Center)
}

// Reveal fades e in while the pointer is in the window, and out when it
// leaves.
func Reveal(e ui.Element, shown bool) ui.Element {
	pose := ui.Pose{Fade: 1}
	if shown {
		pose = ui.Pose{}
	}
	return e.AnimatePose(pose, ui.Animation{Duration: 200 * time.Millisecond})
}

// qualityBadge marks a lossless stream, as it plays.
func qualityBadge(c *ui.Context, st *qqmusic.Stream) {
	if st == nil || !qqmusic.Lossless(st.Level) {
		return
	}
	label := "无损"
	if st.Level == "hires" {
		label = "Hi-Res"
	}
	ui.Text(c, label).FontSize(9.5).FontWeight(800).LetterSpacing(0.6).TextColor(theme.Warm(true, 11)).
		Padding(1, 6).Radius(4).Border(1, theme.Warm(true, 8))
}
