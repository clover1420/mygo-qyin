package app

import "github.com/egoist/mygo/ui"

// queueButton opens the queue above the player without leaving the current page.
func (a *app) queueButton(c *ui.Context) {
	p := a.pal
	surface := playerSurface
	if a.queueOpen {
		surface = playerSelectedSurface
	}
	b := iconButton(c.Key("queue"), iconQueue, 16, p.Muted, "待播清单", surface).
		Disabled(a.pl.Current() == nil)
	if b.Clicked() {
		a.queueOpen = !a.queueOpen
		if a.queueOpen {
			a.queuePopupAt = -1
		}
	}
	ink := a.inkOn(false)
	ui.PopoverBase(c.Key("queue-popover"), b, &a.queueOpen, func(panel ui.Element) {
		w, h := c.Size()
		height := min(float32(len(a.pl.Queue())*60+64), 460, h-topBar(c)-playerBarHeight-24)
		panel.AttachTo(b, ui.AnchorTopEnd, ui.AnchorBottomEnd).
			Margin(0, 0, (playerBarHeight-32)/2+8, 0).Size(min(320, w-32), height).
			Padding(8).TextColor(ink.text).Label("待播清单")
		a.pane(panel, 14, false)
		// Overlay content sits outside the page's Space shortcut scope.
		panel.OnShortcut(0, ui.KeySpace, a.pl.TogglePlay)
		ui.Row(c).Height(40).PaddingX(8).Gap(8).AlignItems(ui.Center).Children(func() {
			ui.Text(c, "待播清单").FontSize(15).FontWeight(700).TextColor(ink.text)
			ui.Textf(c, "%d", len(a.pl.Queue())).FontSize(12).FontFeatures("tnum").TextColor(ink.muted)
			ui.Spacer(c)
			if iconButton(c, iconX, 15, ink.muted, "关闭待播清单").Clicked() {
				a.queueOpen = false
			}
		})
		a.queueTracks(c, &a.queuePopup, &a.queuePopupAt, p).Padding(4, 0, 8, 0)
	})
}
