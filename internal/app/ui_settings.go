package app

import (
	"runtime"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// openSettings uses a separate window on macOS and the main navigation elsewhere.
func (a *app) openSettings() {
	if runtime.GOOS != "darwin" {
		a.showMain()
		if a.route.page == pageSettings {
			a.measureCache()
		}
		a.open(route{page: pageSettings})
		a.redraw()
		return
	}
	if a.settingsWin != nil {
		a.settingsWin.Show()
		a.settingsWin.Focus()
		a.measureCache()
		return
	}
	a.measureCache()
	a.settingsWin = mygo.NewWindow(mygo.WindowOptions{
		Title:           "设置",
		Width:           560,
		Height:          720,
		MinWidth:        480,
		MinHeight:       420,
		StateKey:        "settings",
		DisableMaximize: true,
		DisableMinimize: true,
		AutoHideMenuBar: true,
		BackgroundColor: "light-dark(" + neutralLight[0] + ", " + neutralDark[0] + ")",
		Content:         ui.View(a.settingsView),
	})
	a.settingsWin.OnClosed(func() {
		a.settingsWin = nil
		a.syncDock()
	})
	a.syncDock()
}

// changed applies the settings everywhere and keeps them.
func (a *app) changed() {
	a.settings.save()
	a.settings.applyAppearance()
	if a.notch != nil {
		a.notch.setEnabled(a.settings.Notch)
	}
	a.redraw()
}

// settingsCategory is one of the destinations down the settings' side: the
// rows of the others are not built while it shows.
type settingsCategory int

const (
	settingsGeneral settingsCategory = iota
	settingsAppearance
	settingsPlayback
	settingsAbout
)

var settingsCategories = [...]struct {
	label string
	icon  *ui.SVG
}{
	{"常规", iconSettings},
	{"外观", iconSun},
	{"播放与歌词", iconHeadset},
	{"关于", iconInfo},
}

func (a *app) settingsView(c *ui.Context) {
	t := *c.Theme()
	p := paletteFor(&t)
	t.Accent, t.AccentText = p.accent, p.onAccent
	t.Font = a.fontFamily()
	c.SetTheme(&t)
	c.Root().Background(p.sidebar)
	c.OnShortcut(ui.Cmd, ui.KeyW, func() { a.settingsWin.Close() })
	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		a.settingsSidebar(c)
		ui.Box(c).Width(0.5).Shrink(0).Background(p.hairline)
		a.settingsBody(c)
	})
}

// settingsPage shares the settings controls with the macOS window.
func (a *app) settingsPage(c *ui.Context) {
	top, bottom := a.insets(c)
	ui.Box(c.Key("settings-page")).Fill().Background(a.pal.bg).Padding(top+pageGap, 0, bottom, 0).Children(func() {
		a.startMark(c)
		defer a.endMark(c)
		ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
			a.settingsSidebar(c)
			ui.Box(c).Width(0.5).Shrink(0).Background(a.pal.hairline)
			a.settingsBody(c)
		})
	})
}

// settingsSidebar is the list of destinations, which stays while one of them
// scrolls.
func (a *app) settingsSidebar(c *ui.Context) {
	p := paletteFor(c.Theme())
	ui.Column(c.Key("settings-nav")).Width(184).Shrink(0).Padding(20, 12, 20, 4).Gap(4).Label("设置分类").
		Background(p.sidebar).Children(func() {
		for i, category := range settingsCategories {
			selected := a.settingsTab == int(i)
			button := ui.ButtonBase(c.Key(category.label)).Height(38).PaddingX(10).Gap(10).Radius(10).
				AlignItems(ui.Center).Justify(ui.Start).Label("设置分类：" + category.label).Selected(selected)
			bg, color, weight := ui.Transparent, p.muted, 500
			if selected {
				bg, color, weight = p.press, p.text, 600
			} else if button.Hovered() {
				bg, color = p.hover, p.text
			}
			button.Background(button.AnimateColor("bg", bg, ui.Animation{Duration: 120 * time.Millisecond})).TextColor(color)
			button.Children(func() {
				ui.Icon(c, category.icon).Size(17, 17)
				ui.Text(c, category.label).FontSize(13.5).FontWeight(weight)
			})
			if button.Clicked() {
				a.settingsTab = int(i)
			}
		}
	})
}

// settingsBody scrolls the rows of the destination chosen, and keeps its
// place between visits.
func (a *app) settingsBody(c *ui.Context) {
	ui.Scroll(c.Key(settingsCategories[a.settingsTab].label)).Bind(&a.settingsPane).
		TrackScroll(&a.settingsScroll[a.settingsTab]).Grow(1).Shrink(1).MinWidth(0).
		Background(paletteFor(c.Theme()).bg).Padding(20, 28, 32, 28).Gap(8).Children(func() {
		a.settingsContent(c)
	})
}

func (a *app) settingsContent(c *ui.Context) {
	p := paletteFor(c.Theme())
	s := &a.settings

	// A group of the destination showing, and none of the others: their rows
	// are not built, so that no control of another page takes a key.
	group := func(cat settingsCategory, title string, rows func()) {
		if a.settingsTab != int(cat) {
			return
		}
		ui.Text(c, title).FontSize(12).FontWeight(600).TextColor(p.muted).Margin(14, 4, 4, 4)
		ui.Column(c).Radius(20).Background(p.bg).Border(1, p.hairline).Dividers(1, p.hairline).Children(rows)
	}
	row := func(label, desc string, control func()) {
		ui.Row(c).Padding(12, 16).Gap(16).AlignItems(ui.Center).MinHeight(48).Children(func() {
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(3).Children(func() {
				ui.Text(c, label).FontSize(13.5).FontWeight(500)
				if desc != "" {
					ui.Text(c, desc).FontSize(12).TextColor(p.muted).LineHeight(1.4)
				}
			})
			control()
		})
	}

	group(settingsGeneral, "通用", func() {
		row("外观", "", func() {
			modes := []string{"system", "light", "dark"}
			i := 0
			for k, m := range modes {
				if m == s.Appearance {
					i = k
				}
			}
			if segmented(c.Key("appearance"), &i, "跟随系统", "浅色", "深色") {
				s.Appearance = modes[i]
				a.changed()
			}
		})
		if runtime.GOOS == "windows" {
			row("侧边栏材质", "", func() {
				i := 0
				if s.sidebarVibrancy() == mygo.VibrancyMica {
					i = 1
				}
				if segmented(c.Key("sidebar-material"), &i, "亚克力", "Mica") {
					s.SidebarMaterial = []string{"acrylic", "mica"}[i]
					if a.win != nil {
						a.win.SetVibrancy(s.sidebarVibrancy())
					}
					s.save()
					a.redraw()
				}
			})
		}

		if a.canHideOnClose() {
			row("关闭主窗口时", "", func() {
				i := 0
				if s.QuitOnClose {
					i = 1
				}
				keepRunning := "最小化到托盘"
				switch runtime.GOOS {
				case "darwin":
					keepRunning = "保持后台运行"
				case "linux":
					keepRunning = "隐藏到托盘"
				}
				if segmented(c.Key("close-action"), &i, keepRunning, "退出应用") {
					s.QuitOnClose = i == 1
					s.save()
				}
			})
			// Only on macOS: elsewhere the icon is always in the notification
			// area.
			if runtime.GOOS == "darwin" {
				row("应用图标", "显示在菜单栏时，关闭窗口后可从菜单栏图标打开", func() {
					labels := make([]string, len(appIcons))
					chosen := appIcons[0].label
					for i, ic := range appIcons {
						labels[i] = ic.label
						if ic.id == s.AppIcon {
							chosen = ic.label
						}
					}
					if ui.Select(c.Key("app-icon"), &chosen, labels).Label("应用图标").Width(150).Changed() {
						for _, ic := range appIcons {
							if ic.label == chosen {
								s.AppIcon = ic.id
							}
						}
						a.changed()
						a.changeAppIcon()
					}
				})
			}
		}

		// Only on macOS, the one with a notch player.
		if a.notch != nil {
			row("显示刘海播放器", "在屏幕顶部的刘海处显示迷你播放器，鼠标悬停即可展开", func() {
				if ui.Switch(c.Key("notch"), &s.Notch).Label("显示刘海播放器").Changed() {
					a.changed()
				}
			})
			row("播放时显示实时活动", "在刘海两侧显示封面与音频波形，可能遮挡刘海旁的菜单栏图标", func() {
				if ui.Switch(c.Key("activity"), &s.NotchActivity).Label("播放时显示实时活动").Disabled(!s.Notch).Changed() {
					a.changed()
				}
			})
			row("在无刘海的屏幕上常驻显示", "关闭时，将鼠标移至屏幕顶部中央即可唤出，不会遮挡菜单栏", func() {
				if ui.Switch(c.Key("virtual"), &s.NotchVirtual).Label("在无刘海的屏幕上常驻显示").Disabled(!s.Notch).Changed() {
					a.changed()
				}
			})
		}

		row("音质", "按 QQ 音乐账号权限播放 MP3 或 FLAC，下一首生效", func() {
			labels := make([]string, len(qualityLevels))
			chosen := qualityLevels[1].label
			for i, q := range qualityLevels {
				labels[i] = q.label
				if q.level == s.Quality {
					chosen = q.label
				}
			}
			if ui.Select(c.Key("quality"), &chosen, labels).Label("音质").Width(150).Changed() {
				for _, q := range qualityLevels {
					if q.label == chosen {
						s.Quality = q.level
					}
				}
				a.changed()
			}
		})
		row("音量平衡", "", func() {
			labels := make([]string, len(volumeBalances))
			chosen := volumeBalances[2].label
			for i, balance := range volumeBalances {
				labels[i] = balance.label
				if balance.id == s.volumeBalance() {
					chosen = balance.label
				}
			}
			if ui.Select(c.Key("volume-balance"), &chosen, labels).Label("选择音量平衡").Width(150).Changed() {
				for _, balance := range volumeBalances {
					if balance.label == chosen {
						s.setVolumeBalance(balance.id)
					}
				}
				a.pl.RefreshGain()
				s.save()
				a.redraw()
			}
		})
		row("间奏时显示热评", "前奏、间奏与尾奏时，歌词处换成这首歌的热门评论", func() {
			if ui.Switch(c.Key("hot-comments"), &s.HotComments).Label("切换热评").Changed() {
				s.save()
				a.redraw()
			}
		})

		row("歌曲缓存", a.cacheUsage(), func() {
			labels := make([]string, len(cacheLimits))
			chosen := cacheLimits[2].label
			for i, l := range cacheLimits {
				labels[i] = l.label
				if l.mb == s.CacheMB {
					chosen = l.label
				}
			}
			if ui.Select(c.Key("cache"), &chosen, labels).Label("歌曲缓存").Width(150).Disabled(a.cacheMoving).Changed() {
				for _, l := range cacheLimits {
					if l.label == chosen {
						s.CacheMB = l.mb
					}
				}
				a.changed()
				a.applyCacheLimit()
			}
		})
		a.cacheDirectoryRow(c)
		row("清除缓存", "", func() {
			if ui.Button(c.Key("clear-cache"), "清除").Disabled(a.cacheMoving || a.cacheUsed <= 0).Clicked() {
				a.clearCache()
			}
		})
	})

	group(settingsAppearance, "外观与玻璃", func() {
		row("字体", "界面文字所用的字体，随系统所装字体而定", func() {
			fonts := a.fontChoices()
			labels := make([]string, len(fonts))
			chosen := fonts[0].label
			for i, f := range fonts {
				labels[i] = f.label
				if f.id == s.Font {
					chosen = f.label
				}
			}
			if ui.Select(c.Key("font"), &chosen, labels).Label("字体").Width(150).Changed() {
				for _, f := range fonts {
					if f.label == chosen {
						s.Font = f.id
					}
				}
				a.loadFont()
				a.changed()
			}
		})
		row("液态玻璃播放栏", "播放栏浮起为一条玻璃胶囊，页面从它底下滚过", func() {
			if ui.Switch(c.Key("liquid-glass"), &s.LiquidGlass).Label("液态玻璃播放栏").Changed() {
				a.changed()
			}
		})
		glassRow := func(label, desc string, v *float64) {
			row(label, desc, func() {
				a.valueSlider(c.Key(label), label, label, v, s.LiquidGlass)
			})
		}
		glassRow("玻璃透明度", "越高越能看清背后的内容，文字会更难读", &s.GlassClear)
		glassRow("折射强度", "胶囊边缘弯折背后内容的程度", &s.GlassBend)
		row("界面玻璃", "菜单、弹层与对话框使用磨砂玻璃", func() {
			if ui.Switch(c.Key("glass-panels"), &s.GlassPanels).Label("界面玻璃").Changed() {
				a.changed()
			}
		})
		row("歌词光晕", "正在唱的那一行，在封面的颜色里发光", func() {
			if ui.Switch(c.Key("np-light"), &s.NpLight).Label("歌词光晕").Changed() {
				a.changed()
			}
		})
		row("光晕强度", "越高越亮，字也越容易糊在一起", func() {
			a.valueSlider(c.Key("glow-level"), "光晕强度", "光晕强度", &s.GlowLevel, s.NpLight)
		})
	})

	group(settingsPlayback, "歌词与播放", func() {
		if supportsTaskbarLyrics() {
			row("歌词显示位置", "", func() {
				i := 0
				if s.lyricsMode() == lyricsTaskbar {
					i = 1
				}
				if segmented(c.Key("lyrics-mode"), &i, "桌面", "任务栏") {
					mode := lyricsDesktop
					if i == 1 {
						mode = lyricsTaskbar
					}
					a.setLyricsMode(mode)
				}
			})
		}
		if supportsDesktopLyrics() && s.lyricsMode() == lyricsDesktop {
			row("桌面歌词行数", "双行仅针对中文歌", func() {
				i := 0
				if s.DesktopLyricsDoubleLine {
					i = 1
				}
				if segmented(c.Key("desktop-lyrics-lines"), &i, "单行", "双行") {
					s.DesktopLyricsDoubleLine = i == 1
					s.save()
					a.redraw()
				}
			})
		}
		row("进度条特效", "", func() {
			chosen := progressEffects[0].label
			labels := make([]string, len(progressEffects))
			for k, effect := range progressEffects {
				labels[k] = effect.label
				if effect.id == s.progressEffect() {
					chosen = effect.label
				}
			}
			if ui.Select(c.Key("progress-effect"), &chosen, labels).Label("选择进度条特效").Width(150).Changed() {
				for _, effect := range progressEffects {
					if effect.label == chosen {
						s.setProgressEffect(effect.id)
					}
				}
				s.save()
				a.redraw()
			}
		})
		row("烟雾背景", "播放页的背景化作封面色调的烟雾，缓缓升起，随低音起伏；仍在打磨，会多用一些电量", func() {
			if ui.Switch(c.Key("fog"), &s.Fog).Label("烟雾背景").Changed() {
				a.changed()
			}
		})
	})

	group(settingsAbout, "关于与更新", func() {
		row("当前版本 "+mygo.App.Version(), "", func() {
			if a.updates {
				if ui.Button(c.Key("check-updates"), "检查更新").Clicked() {
					checkForUpdates()
				}
			}
		})
		if a.updates {
			desc := ""
			if t := lastUpdateCheck(); !t.IsZero() {
				desc = "上次检查：" + commentTime(t, time.Now())
			}
			row("自动检查更新", desc, func() {
				on := updateChecks()
				if ui.Switch(c.Key("update-checks"), &on).Label("自动检查更新").Changed() {
					setUpdateChecks(on)
				}
			})
			row("自动下载并安装", "", func() {
				on := updateDownloads()
				if ui.Switch(c.Key("update-downloads"), &on).Label("自动下载并安装").Disabled(!updateChecks()).Changed() {
					setUpdateDownloads(on)
				}
			})
		}
	})
}

// valueSlider is a setting that goes from 0 to 1: a thin bar in the accent
// over the page's press shade, with the value beside it, dimmed while the
// switch it belongs to is off.
func (a *app) valueSlider(c *ui.Context, key, label string, v *float64, on bool) {
	// The settings window draws its own palette, which the app's may not be.
	p := paletteFor(c.Theme())
	ui.Row(c).Gap(10).AlignItems(ui.Center).Children(func() {
		s := ui.SliderBase(c.Key(key), v, 0, 1).Width(150).Height(22).Step(0.01).FocusRing(false).Label(label).
			Disabled(!on).Cursor(ui.CursorPointer)
		if s.Changed() {
			a.settingsDirty = true
			a.redraw()
		}
		if a.settingsDirty && !s.Pressed() {
			a.settingsDirty = false
			a.settings.save()
		}
		held := s.Pressed()
		thick := float32(4)
		if s.Hovered() || held {
			thick = 5
		}
		thick = s.AnimateTo("thick", thick, ui.Snappy)
		at := float32(*v)
		hot := held || s.Hovered() || s.FocusVisible()
		track, fill, knob := p.press, p.accent, paper
		if !on {
			// Off, the slider says so by its colors rather than by a layer.
			track, fill, knob = p.hover, p.faint, p.hover
		}
		s.Draw(func(pel *ui.Painter, r ui.Rect) {
			y := r.Y + (r.H-thick)/2
			pel.Fill(ui.Rect{X: r.X, Y: y, W: r.W, H: thick}, track, thick/2)
			pel.Fill(ui.Rect{X: r.X, Y: y, W: max(thick, r.W*at), H: thick}, fill, thick/2)
			if hot {
				pel.Fill(ui.Rect{X: r.X + r.W*at - 6, Y: r.Y + (r.H-12)/2, W: 12, H: 12}, knob, 6)
			}
		})
		ui.Textf(c, "%d%%", int(100*(*v))).FontSize(12).FontFeatures("tnum").TextColor(p.muted).Width(38)
	})
}
