package lyricsui

import "lilt/internal/core/settings"

// The two places desktop lyrics show: over the desktop, or in the taskbar.
const (
	DesktopMode = "desktop"
	TaskbarMode = "taskbar"
)

// Mode remembers the chosen location even while lyrics are switched off.
// Before the location preference existed, an enabled taskbar selected that mode.
func Mode(s settings.Settings) string {
	if SupportsTaskbarLyrics() && (s.LyricsMode == TaskbarMode || s.LyricsMode == "" && s.TaskbarLyrics) {
		return TaskbarMode
	}
	return DesktopMode
}

// Enabled reports whether the lyrics show at all, where the platform can.
func Enabled(s settings.Settings) bool {
	return SupportsDesktopLyrics() && s.DesktopLyrics || SupportsTaskbarLyrics() && s.TaskbarLyrics
}
