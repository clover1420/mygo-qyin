package pages

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/core/navigation"
	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
)

// The tabs of the purchased page.
const (
	PurchasedAlbums = iota
	PurchasedSongs
)

// PurchasedInfo is what the app has loaded of the listener's purchases.
type PurchasedInfo struct {
	Songs   []qqmusic.Track
	Albums  []qqmusic.AlbumProfile
	Loaded  bool
	Loading bool
	More    bool
	Err     string
}

// PurchasedDeps are what the purchased page needs from the app it runs in.
type PurchasedDeps struct {
	Palette theme.Palette
	Route   *navigation.Route

	SignedIn  bool
	OpenLogin func()
	Insets    func(*ui.Context) (top, bottom float32)

	Info          func(tab int) PurchasedInfo
	State         func(tab int) *ListState
	ListPage      func(*ui.Context, any, *ui.ListState, func())
	Skeleton      func(*ui.Context, func(*ui.Context))
	CardsSkeleton func(*ui.Context)
	PageFoot      func(*ui.Context, func())
	Empty         func(*ui.Context, string)
	TrackRows     func(*ui.Context, any, *ListState, []qqmusic.Track, bool, func(*ui.Context), func(*ui.Context))
	AlbumGrid     func(*ui.Context, []qqmusic.AlbumProfile)
	Play          func([]qqmusic.Track)
	More          func(tab int)
	Reload        func(tab int)
	Show          func(navigation.Route)
}

// PurchasedPage shows what the listener bought: their albums or songs.
func PurchasedPage(c *ui.Context, d PurchasedDeps) {
	if !d.SignedIn {
		top, bottom := d.Insets(c)
		ui.Column(c).Grow(1).Padding(top, 0, bottom, 0).Center().Gap(12).Children(func() {
			ui.Text(c, "登录后查看已购音乐").FontSize(13).TextColor(d.Palette.Muted)
			if ui.Button(c, "登录 QQ 音乐").Clicked() {
				d.OpenLogin()
			}
		})
		return
	}
	tab := d.Route.Tab
	info := d.Info(tab)
	content := navigation.Route{Page: navigation.Purchased}
	header := func(c *ui.Context) { purchasedHeader(c, d, tab) }
	foot := func(c *ui.Context) {
		d.PageFoot(c, func() { d.More(tab) })
		if info.Loaded && len(info.Songs)+len(info.Albums) == 0 {
			label := "暂无已购单曲"
			if tab == PurchasedAlbums {
				label = "暂无已购专辑"
			}
			d.Empty(c, label)
		}
	}
	state := d.State(tab)
	if tab == PurchasedSongs {
		if !info.Loaded && info.Loading {
			d.ListPage(c, content, &state.ListState, func() {
				d.Skeleton(c, header)
			})
			return
		}
		d.TrackRows(c, content, state, info.Songs, true, header, foot)
		return
	}
	d.ListPage(c, content, &state.ListState, func() {
		header(c)
		d.AlbumGrid(c, info.Albums)
		if !info.Loaded && info.Loading {
			d.CardsSkeleton(c)
			return
		}
		foot(c)
	})
}

func purchasedHeader(c *ui.Context, d PurchasedDeps, tab int) {
	Heading(c, d.Palette, "已购音乐", "")
	ui.Row(c).AlignItems(ui.Center).Gap(12).Margin(0, 0, 18, 0).Children(func() {
		if appui.Segmented(c.Key("purchased-kind"), &tab, "专辑", "单曲") {
			next := *d.Route
			next.Tab = tab
			d.Show(next)
		}
		ui.Box(c).Grow(1)
		info := d.Info(d.Route.Tab)
		if len(info.Songs) > 0 {
			appui.PillButton(c.Key("play-purchased"), appui.IconPlay, "播放", d.Palette.Accent, d.Palette.OnAccent).
				OnClick(func() { d.Play(info.Songs) })
		}
		appui.IconButton(c.Key("refresh-purchased"), appui.IconReload, 17, d.Palette.Muted, "刷新已购音乐").
			Disabled(info.Loading).OnClick(func() { d.Reload(d.Route.Tab) })
	})
}
