package app

import "lilt/internal/ui/features/lyricsui"

const (
	lyricsDesktop = lyricsui.DesktopMode
	lyricsTaskbar = lyricsui.TaskbarMode
)

// lyricsMode remembers the chosen location even while lyrics are switched off.
func lyricsMode(s settings) string { return lyricsui.Mode(s) }

func lyricsEnabled(s settings) bool { return lyricsui.Enabled(s) }

// normalizeLyrics migrates independent switches to a single supported location.
func normalizeLyrics(s *settings) {
	enabled := lyricsEnabled(*s)
	s.LyricsMode = lyricsMode(*s)
	s.DesktopLyrics = enabled && s.LyricsMode == lyricsDesktop
	s.TaskbarLyrics = enabled && s.LyricsMode == lyricsTaskbar
}

func (a *app) setLyricsMode(mode string) {
	if mode == lyricsDesktop && supportsDesktopLyrics() || mode == lyricsTaskbar && supportsTaskbarLyrics() {
		a.applyLyricsMode(mode, lyricsEnabled(a.settings))
	}
}

func (a *app) setLyricsEnabled(enabled bool) {
	if !supportsDesktopLyrics() {
		return
	}
	a.applyLyricsMode(lyricsMode(a.settings), enabled)
}

func (a *app) applyLyricsMode(mode string, enabled bool) {
	a.settings.LyricsMode = mode
	a.settings.DesktopLyrics = enabled && mode == lyricsDesktop
	a.settings.TaskbarLyrics = enabled && mode == lyricsTaskbar
	a.syncDesktopLyrics()
	a.syncTaskbarLyrics()
	a.settings.Save()
	a.refreshLyricsTray()
	a.redraw()
}
