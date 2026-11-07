package app

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/media/art"
	"lilt/internal/media/qqmusic"
)

// These load a page of a thread's comments, and like a comment or unlike
// it, with NetEase; tests replace them.
var (
	threadComments = qqmusic.ThreadComments
	likeComment    = qqmusic.LikeComment
)

// commentState is the comments of the song or episode Now Playing shows:
// the hot ones, and the latest, loaded a page at a time as they scroll.
type commentState struct {
	track   int64
	thread  string // whose comments they are, as NetEase names it
	hot     []qqmusic.Comment
	latest  []qqmusic.Comment
	total   int
	more    bool
	loaded  bool
	loading bool
	err     string
	scroll  ui.ScrollState
}

// commentPage is how many comments a page loads.
const commentPage = 30

// commentsFor makes the comments those of tr, loading their first page
// once it changed: Now Playing does it as it shows, for the count on the
// pane's tab.
func (a *app) commentsFor(tr *qqmusic.Track) *commentState {
	cs := &a.comments
	if cs.track != tr.ID {
		*cs = commentState{track: tr.ID, thread: tr.Thread()}
		a.loadComments()
	}
	return cs
}

// loadComments loads the next page of comments.
func (a *app) loadComments() {
	cs := &a.comments
	if cs.loading {
		return
	}
	thread, offset := cs.thread, len(cs.latest)
	cs.loading, cs.err = true, ""
	go func() {
		page, err := threadComments(thread, offset, commentPage)
		a.post(func() {
			if cs.thread != thread {
				return
			}
			cs.loading = false
			if err != nil {
				cs.err = "评论加载失败"
				return
			}
			cs.loaded = true
			if offset == 0 {
				cs.hot = page.Hot
			}
			// New comments shift the pages: those already shown come again.
			seen := make(map[int64]bool, len(cs.latest))
			for _, cm := range cs.latest {
				seen[cm.ID] = true
			}
			for _, cm := range page.Latest {
				if !seen[cm.ID] {
					cs.latest = append(cs.latest, cm)
				}
			}
			cs.total, cs.more = page.Total, page.More && len(page.Latest) > 0
		})
	}()
}

// commentsTab names the comments' pane, with how many there are.
func (a *app) commentsTab() string {
	if n := a.commentCount(); n > 0 {
		return "评论 " + count(n)
	}
	return "评论"
}

// commentCount is the count the comments' tab shows: the last song's
// while the next one's loads, so that it rolls from one to the other.
func (a *app) commentCount() int {
	switch cs := &a.comments; {
	case cs.loaded:
		return cs.total
	case cs.err != "":
		return 0
	}
	return a.tabCount.n
}

// commentFeed is the comments as the Now Playing views read them.
type commentFeed struct{ a *app }

func (f commentFeed) Hot() []qqmusic.Comment    { return f.a.comments.hot }
func (f commentFeed) Latest() []qqmusic.Comment { return f.a.comments.latest }
func (f commentFeed) Total() int                { return f.a.comments.total }
func (f commentFeed) More() bool                { return f.a.comments.more }
func (f commentFeed) Loaded() bool              { return f.a.comments.loaded }
func (f commentFeed) Loading() bool             { return f.a.comments.loading }
func (f commentFeed) Err() string               { return f.a.comments.err }
func (f commentFeed) Scroll() *ui.ScrollState   { return &f.a.comments.scroll }

// rollState is a count as it rolls from one value to another.
type rollState struct {
	n    int
	up   bool       // whether it went up, last it changed
	exit *ui.Motion // where the characters shown go as it changes
}

// rollLine is how far a character rolls, about a line of the tab's text.
const rollLine = 16

// rollCount shows n as text, each character in a slot of its own, as
// the wheels of a counter: one that changes rolls out as the next rolls
// in, upward as n grows and downward as it shrinks, the others staying.
func rollCount(c *ui.Context, st *rollState, n int, text string, fg ui.Color) {
	if n != st.n {
		st.up = n > st.n
		dir := float32(1)
		if !st.up {
			dir = -1
		}
		if st.exit != nil {
			// The characters shown leave from this frame on, which first
			// knows the way.
			st.exit.Y = -dir * rollLine
		}
		st.n, st.exit = n, &ui.Motion{}
	}
	dir := float32(1)
	if !st.up {
		dir = -1
	}
	chars := []rune(text)
	// Clipped to its line, a character leaves out of sight.
	ui.Row(c).ClipY().Children(func() {
		for i, ch := range chars {
			// Slots go by their place from the right, as the wheels of a
			// counter do: the units are always the last.
			k := len(chars) - 1 - i
			roll := ui.Animation{Spring: ui.SpringOptions{Duration: 420 * time.Millisecond, Bounce: 0.15}, Delay: time.Duration(k) * 30 * time.Millisecond}
			slot := ui.Box(c.Key(k)).Transition(ui.ElementTransition{
				Animation: &roll, Size: true,
				Enter: &ui.Motion{Y: dir * rollLine, Collapse: true},
				Exit:  &ui.Motion{Collapse: true},
			})
			slot.Children(func() {
				ui.Text(c.Key(ch), string(ch)).FontSize(13).FontWeight(600).FontFeatures("tnum").TextColor(fg).
					Transition(ui.ElementTransition{Animation: &roll, Colors: true, Enter: &ui.Motion{Y: dir * rollLine}, Exit: st.exit})
			})
		}
	})
}

// toggleCommentLike likes a comment, or unlikes it, at once in the app and
// then with NetEase, which may refuse: it goes back then. Signed out, it
// asks the listener to sign in.
func (a *app) toggleCommentLike(id int64) {
	if a.profile == nil {
		a.openLogin()
		return
	}
	cs := &a.comments
	thread, like := cs.thread, true
	if cm := cs.find(id); cm != nil {
		like = !cm.Liked
	}
	cs.setLiked(id, like)
	go func() {
		err := likeComment(thread, id, like)
		a.post(func() {
			if err != nil && cs.thread == thread {
				cs.setLiked(id, !like)
				a.toast = "操作失败，请稍后再试"
			}
		})
	}()
}

// find returns the comment id, hot or latest, or nil.
func (cs *commentState) find(id int64) *qqmusic.Comment {
	for _, list := range [][]qqmusic.Comment{cs.hot, cs.latest} {
		for i := range list {
			if list[i].ID == id {
				return &list[i]
			}
		}
	}
	return nil
}

// setLiked marks the comment id liked, or not, counting the listener in or
// out, both where it is hot and where it is latest.
func (cs *commentState) setLiked(id int64, like bool) {
	for _, list := range [][]qqmusic.Comment{cs.hot, cs.latest} {
		for i := range list {
			if cm := &list[i]; cm.ID == id && cm.Liked != like {
				cm.Liked = like
				if like {
					cm.LikedCount++
				} else {
					cm.LikedCount = max(cm.LikedCount-1, 0)
				}
			}
		}
	}
}

// avatar shows a listener's picture, round, over empty while it loads.
func (a *app) avatar(c *ui.Context, url string, size float32, empty ui.Color) {
	bmp := bitmap(a.art.Get(qqmusic.Sized(url, int(size*3)), art.Cover))
	box := ui.Box(c).Size(size, size).Radius(size / 2).Shrink(0)
	in := artIn(box, bmp)
	if in < 1 {
		box.Background(empty)
	}
	box.Children(func() { artImage(c, bmp, size/2, in) })
}
