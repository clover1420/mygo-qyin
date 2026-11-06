package app

import (
	"github.com/egoist/mygo/ui"
)

// artistEvents lists what an artist posted, a page at a time.
func (a *app) artistEvents(c *ui.Context, d *artistData) {
	if d.profile.AccountID == 0 {
		a.empty(c, "这位歌手还没有自己的账号，没有动态")
		return
	}
	ui.Column(c).Children(func() {
		for i := range d.events {
			a.event(c.Key(d.events[i].ID), &d.events[i])
		}
		id := d.profile.ID
		a.pageFoot(c, &d.eventsPage, func() { a.loadEvents(id) })
		if d.eventsPage.loaded && len(d.events) == 0 {
			a.empty(c, "暂无动态")
		}
	})
}
