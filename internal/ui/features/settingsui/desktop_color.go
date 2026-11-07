package settingsui

import (
	"fmt"

	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/base/theme"
	"lilt/internal/ui/features/lyricsui"
)

// desktopLyricsColorControl picks the color desktop lyrics are drawn in:
// presets first, then the native picker for any color.
func desktopLyricsColorControl(c *ui.Context, d Deps) {
	p := theme.For(c.Theme())
	s := d.Settings
	color := lyricsui.DesktopLyricsColor(s.DesktopLyricsColor)
	choose := func(value ui.Color) {
		lyricsui.SetDesktopLyricsColor(s, value)
		s.Save()
		d.Redraw()
	}
	ui.Row(c.Key("desktop-lyrics-colors")).AlignItems(ui.Center).Gap(4).Shrink(0).Children(func() {
		custom := true
		for _, preset := range lyricsui.DesktopLyricsColors {
			value := ui.Hex(preset.Hex)
			button := ui.ButtonBase(c.Key(preset.Name)).Size(28, 28).Center().Radius(14).
				Label("桌面歌词颜色："+preset.Name).Tooltip(preset.Name).Checked(color == value)
			if color == value {
				custom = false
				button.Border(1, theme.Neutral(p.Dark, 8))
			} else if button.Hovered() || button.FocusVisible() {
				button.Background(theme.Neutral(p.Dark, 3))
			}
			button.Children(func() {
				ui.Box(c).Size(18, 18).Radius(9).Background(value).Border(1, p.Hairline)
			}).OnClick(func() {
				choose(value)
			})
		}
		ui.Box(c).Width(1).Height(18).Background(p.Hairline).MarginX(5)
		button := ui.ButtonBase(c.Key("custom")).Size(28, 28).Center().Radius(14).
			Label("自定义桌面歌词颜色").Tooltip("自定义颜色").Role(ui.RoleColorWell).Checked(custom).
			Value(fmt.Sprintf("#%02x%02x%02x%02x", color.R, color.G, color.B, color.A))
		open := ui.Local(button, "open", func() bool { return false })
		button.Expanded(*open).OnClick(func() { *open = !*open })
		if custom {
			button.Border(1, theme.Neutral(p.Dark, 8))
		} else if button.Hovered() || button.FocusVisible() {
			button.Background(theme.Neutral(p.Dark, 3))
		}
		button.Children(func() {
			ui.Box(c).Size(18, 18).Radius(9).ConicGradient(lyricsui.DesktopLyricsColorWheel).Border(1, p.Hairline)
		})
		ui.Popover(c.Key("custom-picker"), button, open, func() {
			ui.ColorPicker(c, &color).OnChange(func() {
				choose(color)
			})
		}).Padding(12).Background(p.Bg).Border(1, p.Hairline).Shadow(0, 6, 20, 0, theme.Shade(4))
	})
}
