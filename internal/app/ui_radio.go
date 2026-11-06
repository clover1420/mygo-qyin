package app

import (
	"math/rand/v2"
	"strings"

	"lilt/internal/media/qqmusic"
)

// Radios play on without end: the personal FM, in one of its modes, as
// "fm:<mode>", and the heart mode of the liked songs, as "heart".

const heartRadio = "heart"

var fmModes = []struct{ mode, label string }{
	{qqmusic.FMDefault, "推荐"},
}

func fmRadio(mode string) string { return "fm:" + mode }

// radioLabel names a radio as the listener knows it.
func radioLabel(name string) string {
	if name == heartRadio {
		return "心动模式"
	}
	for _, m := range fmModes {
		if name == fmRadio(m.mode) {
			return "电台 · " + m.label
		}
	}
	return "电台"
}

func playable(tracks []qqmusic.Track) []qqmusic.Track {
	var out []qqmusic.Track
	for _, t := range tracks {
		if t.Playable() {
			out = append(out, t)
		}
	}
	return out
}

// likedPlaylist returns the ID of the listener's liked songs, or 0.
func (a *app) likedPlaylist() int64 {
	for _, pl := range a.mine {
		if pl.SpecialType == 5 {
			return pl.ID
		}
	}
	return 0
}

// radioFeed returns what gives the radio name more tracks: it loads them
// on a goroutine, and gives them to got on the main thread, none if it
// cannot.
func (a *app) radioFeed(name string) func(got func([]qqmusic.Track)) {
	return func(got func([]qqmusic.Track)) {
		var fetch func() ([]qqmusic.Track, error)
		switch {
		case strings.HasPrefix(name, "fm:"):
			mode := strings.TrimPrefix(name, "fm:")
			fetch = func() ([]qqmusic.Track, error) { return qqmusic.PersonalFM(mode) }
		case name == heartRadio:
			// More like the last song queued, from the song it began with.
			q, pid := a.pl.Queue(), a.likedPlaylist()
			if len(q) == 0 || pid == 0 {
				got(nil)
				return
			}
			start, last := q[0].ID, q[len(q)-1].ID
			fetch = func() ([]qqmusic.Track, error) { return qqmusic.HeartMode(last, start, pid) }
		default:
			got(nil)
			return
		}
		go func() {
			// The FM gives a song or three at a time, some of them VIP
			// songs the listener may not play: ask again for those.
			var tracks []qqmusic.Track
			for tries := 0; tries < 3 && len(tracks) == 0; tries++ {
				more, err := fetch()
				if err != nil {
					break
				}
				tracks = playable(more)
			}
			a.post(func() { got(tracks) })
		}()
	}
}

// playFM plays the personal FM in the mode chosen, or goes on playing it.
func (a *app) playFM() {
	if a.profile == nil {
		a.openLogin()
		return
	}
	name := fmRadio(fmModes[a.fmMode].mode)
	if a.pl.Radio() == name {
		if !a.pl.Playing() {
			a.pl.TogglePlay()
		}
		return
	}
	a.radioStarting = name
	feed := a.radioFeed(name)
	feed(func(tracks []qqmusic.Track) {
		if a.radioStarting != name {
			return
		}
		a.radioStarting = ""
		if len(tracks) == 0 {
			a.toast = "电台 暂时无法播放，请稍后再试"
			return
		}
		a.pl.PlayRadio(name, tracks, feed)
	})
}

// exitRadio stops the radio playing, but not the song it plays.
func (a *app) exitRadio() {
	name := a.pl.Radio()
	if name == "" {
		return
	}
	a.pl.EndRadio()
	label := radioLabel(name)
	if name != heartRadio {
		label = "电台"
	}
	a.toast = "已退出" + label
}

// trashFM skips the song playing on the FM, which plays it no more.
func (a *app) trashFM() {
	tr := a.pl.Current()
	if tr == nil {
		return
	}
	id := tr.ID
	a.pl.Next()
	go qqmusic.FMTrash(id)
}

// startHeartMode plays songs like one the listener likes: the one playing,
// if they like it, or else one of their liked songs at random.
func (a *app) startHeartMode() {
	if a.profile == nil {
		a.showMain()
		a.openLogin()
		return
	}
	pid := a.likedPlaylist()
	if pid == 0 {
		a.toast = "喜欢一些歌曲后，即可开启心动模式"
		return
	}
	if a.pl.Radio() == heartRadio {
		a.nowPlaying = true
		return
	}
	var seed *qqmusic.Track
	fromPlaying := false
	if cur := a.pl.Current(); cur != nil && a.liked[cur.ID] {
		s := *cur
		seed, fromPlaying = &s, true
	}
	liked := a.playlists[pid]
	a.radioStarting = heartRadio
	go func() {
		if seed == nil {
			if liked == nil {
				liked, _ = qqmusic.PlaylistDetail(pid)
			}
			if liked != nil {
				if ok := playable(liked.Tracks); len(ok) > 0 {
					seed = &ok[rand.IntN(len(ok))]
				}
			}
		}
		var more []qqmusic.Track
		var err error
		if seed != nil {
			more, err = qqmusic.HeartMode(seed.ID, seed.ID, pid)
		}
		a.post(func() {
			if a.radioStarting != heartRadio {
				return
			}
			a.radioStarting = ""
			if seed == nil || err != nil {
				a.toast = "心动模式暂时无法开启，请稍后再试"
				return
			}
			more = playable(more)
			if fromPlaying && a.pl.Current() != nil && a.pl.Current().ID == seed.ID {
				a.pl.RadioFrom(heartRadio, more, a.radioFeed(heartRadio))
				return
			}
			a.pl.PlayRadio(heartRadio, append([]qqmusic.Track{*seed}, more...), a.radioFeed(heartRadio))
		})
	}()
}
