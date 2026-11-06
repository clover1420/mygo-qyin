package app

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"

	"lilt/internal/nowplaying"
	"lilt/internal/qqmusic"
)

// Now Playing's shell, cover and panes live in lilt/internal/nowplaying;
// the app keeps its listener's state and these names while its other views
// still live here.
type coverState = nowplaying.CoverState
type wallState = nowplaying.WallState
type lyricGlide = nowplaying.Glide

// coverKey keys a cover's element by its song and its fling.
type coverKey struct {
	id  int64
	gen int
}

const (
	paneLyrics   = nowplaying.PaneLyrics
	paneQueue    = nowplaying.PaneQueue
	paneStage    = nowplaying.PaneStage
	paneComments = nowplaying.PaneComments
	paneProgram  = nowplaying.PaneProgram
	paneCaptions = nowplaying.PaneCaptions
	paneWall     = nowplaying.PaneWall

	wallPitch = nowplaying.WallPitch
)

var (
	stageStyles = nowplaying.StageStyles
	stageUnits  = nowplaying.StageUnits
	splitParts  = nowplaying.SplitParts
)

// nowplayingDeps gathers what Now Playing needs, once a build.
func (a *app) nowplayingDeps() nowplaying.Deps {
	return nowplaying.Deps{
		Player:            a.pl,
		Settings:          &a.settings,
		Art:               a.art,
		NpScrub:           &a.npScrub,
		NowPlaying:        &a.nowPlaying,
		Open:              &a.npOpen,
		ChangedAt:         &a.npChangedAt,
		Pane:              &a.npPane,
		PaneTop:           &a.npPaneTop,
		NPArt:             &a.npArt,
		GlowColor:         &a.glowColor,
		CoverState:        &a.npCoverState,
		Volume:            &a.volume,
		MutedFrom:         &a.mutedFrom,
		QueueList:         &a.queueList,
		QueueAt:           &a.queueAt,
		LyricScroll:       &a.lyricScroll,
		LyricGlide:        &a.lyricGlide,
		LyricAt:           &a.lyricAt,
		LyricTrack:        &a.lyricTrack,
		TopBar:            topBar,
		SetButtonsTheme:   a.npButtonsTheme,
		ClearBackdrop:     func() { a.backdropCache = backdropCache{} },
		ResetPanes:        func() { a.queueAt, a.wallState.Track = -1, 0 },
		ResetQueue:        func() { a.queueAt = -1 },
		ResetWall:         func() { a.wallState.Track = 0 },
		Backdrop:          a.backdrop,
		CommentsFor:       func(tr *qqmusic.Track) { a.commentsFor(tr) },
		CaptionsOn:        a.captionsOn,
		CaptionsShown:     a.captionsShown,
		LyricsPaneChosen:  a.lyricsPaneChosen,
		PaneTabs:          a.paneTabs,
		RadioChip:         a.radioChip,
		ArtistLinks:       a.artistLinks,
		Like:              a.likeButton,
		SongMenu:          a.songMenu,
		Cover:             a.cover,
		SongItems:         a.songItems,
		Avatar:            a.avatar,
		QuietShown:        a.quietShown,
		QuietCard:         a.quietCard,
		Skeleton:          func(c *ui.Context) { lyricsSkeleton(c, npShimmer(c)) },
		StageQuiet:        a.stageQuiet,
		StageSkeleton:     func(c *ui.Context, f, w float32) { stageSkeleton(c, npShimmer(c), f, w) },
		BackdropOver:      a.backdropOver,
		Wall:              &a.wallState,
		Comments:          commentFeed{a},
		LoadComments:      a.loadComments,
		ToggleCommentLike: a.toggleCommentLike,
		CommentSkeleton: func(c *ui.Context, n int) {
			s := npShimmer(c)
			s.text(c, 13, 1.3).Width(56).Margin(18, 12, 6, 12)
			commentSkeleton(c, s, n)
		},
		MoreSkeleton:        func(c *ui.Context) { commentSkeleton(c.Key("more"), npShimmer(c), 2) },
		Accent:              a.pal.Accent,
		Captions:            &a.captions,
		SyncCaptions:        a.syncCaptions,
		StartModelDownload:  a.startModelDownload,
		CancelModelDownload: a.cancelModelDownload,
	}
}

func (a *app) lyricsPane(c *ui.Context, tr *qqmusic.Track, h float32) {
	nowplaying.LyricsPane(c, a.nowplayingDeps(), tr, h)
}

func (a *app) queuePane(c *ui.Context) {
	nowplaying.QueuePane(c, a.nowplayingDeps())
}

func (a *app) queueTracks(c *ui.Context, list *ui.ListState, lastAt *int, p palette) ui.Element {
	return nowplaying.QueueTracks(c, nowplaying.QueueDeps{
		Player: a.pl, Cover: a.cover, ArtistLinks: a.artistLinks, SongItems: a.songItems,
	}, list, lastAt, p)
}

func (a *app) npState(c *ui.Context) (open, settled bool) {
	return nowplaying.State(c, a.nowplayingDeps())
}

func (a *app) nowPlayingLayer(c *ui.Context, open, settled bool) {
	nowplaying.Layer(c, a.nowplayingDeps(), open, settled)
}

// npButtonsTheme lights the window buttons over Now Playing's dark backdrop,
// and as the theme over the pages.
func (a *app) npButtonsTheme(dark bool) {
	if a.win == nil {
		return
	}
	theme := mygo.ThemeSystem
	if dark {
		theme = mygo.ThemeDark
	}
	a.win.SetWindowButtonsTheme(theme)
}
