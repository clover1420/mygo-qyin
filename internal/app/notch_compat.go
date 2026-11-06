package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/player"
	"lilt/internal/media/qqmusic"
	notchpkg "lilt/internal/ui/features/notch"
)

// The notch player lives in lilt/internal/ui/features/notch; the app keeps its
// listener's state and these names while its views still live here.
type notch = notchpkg.Notch

const (
	notchOpenW   = notchpkg.OpenW
	notchIdleW   = notchpkg.IdleW
	notchHelloW  = notchpkg.HelloW
	notchLyricsW = notchpkg.LyricsW
	notchOpenH   = notchpkg.OpenH
)

func newNotch(a *app) *notch { return notchpkg.New(a.notchDeps()) }

// notchDeps gathers what the notch needs, once at startup.
func (a *app) notchDeps() notchpkg.Deps {
	return notchpkg.Deps{
		Player:      a.pl,
		Settings:    &a.settings,
		Art:         a.art,
		Palette:     a.pal,
		Quitting:    &a.quitting,
		SyncWindows: a.syncWindows,
		ShowMain:    a.showMain,
		OpenNowPlaying: func() {
			a.nowPlaying = true
			a.showMain()
			if a.win != nil {
				a.win.Invalidate()
			}
		},
		Like: func(c *ui.Context, tr *qqmusic.Track, size float32, color ui.Color) ui.Element {
			return a.likeButton(c, tr, size, color)
		},
		ArtistLinks: func(c *ui.Context, tr *qqmusic.Track, color, hover ui.Color) ui.Element {
			return a.artistLinks(c, tr, color, hover)
		},
	}
}

// windowFrame returns a window's frame, for checking where it went.
func windowFrame(handle uintptr) (x, y, w, h float64) {
	return notchpkg.WindowFrame(handle)
}

// syncWindows redraws both windows when what plays changed in one of
// them, so the other shows it too.
func (a *app) syncWindows() {
	pl := a.pl
	var id int64
	if tr := pl.Current(); tr != nil {
		id = tr.ID
	}
	sig := playerSig{id, pl.Playing(), pl.Loading(), pl.Shuffle(), pl.Repeat(), a.volume, a.nowPlaying, pl.Err()}
	if sig != a.sig {
		a.sig = sig
		a.redraw()
		if sig.playing && !sig.loading && a.notch != nil {
			a.notch.Greet()
		}
	}
}

type playerSig struct {
	track      int64
	playing    bool
	loading    bool
	shuffle    bool
	repeat     player.RepeatMode
	volume     float64
	nowPlaying bool
	err        string
}
