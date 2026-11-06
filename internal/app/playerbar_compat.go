package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/playerbar"
)

// The player bar lives in lilt/internal/playerbar; the app keeps its
// listener's state and these names while its views still live here.
type pickedUp = playerbar.Picked

const playerBarHeight = playerbar.Height

type popover = playerbar.Popover

var popoverColors = playerbar.PopoverColors

// playerbarDeps gathers what the bar needs, once a frame.
func (a *app) playerbarDeps() playerbar.Deps {
	return playerbar.Deps{
		Player:                 a.pl,
		Settings:               &a.settings,
		Palette:                a.pal,
		Volume:                 &a.volume,
		VolumeOpen:             &a.volumeOpen,
		MutedFrom:              &a.mutedFrom,
		Picked:                 &a.picked,
		BarScrub:               &a.barScrub,
		QueueOpen:              &a.queueOpen,
		QueuePopup:             &a.queuePopup,
		QueuePopupAt:           &a.queuePopupAt,
		UnderBottom:            a.underBottom,
		Cover:                  a.cover,
		Like:                   a.likeButton,
		ArtistLinks:            a.artistLinks,
		ProgressFill:           a.progressFill,
		Pane:                   a.pane,
		Ink:                    a.playerbarInk,
		QueueTracks:            a.queueTracks,
		TopBar:                 topBar,
		OpenNowPlaying:         func() { a.nowPlaying = true },
		SetLyricsEnabled:       a.setLyricsEnabled,
		SetDesktopLyricsLocked: a.setDesktopLyricsLocked,
	}
}

func (a *app) playerbarInk(onDark bool) playerbar.Ink {
	ink := a.inkOn(onDark)
	return playerbar.Ink{Text: ink.text, Muted: ink.muted}
}

func (a *app) playerBar(c *ui.Context) { playerbar.Bar(c, a.playerbarDeps()) }

func (a *app) transport(c *ui.Context, color, dim, on ui.Color, size float32, surface ...buttonSurface) {
	playerbar.Transport(c, a.pl, color, dim, on, size, surface...)
}

func (a *app) volumeButton(c *ui.Context, color ui.Color) {
	playerbar.VolumeButton(c, &a.volume, &a.volumeOpen, a.pl, a.pal, color)
}

func (a *app) volumeSlider(c *ui.Context, color, track ui.Color, width float32, surface ...buttonSurface) {
	playerbar.VolumeSlider(c, &a.volume, &a.mutedFrom, a.pl, color, track, width, surface...)
}

func (a *app) nativeVolumeSlider(c *ui.Context, pop popover) {
	playerbar.NativeVolumeSlider(c, &a.volume, a.pl, pop, a.pal.Accent)
}

func (a *app) scrollVolume(ev ui.InputEvent) bool {
	return playerbar.ScrollVolume(&a.volume, a.pl, ev)
}
