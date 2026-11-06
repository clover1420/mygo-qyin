package app

import (
	"lilt/internal/qqmusic"
)

// albumDetail loads an album from NetEase; tests replace it.
var albumDetail = qqmusic.AlbumDetail

// albumData is what the app has of an album.
type albumData struct {
	album                         *qqmusic.AlbumProfile
	loading                       bool
	err                           string
	likeUID                       int64
	liked, likeKnown, likeLoading bool
}

func (a *app) loadAlbum(id int64) {
	d := a.albums[id]
	if d == nil {
		d = &albumData{}
		a.albums[id] = d
	}
	if d.album != nil || d.loading {
		if d.album != nil {
			a.loadAlbumLike(d)
		}
		return
	}
	d.loading, d.err = true, ""
	go func() {
		al, err := albumDetail(id)
		a.post(func() {
			d.loading = false
			if err != nil {
				d.err = "专辑加载失败"
				return
			}
			d.album = al
			a.loadAlbumLike(d)
		})
	}()
}
