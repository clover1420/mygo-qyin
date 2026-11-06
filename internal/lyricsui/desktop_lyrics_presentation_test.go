package lyricsui

import (
	"image/png"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/lyrics"
	"lilt/internal/theme"
)

// Instrumental gaps used to clear the current line; retain it until its successor starts.
func TestDesktopLyricTimeline(t *testing.T) {
	lines := lyrics.Build("", "[00:01]译文", "[1000,2000]一(1000,1000)二(2000,1000)\n[8000,2000]三(8000,2000)")
	for _, tc := range []struct {
		pos   time.Duration
		text  string
		index int
		next  time.Duration
	}{
		{0, "", -1, time.Second}, {time.Second, "一二", 0, 7 * time.Second}, {3 * time.Second, "一二", 0, 5 * time.Second},
		{8 * time.Second, "三", 1, 0}, {10 * time.Second, "三", 1, 0}, {2 * time.Second, "一二", 0, 6 * time.Second},
	} {
		got := desktopLyricAt(lines, true, tc.pos, false)
		if got.text != tc.text || got.index != tc.index || got.next != tc.next {
			t.Errorf("position %s: text=%q index=%d next=%s", tc.pos, got.text, got.index, got.next)
		}
	}
	if got := desktopLyricAt(lines, false, time.Second, false); got.index >= 0 || got.text == lines[0].Text {
		t.Fatal("loading retained previous track lyrics")
	}
	if got := desktopLyricAt(nil, true, 0, false); got.index >= 0 {
		t.Fatal("missing lyrics retained a singing line")
	}
}

// A Han-only Japanese line keeps its translation when another line establishes kana.
func TestDesktopLyricTranslationPolicy(t *testing.T) {
	for _, tc := range []struct {
		name  string
		lines []lyrics.Line
		want  bool
	}{
		{"Chinese with English", []lyrics.Line{{Text: "我们一起 say goodbye"}}, false},
		{"Japanese kana later", []lyrics.Line{{Text: "愛"}, {Text: "君をのせて"}}, true},
		{"Japanese shinjitai", []lyrics.Line{{Text: "学校 国体 来世 会場 没頭"}}, true},
		{"unknown traditional Han", []lyrics.Line{{Text: "山河歲月"}}, true},
		{"Korean", []lyrics.Line{{Text: "사랑해"}}, true},
		{"English", []lyrics.Line{{Text: "Let it be"}}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := DesktopTranslation(tc.lines); got != tc.want {
				t.Fatalf("show translation=%t, want %t", got, tc.want)
			}
		})
	}
	line := lyrics.Line{At: time.Second, End: 3 * time.Second, Text: "山河岁月", Translation: "山河，岁月。"}
	if got := desktopLyricAt([]lyrics.Line{line}, true, 2*time.Second, false); got.translation != "" {
		t.Fatal("duplicate translation was displayed")
	}
}

// Real word gaps wait for the next word while active words request bounded highlight updates.
func TestDesktopLyricWordScheduling(t *testing.T) {
	line := lyrics.Line{At: time.Second, End: 8 * time.Second, Text: "慢唱", Timed: true, Words: []lyrics.Word{{Text: "慢", Start: time.Second, End: 3 * time.Second}, {Text: "唱", Start: 5 * time.Second, End: 8 * time.Second}}}
	for _, tc := range []struct{ position, want time.Duration }{{2 * time.Second, 40 * time.Millisecond}, {4 * time.Second, time.Second}, {8 * time.Second, 0}} {
		content := desktopLyricAt([]lyrics.Line{line}, true, tc.position, false)
		if got := desktopLyricWake(content, tc.position); got != tc.want {
			t.Errorf("position=%s wake=%s want=%s", tc.position, got, tc.want)
		}
	}
	line.Timed = false
	if got := desktopLyricWake(desktopLyricAt([]lyrics.Line{line}, true, 2*time.Second, false), 2*time.Second); got != 0 {
		t.Fatalf("LRC used estimated word timing: %s", got)
	}
}

// Line.End used to clear both texts during accompaniment; keep them bright until the next line.
func TestDesktopLyricInterludeRetainsTranslation(t *testing.T) {
	lines := lyrics.Build("", "[00:01]留在我身边\n[00:10]下一句译文", "[1000,2000]Stay with me(1000,2000)\n[10000,2000]Next line(10000,2000)")
	content := desktopLyricAt(lines, true, 9*time.Second, false)
	if content.index != 0 || content.text != "Stay with me" || content.translation != "留在我身边" {
		t.Fatalf("instrumental gap cleared the line or translation: %+v", content)
	}
	if content.line.Words[0].Progress(9*time.Second) != 1 || desktopLyricWake(content, 9*time.Second) != time.Second {
		t.Fatal("finished words did not stay bright while waiting for the next line")
	}
	if next := desktopLyricAt(lines, true, 10*time.Second, false); next.index != 1 || next.translation != "下一句译文" {
		t.Fatalf("successor did not replace the retained line: %+v", next)
	}
}

// A backward or paused seek discards a pending crossfade instead of showing the old line.
func TestDesktopLyricSeekCancelsTransition(t *testing.T) {
	start := time.Unix(0, 0)
	first := desktopLyricContent{index: 0, text: "first"}
	second := desktopLyricContent{index: 1, text: "second"}
	var state desktopLyricViewState
	state.step(1, "", 0, first, time.Second, start, true, false)
	state.step(1, "", 0, second, 2*time.Second, start.Add(time.Second), true, false)
	if state.previous.text != "first" || state.changed.IsZero() {
		t.Fatal("ordinary line change did not start its transition")
	}
	state.step(1, "", 0, first, time.Second, start.Add(1100*time.Millisecond), true, false)
	if state.previous.text != "" || !state.changed.IsZero() || state.current.text != "first" {
		t.Fatal("backward seek kept an obsolete transition")
	}
	state.step(1, "", 0, second, 2*time.Second, start.Add(2*time.Second), true, false)
	state.step(1, "", 0, first, time.Second, start.Add(2100*time.Millisecond), false, false)
	if state.previous.text != "" || !state.changed.IsZero() {
		t.Fatal("paused seek kept an obsolete transition")
	}
	state.step(2, "", 0, second, 0, start.Add(3*time.Second), true, false)
	if state.previous.text != "" || !state.changed.IsZero() {
		t.Fatal("track switch retained old content")
	}
}

// A short forward seek across a line cancels crossfade even below the clock-drift threshold.
func TestDesktopLyricSmallSeekCancelsTransition(t *testing.T) {
	start := time.Unix(0, 0)
	var state desktopLyricViewState
	state.step(1, "", 0, desktopLyricContent{index: 0, text: "first"}, 1900*time.Millisecond, start, true, false)
	state.step(1, "", 1, desktopLyricContent{index: 1, text: "second"}, 2100*time.Millisecond, start.Add(20*time.Millisecond), true, false)
	if state.previous.text != "" || !state.changed.IsZero() || state.current.text != "second" {
		t.Fatal("short forward seek retained an obsolete line")
	}
}

// Reduced motion cancels an in-flight sentence transition instead of leaving a fading old line.
func TestDesktopLyricReducedMotionCancelsTransition(t *testing.T) {
	start := time.Unix(0, 0)
	first := desktopLyricContent{index: 0, text: "first"}
	second := desktopLyricContent{index: 1, text: "second"}
	var state desktopLyricViewState
	state.step(1, "", 0, first, time.Second, start, true, false)
	state.step(1, "", 0, second, 2*time.Second, start.Add(time.Second), true, false)
	if state.previous.text == "" {
		t.Fatal("fixture did not begin a sentence transition")
	}
	state.step(1, "", 0, second, 2050*time.Millisecond, start.Add(1050*time.Millisecond), true, true)
	if state.previous.text != "" || !state.changed.IsZero() || state.current.text != "second" {
		t.Fatal("reduced motion kept the old sentence or its transition")
	}
}

// Export the actual lyric layers across a sentence change for visual inspection of overlap and motion.
func TestDesktopLyricsTransitionFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip("set NM_FRAMES to inspect lyric transitions")
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	for _, scene := range []struct{ name, previous, current, translation string }{
		{"chinese", "风吹过遥远的山海", "我们一起听音乐", ""},
		{"shorter", "让每个平凡瞬间化作旋律，陪伴我们走过漫长的夜晚", "直到远方", ""},
		{"translation", "Stay with me through the night", "Let the music carry us home", "让音乐陪伴我们回家"},
	} {
		t.Run(scene.name, func(t *testing.T) {
			var state desktopLyricViewState
			state.previous = desktopLyricContent{text: scene.previous}
			started := false
			tt := ui.NewTester(func(c *ui.Context) {
				if !started {
					state.changed, started = c.Now(), true
				}
				previous := DesktopLayouts(c, desktopLyricContent{text: scene.previous}, 0, 32, 432)
				current := DesktopLayouts(c, desktopLyricContent{text: scene.current, translation: scene.translation}, 0, 32, 432)
				blend := state.blend(c)
				ui.Box(c).Fill().Background(theme.Neutral(false, 2)).PaddingX(24).Children(func() {
					drawDesktopLyricLayers(c, current, previous, 140, blend)
				})
			}, 480, 168)
			tt.Frame()
			elapsed := time.Duration(0)
			for _, ms := range []int{0, 40, 80, 120, 180, 240, 480} {
				at := time.Duration(ms) * time.Millisecond
				tt.Advance(at - elapsed)
				elapsed = at
				f, err := os.Create(filepath.Join(dir, "desktop-lyrics-transition-"+scene.name+"-"+strconv.Itoa(ms)+"ms.png"))
				if err != nil {
					t.Fatal(err)
				}
				err = png.Encode(f, tt.Image())
				closeErr := f.Close()
				if err != nil || closeErr != nil {
					t.Fatalf("save transition frame: %v / %v", err, closeErr)
				}
			}
		})
	}
}
