//go:build !windows

package app

func desktopLyricsNativeSnapshot(a *app) map[string]any {
	out := map[string]any{"enabled": a.settings.DesktopLyrics, "locked": a.settings.DesktopLyricsLocked, "mainVisible": a.win.IsVisible()}
	if d := a.desktopLyrics; d != nil && d.Win != nil {
		out["visible"] = d.Win.IsVisible()
	}
	return out
}
