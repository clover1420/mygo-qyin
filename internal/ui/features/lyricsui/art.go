package lyricsui

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/art"
)

func bitmap(a *art.Artwork) *ui.Bitmap {
	if a == nil {
		return nil
	}
	return a.BMP
}
