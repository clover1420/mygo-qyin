package app

// openSettings shows the settings as a pane over the main window, where
// their glass has the app to show through: a window of its own would have
// nothing under it.
func (a *app) openSettings() {
	a.measureCache()
	// Over the pages, not over Now Playing, which it puts away as a page of
	// its own would.
	a.nowPlaying = false
	a.settingsOpen = true
	a.showMain()
	a.redraw()
}

// changed applies the settings everywhere and keeps them.
func (a *app) changed() {
	a.settings.Save()
	a.settings.ApplyAppearance()
	if a.notch != nil {
		a.notch.SetEnabled(a.settings.Notch)
	}
	a.redraw()
}
