package nowplaying

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
)

// CommentFeed is the comments of the song or episode Now Playing shows, as
// the app holds them.
type CommentFeed interface {
	Hot() []qqmusic.Comment
	Latest() []qqmusic.Comment
	Total() int
	More() bool
	Loaded() bool
	Loading() bool
	Err() string
	Scroll() *ui.ScrollState
}

type commentKey struct {
	hot bool
	id  int64
}

// CommentsPane lists the comments of the song playing: the hot ones, then
// the latest, which load on as they scroll into view.
func CommentsPane(c *ui.Context, d Deps) {
	cs := d.Comments
	if len(cs.Hot()) == 0 && len(cs.Latest()) == 0 {
		if cs.Loading() || !cs.Loaded() && cs.Err() == "" {
			ui.Column(c).Grow(1).Padding(4, 8, 0, 0).Gap(2).Children(func() {
				d.CommentSkeleton(c, 6)
			})
			return
		}
		ui.Column(c).Grow(1).Center().Gap(14).Children(func() {
			switch {
			case cs.Err() != "":
				ui.Text(c, cs.Err()).FontSize(15).TextColor(white.Alpha(0.5))
				if appui.PillButton(c.Key("retry"), appui.IconComment, "重试", white.Alpha(0.14), white).Clicked() {
					d.LoadComments()
				}
			default:
				ui.Icon(c, appui.IconComment).Size(40, 40).TextColor(white.Alpha(0.3))
				ui.Text(c, "还没有评论，快来抢沙发").FontSize(17).FontWeight(600).TextColor(white.Alpha(0.45))
			}
		})
		return
	}
	now := c.Now()
	scroll := ui.Scroll(c.Key("comments")).Grow(1).TrackScroll(cs.Scroll()).Padding(4, 8, 48, 0)
	var end ui.Rect
	scroll.Children(func() {
		ui.Column(c).Gap(2).Children(func() {
			if hot := cs.Hot(); len(hot) > 0 {
				commentHeading(c, "精彩评论")
				for i := range hot {
					commentRow(c, d, &hot[i], now)
				}
			}
			commentHeading(c, "最新评论 "+appui.Count(cs.Total()))
			latest := cs.Latest()
			for i := range latest {
				commentRow(c, d, &latest[i], now)
			}
			if cs.Loading() {
				d.MoreSkeleton(c)
			}
			ui.Row(c.Key("foot")).Height(56).Center().Children(func() {
				switch {
				case cs.Loading():
				case cs.Err() != "":
					retry := ui.ButtonBase(c.Key("retry")).Padding(6, 14).Radius(999).Background(white.Alpha(0.1)).Label("重试")
					retry.Children(func() {
						ui.Text(c, "加载失败，点按重试").FontSize(12.5).TextColor(white.Alpha(0.7))
					})
					if retry.Clicked() {
						d.LoadComments()
					}
				case !cs.More():
					ui.Text(c, "没有更多评论了").FontSize(12).TextColor(white.Alpha(0.35))
				}
			})
			end = ui.Box(c.Key("end")).Height(0).Bounds()
		})
	})
	// The next page, once the foot is within a screen of coming into view.
	if box := scroll.Bounds(); cs.More() && !cs.Loading() && cs.Err() == "" && end.W > 0 && end.Y < box.Y+box.H*2 {
		d.LoadComments()
	}
}

func commentHeading(c *ui.Context, title string) {
	ui.Text(c, title).FontSize(13).FontWeight(700).TextColor(white.Alpha(0.55)).LetterSpacing(0.2).Padding(18, 12, 6, 12)
}

// commentRow is a comment: who wrote it, what, and what it replies to,
// when, from where, and how many liked it, with a thumb to like it too.
func commentRow(c *ui.Context, d Deps, cm *qqmusic.Comment, now time.Time) {
	ui.Row(c).Padding(12).Gap(12).Radius(14).AlignItems(ui.Start).Children(func() {
		d.Avatar(c, cm.User.AvatarURL, 34, white.Alpha(0.1))
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(6).Children(func() {
			ui.Row(c).Gap(8).AlignItems(ui.Center).Children(func() {
				ui.Text(c, cm.User.Nickname).FontSize(13).FontWeight(600).TextColor(white.Alpha(0.6)).SingleLine().Shrink(1).MinWidth(0)
				ui.Spacer(c)
				if cm.LikedCount > 0 {
					ui.Text(c, appui.Count(cm.LikedCount)+" 赞").FontSize(12).TextColor(white.Alpha(0.6))
				}
			})
			ui.Text(c, cm.Content).FontSize(15).LineHeight(1.55).TextColor(white.Alpha(0.92))
			if len(cm.BeReplied) > 0 {
				r := cm.BeReplied[0]
				ui.Box(c).Padding(8, 12).Radius(12).Background(white.Alpha(0.07)).Children(func() {
					ui.RichText(c,
						ui.Span{Text: "@" + r.User.Nickname + "：", Color: white.Alpha(0.75), Weight: 600},
						ui.Span{Text: r.Content, Color: white.Alpha(0.6)},
					).FontSize(13).LineHeight(1.5).MaxLines(3)
				})
			}
			meta := appui.RelativeTime(time.UnixMilli(cm.Time), now)
			if loc := cm.IPLocation.Location; loc != "" {
				meta += " · " + loc
			}
			ui.Text(c, meta).FontSize(11.5).TextColor(white.Alpha(0.38))
		})
	})
}

// commentLike is a comment's thumb, with how many liked it, which likes it
// or unlikes it: in red once liked.
func commentLike(c *ui.Context, d Deps, cm *qqmusic.Comment) {
	label, fg := "赞", white.Alpha(0.42)
	if cm.Liked {
		label, fg = "取消赞", d.Accent
	}
	b := ui.ButtonBase(c).Gap(4).Padding(3, 6).Margin(-3, -6).Radius(999).AlignItems(ui.Center).FocusRing(false).Label(label)
	if b.Hovered() && !cm.Liked {
		fg = white.Alpha(0.75)
	}
	b.TextColor(b.AnimateColor("fg", fg, ui.Animation{Duration: 120 * time.Millisecond}))
	b.Children(func() {
		icon := appui.IconThumb
		if cm.Liked {
			icon = appui.IconThumbFill
		}
		ui.Icon(c, icon).Size(13, 13)
		if cm.LikedCount > 0 {
			ui.Text(c, appui.Count(cm.LikedCount)).FontSize(12).FontFeatures("tnum")
		}
	})
	if b.Clicked() {
		d.ToggleCommentLike(cm.ID)
	}
}
