package pages

import (
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/appui"
	"lilt/internal/art"
	"lilt/internal/player"
	"lilt/internal/qqmusic"
	"lilt/internal/theme"
)

// RadioMode is one way the personal FM plays.
type RadioMode struct{ Mode, Label string }

// RadioDeps are what the radio page needs from the app it runs in.
type RadioDeps struct {
	Palette theme.Palette
	Art     *art.Cache
	Player  *player.Player
	Modes   []RadioMode
	FMIndex *int
	Scroll  *ui.ScrollState

	Insets    func(*ui.Context) (top, bottom float32)
	StartMark func(*ui.Context)
	EndMark   func(*ui.Context)

	Radio         func() string
	RadioLabel    func(string) string
	RadioStarting func() string
	StartingHeart func() bool
	SignedIn      bool
	OnHeart       bool

	PlayFM      func()
	ExitRadio   func()
	StartHeart  func()
	ModeChanged func()

	Like        func(*ui.Context, *qqmusic.Track)
	ArtistLinks func(*ui.Context, *qqmusic.Track, ui.Color, ui.Color, ...bool) ui.Element
	CoverFill   func(*ui.Context, string, int, float32)
}

// FMPage is the personal FM: what it plays, large, over its cover's
// colors, with the FM's own controls, and heart mode below.
func FMPage(c *ui.Context, d RadioDeps) {
	top, bottom := d.Insets(c)
	ui.Scroll(c.Key("fm")).TrackScroll(d.Scroll).Grow(1).Padding(top+PageGap, Pad, bottom+48, Pad).ScrollbarInsets(top, 0, bottom, 0).Children(func() {
		d.StartMark(c)
		defer d.EndMark(c)
		// The mode the FM plays in, if it plays.
		for i, m := range d.Modes {
			if d.Radio() == "fm:"+m.Mode {
				*d.FMIndex = i
			}
		}
		ui.Row(c).AlignItems(ui.End).Children(func() {
			ui.Box(c).Grow(1).Children(func() { Heading(c, d.Palette, "电台", "无尽推荐") })
			mode := *d.FMIndex
			ui.Box(c).Margin(0, 0, 22, 0).Children(func() {
				labels := make([]string, len(d.Modes))
				for i, m := range d.Modes {
					labels[i] = m.Label
				}
				if appui.Segmented(c.Key("fm-mode"), &mode, labels...) {
					*d.FMIndex = mode
					// Playing the FM, it changes to the mode at once.
					if strings.HasPrefix(d.Radio(), "fm:") {
						d.ModeChanged()
					}
				}
			})
		})
		fmCard(c, d)
	})
}

// fmCard shows the song the FM plays, or how to start it.
func fmCard(c *ui.Context, d RadioDeps) {
	pl := d.Player
	onFM := strings.HasPrefix(d.Radio(), "fm:")
	var tr *qqmusic.Track
	if onFM {
		tr = pl.Current()
	}
	var cover *art.Artwork
	if tr != nil {
		cover = d.Art.Get(qqmusic.Sized(tr.Album.PicURL, 600), true)
	}
	blurred := backdropBitmap(cover)
	card := ui.Box(c.Key("fm-card")).Height(320).Radius(26).Background(theme.Neutral(true, 3)).TextColor(theme.Paper).
		Shadow(0, 18, 44, -16, d.Palette.CardShade)
	in := appui.ArtIn(card, blurred)
	card.Children(func() {
		if in < 1 {
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(26).
				LinearGradient(ui.LinearGradient{Angle: 135, Oklab: true, From: theme.Accent(true, 6), To: theme.Neutral(true, 2)})
		}
		appui.ArtImage(c, blurred, 24, in)
		ui.Box(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Radius(26).
			LinearGradient(ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{
				{Color: theme.Ink.Alpha(0.35), At: 0},
				{Color: theme.Ink.Alpha(0.1), At: 1},
			}})
		ui.Row(c).Absolute().Top(0).Left(0).Right(0).Bottom(0).Padding(36, 40).Gap(40).AlignItems(ui.Center).Children(func() {
			fmCovers(c, d, tr)
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
				label := "电台 · " + d.Modes[*d.FMIndex].Label
				if onFM {
					label = d.RadioLabel(d.Radio())
				}
				ui.Row(c).Gap(6).AlignItems(ui.Center).Children(func() {
					ui.Icon(c, appui.IconRadio).Size(14, 14).TextColor(theme.Paper.Alpha(0.75))
					ui.Text(c, label).FontSize(12).FontWeight(600).TextColor(theme.Paper.Alpha(0.75))
				})
				if tr == nil {
					ui.Text(c, "听见更多好音乐").FontSize(30).FontWeight(800).LetterSpacing(-0.6)
					ui.Text(c, "根据你的口味，一首接一首，不停推荐").FontSize(13.5).TextColor(theme.Paper.Alpha(0.75))
					ui.Row(c).Margin(18, 0, 0, 0).Children(func() {
						label := "开始收听"
						if d.RadioStarting() != "" && strings.HasPrefix(d.RadioStarting(), "fm:") {
							label = "正在调频…"
						}
						if appui.PillButton(c.Key("start"), appui.IconPlay, label, theme.Paper, theme.Neutral(false, 12)).Clicked() {
							d.PlayFM()
						}
					})
					return
				}
				ui.Text(c, tr.Name).FontSize(30).FontWeight(800).LetterSpacing(-0.6).MaxLines(2).LineHeight(1.2)
				switch {
				case pl.Loading():
					ui.Text(c, "正在缓冲…").FontSize(14).TextColor(theme.Paper.Alpha(0.7)).SingleLine()
				case pl.Err() != "":
					ui.Text(c, pl.Err()).FontSize(14).TextColor(theme.Paper).SingleLine()
				default:
					d.ArtistLinks(c, tr, theme.Paper.Alpha(0.7), theme.Paper, true).FontSize(14).SingleLine()
				}
				ui.Row(c).Gap(14).AlignItems(ui.Center).Margin(20, 0, 0, -8).Children(func() {
					d.Like(c.Key("like"), tr)
					play := ui.ButtonBase(c.Key("play")).Size(56, 56).Center().Radius(999).Background(theme.Paper).TextColor(theme.Neutral(false, 12)).Label("播放/暂停")
					pose := ui.Pose{}
					if play.Pressed() {
						pose.Scale = 0.92
					} else if play.Hovered() {
						pose.Scale = 1.05
					}
					play.AnimatePose(pose, ui.Snappy)
					play.Children(func() { appui.PlayPause(c, play, pl.Playing(), 22) })
					if play.Clicked() {
						pl.TogglePlay()
					}
					if appui.IconButton(c.Key("next"), appui.IconNext, 20, theme.Paper.Alpha(0.8), "下一首").Clicked() {
						pl.Next()
					}
				})
			})
		})
	})
}

// fmCovers stacks the cover playing over the next one's, which peeks out
// behind it, tilted, as the FM's own player shows them.
func fmCovers(c *ui.Context, d RadioDeps, tr *qqmusic.Track) {
	const size = 232
	ui.Box(c).Size(size+36, size).Shrink(0).Children(func() {
		var next *qqmusic.Track
		if tr != nil {
			if order, at := d.Player.PlayOrder(); at >= 0 && at+1 < len(order) {
				next = &d.Player.Queue()[order[at+1]]
			}
		}
		if next != nil {
			back := ui.Box(c.Key(next.ID)).Absolute().Top(16).Left(36).Size(size-32, size-32).Radius(14).Opacity(0.55).Background(theme.Paper.Alpha(0.12)).
				Shadow(0, 10, 24, -8, theme.Ink.Alpha(0.4))
			back.AnimatePose(ui.Pose{Rotate: 6, X: 14}, ui.Animation{Spring: ui.SpringOptions{Duration: 500 * time.Millisecond, Bounce: 0.2}})
			back.Children(func() { d.CoverFill(c, next.Album.PicURL, 400, 14) })
		}
		front := ui.Box(c.Key("front")).Absolute().Top(0).Left(0).Size(size, size).Radius(20).
			Background(theme.Paper.Alpha(0.12)).Shadow(0, 20, 44, -10, theme.Ink.Alpha(0.55))
		if tr == nil {
			front.Center().Children(func() {
				ui.Icon(c, appui.IconRadio).Size(64, 64).TextColor(theme.Paper.Alpha(0.85))
			})
			return
		}
		scale := float32(1)
		if !d.Player.Playing() {
			scale = 0.94
		}
		front.AnimatePose(ui.Pose{Scale: scale}, ui.Animation{Spring: ui.SpringOptions{Duration: 600 * time.Millisecond, Bounce: 0.25}})
		front.Children(func() {
			ui.Box(c.Key(tr.ID)).Fill().Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.92}}).
				Children(func() { d.CoverFill(c, tr.Album.PicURL, 600, 16) })
		})
	})
}

// HeartCard starts heart mode: songs like those the listener likes.
func HeartCard(c *ui.Context, d RadioDeps) {
	p := d.Palette
	on := d.OnHeart
	ui.Row(c).Padding(22, 26).Gap(20).Radius(20).AlignItems(ui.Center).Background(p.Hover).Children(func() {
		ui.Box(c).Size(56, 56).Radius(14).Center().Shrink(0).
			LinearGradient(ui.LinearGradient{Angle: 135, Oklab: true, From: theme.Accent(false, 9), To: theme.Accent(false, 10)}).
			TextColor(theme.Paper).Children(func() { ui.Icon(c, appui.IconHeartFill).Size(26, 26) })
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, "从你喜欢的音乐出发").FontSize(15).FontWeight(700)
			ui.Text(c, "根据你红心过的歌曲，智能推荐风格相似的音乐，与你的喜欢穿插播放").FontSize(12.5).TextColor(p.Muted)
		})
		if on {
			if appui.PillButton(c.Key("exit-heart"), appui.IconX, "退出心动模式", p.Press.Over(p.Bg), p.Text).Clicked() {
				d.ExitRadio()
			}
		}
		label := "开启心动模式"
		switch {
		case !d.SignedIn:
			label = "登录后开启"
		case on:
			label = "正在播放"
		case d.StartingHeart():
			label = "正在开启…"
		}
		if appui.PillButton(c.Key("heart"), appui.IconHeartFill, label, p.Accent, p.OnAccent).Clicked() {
			d.StartHeart()
		}
	})
}
