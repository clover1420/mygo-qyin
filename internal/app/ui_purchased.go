package app

import (
	"lilt/internal/media/qqmusic"
)

const (
	purchasedAlbums = iota
	purchasedSongs
	purchasedLimit = 30
)

// These read the current account's purchases; tests replace them.
var (
	readPurchasedSongs  = qqmusic.PurchasedSongs
	readPurchasedAlbums = qqmusic.PurchasedAlbums
)

type purchasedData struct {
	songs  []qqmusic.Track
	albums []qqmusic.AlbumProfile
	page   paged
}

func (a *app) loadPurchased(tab int, again bool) {
	if a.profile == nil {
		return
	}
	if a.purchased[tab] != nil && !again {
		return
	}
	d := &purchasedData{}
	a.purchased[tab] = d
	a.morePurchased(tab, d)
}

func (a *app) morePurchased(tab int, d *purchasedData) {
	if a.profile == nil || d.page.loading {
		return
	}
	d.page.loading, d.page.err = true, ""
	offset := len(d.songs) + len(d.albums)
	go func() {
		var songs []qqmusic.Track
		var albums []qqmusic.AlbumProfile
		var more bool
		var err error
		if tab == purchasedSongs {
			songs, more, err = readPurchasedSongs(purchasedLimit, offset)
		} else {
			albums, more, err = readPurchasedAlbums(purchasedLimit, offset)
		}
		a.post(func() {
			d.page.loading = false
			if err != nil {
				d.page.err = "已购音乐加载失败"
				return
			}
			d.songs = append(d.songs, songs...)
			d.albums = append(d.albums, albums...)
			d.page.loaded, d.page.more = true, more && len(songs)+len(albums) > 0
		})
	}()
}
