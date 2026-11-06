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

var volumeBalances = []struct{ id, label string }{
	{volumeBalanceOff, "关闭"},
	{volumeBalanceQuiet, "安静"},
	{volumeBalanceStandard, "标准"},
	{volumeBalanceHigh, "高音量"},
}

var progressEffects = []struct{ id, label string }{
	{progressStandard, "标准"},
	{progressSmoke, "烟雾"},
	{progressSquiggly, "波浪"},
}

// qualityLevel is the stream level QQ Music names a quality by.
var qualityLevels = []struct{ level, label string }{
	{"standard", "标准（128k）"},
	{"exhigh", "极高（320k）"},
	{"lossless", "无损（FLAC）"},
}

// cacheLimits are the sizes the song cache may take, in megabytes.
var cacheLimits = []struct {
	mb    int
	label string
}{
	{0, "关闭"},
	{1024, "1 GB"},
	{2048, "2 GB"},
	{5120, "5 GB"},
	{10240, "10 GB"},
}

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
