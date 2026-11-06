package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
)

// artistLinks shows a track's artists as its credits read, each a link to
// their page in color, brightening to hover under the pointer, and then,
// with album, a link to its album. An artist NetEase has no page for,
// credited by name alone, is plain text. An episode's host and podcast,
// in its artists' and album's place, link to the podcast's page.
func (a *app) artistLinks(c *ui.Context, tr *qqmusic.Track, color, hover ui.Color, album ...bool) ui.Element {
	podcast := route{Page: pagePodcast, ID: tr.PodcastID}
	return ui.RichText(c).TextColor(color).Children(func() {
		for i, ar := range tr.Artists {
			if i > 0 {
				ui.Text(c, ", ")
			}
			switch {
			case tr.PodcastID != 0:
				a.link(c.Key(i), ar.Name, color, hover, podcast)
			case ar.ID != 0:
				a.link(c.Key(i), ar.Name, color, hover, route{Page: pageArtist, ID: ar.ID})
			default:
				ui.Text(c, ar.Name)
			}
		}
		if len(album) > 0 && album[0] && tr.Album.Name != "" {
			ui.Text(c, " — ")
			switch {
			case tr.PodcastID != 0:
				a.link(c.Key("album"), tr.Album.Name, color, hover, podcast)
			case tr.Album.ID != 0:
				a.link(c.Key("album"), tr.Album.Name, color, hover, route{Page: pageAlbum, ID: tr.Album.ID})
			default:
				ui.Text(c, tr.Album.Name)
			}
		}
	})
}

// link is text in a RichText that opens a page.
func (a *app) link(c *ui.Context, text string, color, hover ui.Color, r route) {
	link := ui.Link(c, text, "").TextColor(color)
	if link.Hovered() {
		link.TextColor(hover)
	}
	// Once built: a song's row is in a list, which a new page resets.
	link.OnClick(func() { a.openPage(r) })
}

// openPage opens a page in the main window, from wherever a link to it
// shows: over Now Playing, which closes, or the notch.
func (a *app) openPage(r route) {
	a.open(r)
	a.showMain()
	a.redraw()
}

// openArtist opens an artist's page in the main window.
func (a *app) openArtist(id int64) { a.openPage(route{Page: pageArtist, ID: id}) }

// These load an artist, their songs, albums and events from NetEase, and
// follow them; tests replace them.
var (
	artistDetail      = qqmusic.ArtistDetail
	artistLatestSongs = qqmusic.ArtistLatestSongs
	albumsBy          = qqmusic.ArtistAlbums
	accountEvents     = qqmusic.Events
	followArtist      = qqmusic.FollowArtist
)

// artistData is what the app has of an artist: who they are, with their
// most popular songs, and their newest songs, albums and events once
// asked for.
type artistData struct {
	profile       *qqmusic.ArtistProfile
	loading       bool
	err           string
	latest        []qqmusic.Track
	latestLoading bool
	latestErr     string
	albums        []qqmusic.AlbumProfile
	albumsPage    paged
	events        []qqmusic.Event
	eventsPage    paged
	eventsBefore  int64 // the time of the last event loaded, the next page's start
	following     bool  // a follow or unfollow is on its way
}

// latestCount is how many of an artist's newest songs their page lists;
// albumsPage and eventsPage how many albums and events load at a time.
const (
	latestCount = 100
	albumsPage  = 30
	eventsPage  = 20
)

func (a *app) artist(id int64) *artistData {
	d := a.artists[id]
	if d == nil {
		d = &artistData{}
		a.artists[id] = d
	}
	return d
}

// loadArtist loads an artist's page, and what it lists.
func (a *app) loadArtist(id int64) {
	d := a.artist(id)
	switch a.artistTab {
	case artistLatest:
		a.loadLatest(id)
	case artistAlbums:
		if !d.albumsPage.loaded {
			a.loadAlbums(id)
		}
	}
	if d.profile != nil || d.loading {
		if a.artistTab == artistEvents && !d.eventsPage.loaded {
			a.loadEvents(id)
		}
		return
	}
	d.loading, d.err = true, ""
	go func() {
		ar, err := artistDetail(id)
		a.post(func() {
			d.loading = false
			if err != nil {
				d.err = "歌手加载失败"
				return
			}
			d.profile = ar
			// Their events are their account's, which their profile names.
			if a.artistTab == artistEvents && a.route == (route{Page: pageArtist, ID: id}) {
				a.loadEvents(id)
			}
		})
	}()
}

func (a *app) loadLatest(id int64) {
	d := a.artist(id)
	if d.latest != nil || d.latestLoading {
		return
	}
	d.latestLoading, d.latestErr = true, ""
	go func() {
		songs, err := artistLatestSongs(id, latestCount)
		a.post(func() {
			d.latestLoading = false
			if err != nil {
				d.latestErr = "歌曲加载失败"
				return
			}
			d.latest = songs
		})
	}()
}

// loadAlbums loads the next page of an artist's albums.
func (a *app) loadAlbums(id int64) {
	d := a.artist(id)
	if d.albumsPage.loading {
		return
	}
	d.albumsPage.loading, d.albumsPage.err = true, ""
	offset := len(d.albums)
	go func() {
		albums, more, err := albumsBy(id, albumsPage, offset)
		a.post(func() {
			d.albumsPage.loading = false
			if err != nil {
				d.albumsPage.err = "专辑加载失败"
				return
			}
			d.albums = append(d.albums, albums...)
			d.albumsPage.loaded, d.albumsPage.more = true, more && len(albums) > 0
		})
	}()
}

// loadEvents loads the next page of an artist's events, once their profile
// says whose they are.
func (a *app) loadEvents(id int64) {
	d := a.artist(id)
	if d.profile == nil || d.profile.AccountID == 0 || d.eventsPage.loading {
		return
	}
	d.eventsPage.loading, d.eventsPage.err = true, ""
	uid, before := d.profile.AccountID, d.eventsBefore
	if !d.eventsPage.loaded {
		before = -1
	}
	go func() {
		events, last, more, err := accountEvents(uid, eventsPage, before)
		a.post(func() {
			d.eventsPage.loading = false
			if err != nil {
				d.eventsPage.err = "动态加载失败"
				return
			}
			d.events = append(d.events, events...)
			d.eventsPage.loaded = true
			// A page that starts no earlier would load forever.
			d.eventsPage.more = more && len(events) > 0 && (before == -1 || last < before)
			d.eventsBefore = last
		})
	}()
}

// toggleFollow follows an artist, or unfollows them, at once, going back if
// NetEase refuses. Signed out, it asks the listener to sign in.
func (a *app) toggleFollow(id int64) {
	d := a.artist(id)
	if d.profile == nil || d.following {
		return
	}
	if a.profile == nil {
		a.openLogin()
		return
	}
	ar := d.profile
	follow := !ar.Followed
	ar.Followed, d.following = follow, true
	go func() {
		err := followArtist(id, follow)
		a.post(func() {
			d.following = false
			switch {
			case err != nil:
				ar.Followed = !follow
				a.toast = "操作失败，请稍后再试"
			case follow:
				a.toast = "已关注"
			default:
				a.toast = "已取消关注"
			}
		})
	}()
}

// scrollPage scrolls a page of cards and rows, inset as one of tracks is:
// a tab of a page that lists tracks puts its header in the same place.
func (a *app) scrollPage(c *ui.Context, content func()) {
	a.scrollPageWithState(c, &a.pageState().scroll, content)
}

func (a *app) scrollPageWithState(c *ui.Context, state *ui.ScrollState, content func()) {
	top, bottom := a.insets(c)
	ui.Scroll(c.Key(a.route)).TrackScroll(state).Grow(1).Padding(top+pageGap, pagePad-12, bottom+32, pagePad-12).ScrollbarInsets(top, 0, bottom, 0).Children(func() {
		a.startMark(c)
		defer a.endMark(c)
		content()
	})
}

// listPageWithKey gives album grids the same row shell as track lists, keeping
// their shared tab control in one place as the tabs change.
func (a *app) listPageWithKey(c *ui.Context, key any, state *ui.ListState, content func()) {
	top, bottom := a.insets(c)
	// Detach the song tab's offset tracker before the list restores this
	// tab's row anchor. Both tabs share the animated header's identity.
	ui.List(c.Key(key), state, 1).TrackScroll(nil).Grow(1).
		Padding(top+pageGap, pagePad-12, bottom+32, pagePad-12).ScrollbarInsets(top, 0, bottom, 0).
		Rows(func(item ui.ListRow) {
			c := item.Context
			ui.Column(c).Children(func() {
				a.startMark(c)
				defer a.endMark(c)
				content()
			})
		})
}

// empty says a list has nothing in it.
func (a *app) empty(c *ui.Context, msg string) {
	ui.Text(c, msg).FontSize(13).TextColor(a.pal.Muted).Padding(40, 0).AlignSelf(ui.Center)
}

func (a *app) artistHeaderSkeleton(c *ui.Context, s shimmer) {
	ui.Row(c).Gap(28).AlignItems(ui.End).Padding(0, 0, 8, 0).Children(func() {
		s.Bone(c, artistPhotoSize/2).Size(artistPhotoSize, artistPhotoSize)
		ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(8).Children(func() {
			s.Text(c, 12, 1.3).Width(24)
			s.Text(c, 30, 1.2).WidthPercent(32)
			s.Text(c, 13, 1.3).Width(160)
			ui.Column(c).Margin(2, 0, 0, 0).Children(func() {
				s.Text(c, 12.5, 1.45).WidthPercent(78)
				s.Text(c, 12.5, 1.45).WidthPercent(52)
			})
			ui.Row(c).Gap(10).Margin(10, 0, 0, 0).Children(func() {
				s.Bone(c, 18).Size(92, 36)
				s.Bone(c, 18).Size(120, 36)
			})
		})
	})
	a.section(c, "歌曲")
}
