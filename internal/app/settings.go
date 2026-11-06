package app

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/egoist/mygo"

	settingspkg "lilt/internal/settings"
)

// settings are the listener's choices; the model and its persistence live in
// lilt/internal/settings while the views still live here.
type settings = settingspkg.Settings

var defaultSettings = settingspkg.Default

var loadSettings = func() settings {
	s := settingspkg.Load()
	normalizeLyrics(&s)
	return s
}

const (
	volumeBalanceOff      = settingspkg.VolumeBalanceOff
	volumeBalanceQuiet    = settingspkg.VolumeBalanceQuiet
	volumeBalanceStandard = settingspkg.VolumeBalanceStandard
	volumeBalanceHigh     = settingspkg.VolumeBalanceHigh

	progressStandard = settingspkg.ProgressStandard
	progressSmoke    = settingspkg.ProgressSmoke
	progressSquiggly = settingspkg.ProgressSquiggly
)

// oldName is the app's previous name; older NetEase data stays separate.
const oldName = "Lilt"

// moveOldData moves the user data and the song cache kept under oldName to
// the directories of the app's name, unless those hold something already:
// the login, settings and where playback was left go on. It must run before
// anything reads them.
func moveOldData() {
	previous := oldName
	if strings.HasSuffix(mygo.App.Name(), " Dev") {
		previous += " Dev"
	}
	for _, name := range []mygo.PathName{mygo.PathUserData, mygo.PathCache} {
		dir, err := mygo.App.Path(name)
		if err != nil {
			continue
		}
		moveDir(filepath.Join(filepath.Dir(dir), previous), dir)
	}
}

// moveDir moves the directory old to dir when dir is empty, as Path made
// it, and old exists.
func moveDir(old, dir string) {
	if old == dir {
		return
	}
	if entries, err := os.ReadDir(dir); err != nil || len(entries) > 0 {
		return
	}
	if _, err := os.Stat(old); err != nil {
		return
	}
	if os.Remove(dir) == nil {
		if err := os.Rename(old, dir); err != nil {
			os.MkdirAll(dir, 0o700)
		}
	}
}
