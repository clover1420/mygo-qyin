package pages

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/base/appui"
	"lilt/internal/ui/base/theme"
)

// NameState is the playlist-name dialog's state, which the app keeps.
type NameState struct {
	Open   *bool
	Name   *string
	Rename int64
	Busy   bool
}

// DeleteState is the delete/unfavorite dialog's state, which the app keeps.
type DeleteState struct {
	Open  *bool
	Name  string
	Saved bool
}

// PlaylistEditDeps are what the playlist dialogs need from the app.
type PlaylistEditDeps struct {
	Palette  theme.Palette
	PaneLook func(ui.Element, float32, bool) ui.Element

	Name         func() NameState
	Delete       func() DeleteState
	SubmitName   func()
	DeleteChosen func()
}

// PlaylistNameDialog asks for the name of a playlist to make, or a new
// name for one.
func PlaylistNameDialog(c *ui.Context, d PlaylistEditDeps) {
	p := d.Palette
	s := d.Name()
	ui.DialogBase(c.Key("playlist-name"), s.Open, func(backdrop, panel ui.Element) {
		backdrop.Background(theme.Ink.Alpha(0.35))
		panel.Width(360).Padding(24).Gap(16)
		d.PaneLook(panel, 20, false)
		panel.Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.94}, Exit: &ui.Motion{Scale: 0.96}})
		title, action := "新建歌单", "创建"
		if s.Rename != 0 {
			title, action = "重命名歌单", "保存"
		}
		if s.Busy {
			action = "正在创建…"
		}
		ui.Text(c, title).FontSize(17).FontWeight(700)
		in := ui.TextInput(c.Key("name"), s.Name).Placeholder("歌单名称").Label("歌单名称").FontSize(14).AutoFocus()
		if in.Submitted() {
			d.SubmitName()
		}
		ui.Row(c).Gap(10).Justify(ui.End).Children(func() {
			if appui.PillButton(c.Key("cancel"), nil, "取消", p.Hover.Over(p.Bg), p.Text).Clicked() {
				*s.Open = false
			}
			ok := appui.PillButton(c.Key("ok"), nil, action, p.Accent, p.OnAccent)
			if strings.TrimSpace(*s.Name) == "" {
				ok.Disabled(true).Opacity(0.5)
			}
			if ok.Clicked() {
				d.SubmitName()
			}
		})
	})
}

// DeletePlaylistDialog confirms deleting an owned playlist or unfavoriting
// a saved one.
func DeletePlaylistDialog(c *ui.Context, d PlaylistEditDeps) {
	p := d.Palette
	s := d.Delete()
	ui.DialogBase(c.Key("playlist-delete"), s.Open, func(backdrop, panel ui.Element) {
		backdrop.Background(theme.Ink.Alpha(0.35))
		panel.Width(360).Padding(24).Gap(8)
		d.PaneLook(panel, 20, false)
		panel.Transition(ui.ElementTransition{Animation: &ui.Snappy, Enter: &ui.Motion{Scale: 0.94}, Exit: &ui.Motion{Scale: 0.96}})
		title, action := "删除歌单", "删除"
		message := "确定删除歌单“" + s.Name + "”吗？删除后无法恢复。"
		if s.Saved {
			title, action = "取消收藏歌单", "取消收藏"
			message = "确定取消收藏歌单“" + s.Name + "”吗？"
		}
		ui.Text(c, title).FontSize(17).FontWeight(700)
		ui.Text(c, message).FontSize(13).TextColor(p.Muted).LineHeight(1.45)
		ui.Row(c).Gap(10).Justify(ui.End).Margin(8, 0, 0, 0).Children(func() {
			if appui.PillButton(c.Key("cancel"), nil, "取消", p.Hover.Over(p.Bg), p.Text).Clicked() {
				*s.Open = false
			}
			if appui.PillButton(c.Key("ok"), nil, action, p.Accent, p.OnAccent).Clicked() {
				d.DeleteChosen()
			}
		})
	})
}
