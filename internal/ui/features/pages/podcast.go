package pages

import (
	"fmt"
	"strings"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/core/navigation"
	"lilt/internal/media/player"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
)

// PodcastInfo is what the app has of a podcast.
type PodcastInfo struct {
	Podcast     *qqmusic.Podcast
	Loading     bool
	Err         string
	Programs    []qqmusic.Program
	PageLoaded  bool
	PageLoading bool
	PageMore    bool
	PageErr     string
}

// PodcastDeps are what a podcast's page needs from the app it runs in.
type PodcastDeps struct {
	Palette theme.Palette
	Route   *navigation.Route
	Player  *player.Player

	Info            func() PodcastInfo
	Load            func(int64)
	ProgramsMore    func(int64)
	PlayProgram     func(int)
	ToggleSubscribe func()

	Centered   func(*ui.Context, string, func())
	Skeleton   func(*ui.Context)
	ScrollPage func(*ui.Context, func())
	PageFoot   func(*ui.Context, func())
	Empty      func(*ui.Context, string)

	Cover     func(*ui.Context, string, int, float32, float32) ui.Element
	CoverFill func(*ui.Context, string, int, float32)
	Open      func(navigation.Route)
	Link      func(*ui.Context, string, ui.Color, ui.Color, navigation.Route)
	Bars      func(*ui.Context, ui.Color)
	Like      func(*ui.Context, *qqmusic.Track)
	Liked     func(int64) bool
}

// PodcastPage shows a podcast and its episodes.
func PodcastPage(c *ui.Context, d PodcastDeps, id int64) {
	p := d.Palette
	info := d.Info()
	if info.Podcast == nil {
		if !info.Loading {
			d.Centered(c, info.Err, func() { d.Load(id) })
			return
		}
		d.Skeleton(c)
		return
	}
	pod := info.Podcast
	d.ScrollPage(c, func() {
		ui.Row(c).Gap(28).AlignItems(ui.End).Padding(0, 0, 8, 0).Children(func() {
			ui.Box(c).Size(212, 212).Radius(14).Shadow(0, 18, 40, -10, p.CardShade).Children(func() {
				d.CoverFill(c, pod.PicURL, 600, 14)
			})
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
				kind := "播客"
				if pod.Category != "" {
					kind += " · " + pod.Category
				}
				ui.Text(c, kind).FontSize(12).FontWeight(600).TextColor(p.Accent)
				ui.Text(c, pod.Name).FontSize(30).FontWeight(800).LetterSpacing(-0.6).MaxLines(2).LineHeight(1.2)
				ui.RichText(c).FontSize(13).TextColor(p.Muted).Children(func() {
					if pod.DJ.UserID != 0 {
						d.Link(c.Key("dj"), pod.DJ.Nickname, p.Muted, p.Text, navigation.Route{Page: navigation.User, ID: pod.DJ.UserID})
						ui.Text(c, " · ")
					}
					ui.Text(c, fmt.Sprintf("%d 期 · %s 人订阅", pod.ProgramCount, appui.Count(int(pod.SubCount))))
				})
				if desc := strings.TrimSpace(pod.Description); desc != "" {
					ui.Text(c, strings.ReplaceAll(desc, "\n", " ")).FontSize(12.5).TextColor(p.Muted).MaxLines(2).LineHeight(1.45).Margin(2, 0, 0, 0)
				}
				ui.Row(c).Gap(10).Margin(10, 0, 0, 0).Children(func() {
					if appui.PillButton(c.Key("play"), appui.IconPlay, "播放", p.Accent, p.OnAccent).Clicked() {
						d.PlayProgram(0)
					}
					icon, label := appui.IconPlus, "订阅"
					if pod.Subed {
						icon, label = appui.IconCheck, "已订阅"
					}
					if appui.PillButton(c.Key("subscribe"), icon, label, p.Hover.Over(p.Bg), p.Text).Clicked() {
						d.ToggleSubscribe()
					}
				})
			})
		})
		Section(c, "节目")
		for i := range info.Programs {
			ProgramRow(c, d, info, i)
		}
		d.PageFoot(c, func() { d.ProgramsMore(id) })
		if info.PageLoaded && len(info.Programs) == 0 {
			d.Empty(c, "暂无节目")
		}
	})
}

// ProgramRow is an episode in its podcast's list, which a click plays.
func ProgramRow(c *ui.Context, d PodcastDeps, info PodcastInfo, i int) {
	p := d.Palette
	pr := &info.Programs[i]
	cur := d.Player.Current()
	playing := cur != nil && cur.ProgramID == pr.ID
	row := ui.ButtonBase(c).Height(60).PaddingX(12).Gap(14).Radius(14).AlignItems(ui.Center).FocusRing(false).
		Label(pr.Name).Disabled(pr.MainSong.ID == 0)
	switch {
	case playing && row.Pressed():
		row.Background(theme.Accent(p.Dark, 5))
	case playing && row.Hovered():
		row.Background(theme.Accent(p.Dark, 4))
	case playing:
		row.Background(theme.Accent(p.Dark, 3))
	case row.Hovered():
		row.Background(p.Hover)
	}
	row.Children(func() {
		cover := pr.CoverURL
		if cover == "" {
			cover = info.Podcast.PicURL
		}
		d.Cover(c, cover, 96, 40, 6)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
			title := ui.Text(c, pr.Name).FontSize(13.5).FontWeight(500).SingleLine()
			if playing {
				title.TextColor(p.Accent)
			}
			ui.Text(c, time.UnixMilli(pr.CreateTime).Format("2006-01-02")).FontSize(12).TextColor(p.Muted)
		})
		if playing {
			d.Bars(c, p.Accent)
		}
		// A liked episode's heart shows; the others' as the pointer is on them.
		ui.Box(c).Size(32, 32).Center().Children(func() {
			if d.Liked(pr.ID) || row.Hovered() {
				tr := pr.Track(info.Podcast)
				d.Like(c.Key("like"), &tr)
			}
		})
		ui.Text(c, appui.Clock(time.Duration(pr.Duration)*time.Millisecond)).FontSize(12).FontFeatures("tnum").TextColor(p.Muted)
	})
	if row.Clicked() {
		if playing {
			d.Player.TogglePlay()
		} else {
			d.PlayProgram(i)
		}
	}
}

// PodcastRow is a podcast in a list, opening its page.
func PodcastRow(c *ui.Context, d PodcastDeps, pod *qqmusic.Podcast) {
	p := d.Palette
	row := ui.ButtonBase(c).Padding(10, 12).Gap(14).Radius(14).AlignItems(ui.Center).FocusRing(false).Label(pod.Name)
	if row.Hovered() {
		row.Background(p.Hover)
	}
	row.Children(func() {
		d.Cover(c, pod.PicURL, 160, 64, 8)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(4).Children(func() {
			ui.Text(c, pod.Name).FontSize(14).FontWeight(600).SingleLine()
			meta := []string{fmt.Sprintf("%d 期", pod.ProgramCount)}
			if pod.DJ.Nickname != "" {
				meta = append([]string{pod.DJ.Nickname}, meta...)
			}
			ui.Text(c, strings.Join(meta, " · ")).FontSize(12.5).TextColor(p.Muted).SingleLine()
			if pod.Description != "" {
				ui.Text(c, pod.Description).FontSize(12.5).TextColor(p.Muted).SingleLine()
			}
		})
	})
	if row.Clicked() {
		d.Open(navigation.Route{Page: navigation.Podcast, ID: pod.ID})
	}
}

// PodcastCard is a podcast's cover over its name and its newest episode,
// or why to listen.
func PodcastCard(c *ui.Context, d Deps, pod *qqmusic.Podcast, play func()) {
	sub := pod.LastProgram
	if sub == "" {
		sub = pod.RcmdText
	}
	if sub == "" {
		sub = pod.DJ.Nickname
	}
	if CoverCard(c, d, pod.Name, pod.PicURL, pod.PlayCount, sub, play) {
		d.Open(navigation.Route{Page: navigation.Podcast, ID: pod.ID})
	}
}

// PodcastsInfo is what the podcasts tab lists, as far as it has loaded.
type PodcastsInfo struct {
	Loaded     bool
	Loading    bool
	Err        string
	Subs       []qqmusic.Podcast
	Picks      []qqmusic.Podcast
	Hot        []qqmusic.Podcast
	HotMore    bool
	HotLoading bool
	HotErr     string
}

// PodcastsDeps are what the podcasts tab needs from the app it runs in.
type PodcastsDeps struct {
	Palette theme.Palette
	Sidebar float32

	Info   func() PodcastsInfo
	Scroll *ui.ScrollState

	Insets    func(*ui.Context) (top, bottom float32)
	StartMark func(*ui.Context)
	EndMark   func(*ui.Context)

	Centered func(*ui.Context, string, func())
	LoadTab  func()
	LoadMore func()
	Skeleton func(*ui.Context)
	Card     func(*ui.Context, *qqmusic.Podcast)
	PageFoot func(*ui.Context, func())
}

// PodcastsPage lists the listener's subscriptions, NetEase's picks and
// the most popular podcasts.
func PodcastsPage(c *ui.Context, d PodcastsDeps) {
	info := d.Info()
	if !info.Loaded && !info.Loading {
		d.Centered(c, info.Err, d.LoadTab)
		return
	}
	top, bottom := d.Insets(c)
	ui.Scroll(c.Key("podcasts")).TrackScroll(d.Scroll).Grow(1).Padding(top+PageGap, Pad, bottom+48, Pad).ScrollbarInsets(top, 0, bottom, 0).Children(func() {
		d.StartMark(c)
		defer d.EndMark(c)
		Heading(c, d.Palette, "播客", "")
		cols := Columns(c, d.Sidebar, 150)
		if !info.Loaded {
			d.Skeleton(c)
			return
		}
		// The first list starts right under the heading.
		first := true
		list := func(title string, pods []qqmusic.Podcast) {
			if len(pods) == 0 {
				return
			}
			if s := Section(c, title); first {
				s.Margin(0, 0, 14, 0)
			}
			first = false
			ui.Grid(c.Key(title)).Columns(cols).GapX(20).GapY(28).Children(func() {
				for i := range pods {
					d.Card(c.Key(pods[i].ID), &pods[i])
				}
			})
		}
		list("我的订阅", info.Subs)
		list("精选播客", info.Picks)
		list("热门播客", info.Hot)
		d.PageFoot(c, d.LoadMore)
	})
}
