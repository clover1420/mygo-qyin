package app

import (
	"github.com/egoist/mygo"

	"lilt/internal/ui/features/lyricsui"
)

// The lyrics windows live in lilt/internal/ui/features/lyricsui; the app keeps these
// names for its controller fields and tests.
type desktopLyrics = lyricsui.Desktop
type taskbarLyrics = lyricsui.Taskbar
type taskbarLyricsPlacement = lyricsui.Placement

func (a *app) refreshLyricsTray() {
	if a.tray != nil {
		a.tray.SetMenu(a.trayMenu())
	}
}

func supportsDesktopLyrics() bool { return lyricsui.SupportsDesktopLyrics() }
func supportsTaskbarLyrics() bool { return lyricsui.SupportsTaskbarLyrics() }

func (a *app) lyricsDeps() lyricsui.Deps {
	return lyricsui.Deps{
		Player:      a.pl,
		Settings:    &a.settings,
		Art:         a.art,
		Post:        a.post,
		Redraw:      a.redraw,
		Quitting:    func() bool { return a.quitting },
		MainWindow:  func() *mygo.Window { return a.win },
		RefreshTray: a.refreshLyricsTray,
		SetEnabled:  a.setDesktopLyricsEnabled,
	}
}

func (a *app) setDesktopLyricsEnabled(enabled bool) {
	if !supportsDesktopLyrics() {
		return
	}
	if enabled {
		a.applyLyricsMode(lyricsDesktop, true)
	} else if a.settings.DesktopLyrics {
		a.setLyricsEnabled(false)
	}
}

func (a *app) setDesktopLyricsLocked(locked bool) {
	if !supportsDesktopLyrics() || !a.settings.DesktopLyrics {
		return
	}
	if a.desktopLyrics == nil {
		a.syncDesktopLyrics()
	}
	if a.desktopLyrics != nil {
		a.desktopLyrics.SetLocked(locked)
	}
}

func (a *app) syncDesktopLyrics() {
	if !supportsDesktopLyrics() {
		return
	}
	if a.desktopLyrics == nil && a.settings.DesktopLyrics {
		a.desktopLyrics = lyricsui.NewDesktop(a.lyricsDeps())
	}
	d := a.desktopLyrics
	if d == nil {
		return
	}
	if !a.settings.DesktopLyrics {
		d.Destroy()
		return
	}
	d.Sync()
}

func (a *app) setTaskbarLyricsEnabled(enabled bool) {
	if !supportsTaskbarLyrics() {
		return
	}
	if enabled {
		a.applyLyricsMode(lyricsTaskbar, true)
	} else if a.settings.TaskbarLyrics {
		a.setLyricsEnabled(false)
	}
}

func (a *app) syncTaskbarLyrics() {
	if !supportsTaskbarLyrics() {
		return
	}
	if !a.settings.TaskbarLyrics || a.quitting {
		if d := a.taskbarLyrics; d != nil {
			a.taskbarLyrics = nil
			d.Destroy()
		}
		return
	}
	if a.taskbarLyrics == nil {
		a.taskbarLyrics = lyricsui.NewTaskbar(a.lyricsDeps())
	}
	a.taskbarLyrics.Sync()
}
