package app

import (
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/audio"
	"github.com/egoist/mygo/ui"

	"lilt/internal/lyrics"
	"lilt/internal/lyricsui"
	"lilt/internal/qqmusic"
)

// Locking removes the hover adjustment plate so only the lyrics remain over the desktop.
func TestDesktopLyricsLockedPreservesTransparency(t *testing.T) {
	_, d := desktopLyricsViewFixture()
	tt := ui.NewTester(d.View, 760, 180)
	tt.Move(200, 80)
	tt.Frame()
	frame := tt.Image()
	if _, _, _, a := frame.At(8, 100).RGBA(); a == 0 {
		t.Fatal("unlocked hover did not reveal the adjustment plate")
	}
	if _, _, _, a := frame.At(380, 20).RGBA(); a == 0 {
		t.Fatal("hover did not reveal the toolbar")
	}
	d.SetLocked(true)
	tt.Frame()
	frame = tt.Image()
	if _, _, _, a := frame.At(8, 100).RGBA(); a != 0 {
		t.Fatal("locked lyrics kept the full-window adjustment plate")
	}
	if _, _, _, a := frame.At(380, 20).RGBA(); a != 0 {
		t.Fatal("locked lyrics kept the playback toolbar")
	}
}

func desktopLyricsViewFixture() (*app, *desktopLyrics) {
	a := newTestApp()
	a.settings = defaultSettings()
	a.settings.DesktopLyrics = true
	a.pl.Cue([]qqmusic.Track{{ID: 7, Name: "测试歌曲", Duration: 60000}})
	a.pl.SetLyrics(lyrics.Build("[00:01]我们一起听音乐", "", ""))
	a.pl.Seek(2 * time.Second)
	d := lyricsui.NewDesktop(a.lyricsDeps())
	a.desktopLyrics = d
	return a, d
}

// Toolbar clicks adjust typography, lock through an external recovery path, and close only lyrics.
func TestDesktopLyricsViewActions(t *testing.T) {
	a, d := desktopLyricsViewFixture()
	tt := ui.NewTester(d.View, 760, 180)
	tt.Move(200, 80)
	tt.Frame()
	before := d.FontSize()
	if err := tt.Click("增大桌面歌词字号"); err != nil {
		t.Fatal(err)
	}
	if d.FontSize() != before+2 {
		t.Fatal("font increase did not reach settings")
	}
	if err := tt.Click("减小桌面歌词字号"); err != nil {
		t.Fatal(err)
	}
	if d.FontSize() != before {
		t.Fatal("font decrease did not reach settings")
	}
	if err := tt.Click("锁定桌面歌词"); err != nil {
		t.Fatal(err)
	}
	if !a.settings.DesktopLyricsLocked {
		t.Fatal("lock did not reach the window owner")
	}
	d.SetLocked(false)
	tt.Frame()
	tt.Move(200, 80)
	tt.Frame()
	if err := tt.Click("关闭桌面歌词"); err != nil {
		t.Fatal(err)
	}
	if a.settings.DesktopLyrics || a.pl.Current() == nil {
		t.Fatal("close changed playback or left lyrics enabled")
	}
}

// The floating player's pause button acts on the same player without opening the main window.
func TestDesktopLyricsViewPause(t *testing.T) {
	a, d := desktopLyricsViewFixture()
	out := audio.NewPlayer(&silence{})
	out.SetVolume(0)
	defer out.Close()
	a.pl.Attach(out)
	tt := ui.NewTester(d.View, 760, 180)
	tt.Move(200, 80)
	tt.Frame()
	if err := tt.Click("播放/暂停"); err != nil {
		t.Fatal(err)
	}
	if a.pl.Playing() || a.nowPlaying {
		t.Fatal("toolbar did not pause independently")
	}
}

// Dragging the width grip changes width and does not activate a playback control.
func TestDesktopLyricsViewResize(t *testing.T) {
	a, d := desktopLyricsViewFixture()
	tt := ui.NewTester(d.View, 760, 180)
	tt.Move(750, 100)
	tt.Frame()
	tt.Press(750, 100)
	tt.Move(710, 100)
	tt.Release(710, 100)
	if d.Width >= lyricsui.DesktopLyricsWidth {
		t.Fatalf("drag left did not narrow the window: %v", d.Width)
	}
	if a.pl.Playing() {
		t.Fatal("resize activated playback")
	}
}

// Exported lyric frames are inspected visually instead of asserting on typography pixels.
func TestDesktopLyricsFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip("set NM_FRAMES to inspect desktop lyrics")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, scene := range []struct {
		name, text, translation, qrc string
		width                        int
		dark                         bool
		position                     time.Duration
		hover                        bool
		locked                       bool
		doubleLine                   bool
	}{
		{name: "chinese-light", text: "[00:01]我们一起听音乐", width: 760, position: 2 * time.Second},
		{name: "chinese-hover-dark", text: "[00:01]我们一起听音乐", width: 760, dark: true, hover: true, position: 2 * time.Second},
		{name: "chinese-locked-dark", text: "[00:01]我们一起听音乐", width: 760, dark: true, hover: true, locked: true, position: 2 * time.Second},
		{name: "japanese-dark", text: "[00:01]君をのせて 空へ続く道", translation: "[00:01]伴随着你，走向天空的路", width: 760, dark: true, position: 2 * time.Second},
		{name: "chinese-double-dark", text: "[00:01]我们一起听音乐\n[00:12]让声音陪伴远方", width: 760, dark: true, doubleLine: true, position: 2 * time.Second},
		{name: "chinese-double-long", text: "[00:01]我们一起听音乐\n[00:12]让音乐陪伴每一个普通的日子，走过漫长的道路直到远方", width: 480, doubleLine: true, position: 2 * time.Second},
		{name: "japanese-double-dark", text: "[00:01]君をのせて 空へ続く道\n[00:12]光る地平線", translation: "[00:01]伴随着你，走向天空的路\n[00:12]闪耀的地平线", width: 760, dark: true, doubleLine: true, position: 2 * time.Second},
		{name: "long-light", text: "[00:01]我们一起走过漫长的道路，让音乐陪伴每一个普通的日子直到远方", width: 480, position: 2 * time.Second},
		{name: "word-hold-dark", text: "[00:01]慢唱", qrc: "[1000,7000]慢(1000,3000)唱(6000,2000)", width: 760, dark: true, position: 2 * time.Second},
		{name: "interlude-dark", text: "[00:01]Stay with me\n[00:12]Next line", translation: "[00:01]留在我身边\n[00:12]下一句译文", qrc: "[1000,2000]Stay with me(1000,2000)\n[12000,2000]Next line(12000,2000)", width: 760, dark: true, position: 6 * time.Second},
	} {
		t.Run(scene.name, func(t *testing.T) {
			a, d := desktopLyricsViewFixture()
			a.settings.DesktopLyricsLocked = scene.locked
			a.settings.DesktopLyricsDoubleLine = scene.doubleLine
			a.pl.SetLyrics(lyrics.Build(scene.text, scene.translation, scene.qrc))
			a.pl.Seek(scene.position)
			tt := ui.NewTester(func(c *ui.Context) {
				ui.Box(c).Fill().Background(neutral(scene.dark, 2)).Children(func() { d.View(c) })
			}, scene.width, 240)
			tt.Move(-10, -10)
			if scene.hover {
				tt.Move(200, 80)
			}
			tt.Advance(time.Second)
			file, err := os.Create(filepath.Join(dir, "desktop-lyrics-"+scene.name+".png"))
			if err != nil {
				t.Fatal(err)
			}
			if err := png.Encode(file, tt.Image()); err != nil {
				file.Close()
				t.Fatal(err)
			}
			if err := file.Close(); err != nil {
				t.Fatal(err)
			}
		})
	}
}
