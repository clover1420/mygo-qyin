package app

import "lilt/internal/art"

// The app reads covers through these names while its views still live in
// this package; feature packages import lilt/internal/art directly.
type artwork = art.Artwork
type artCache = art.Cache

var (
	newArtCache = art.New

	fetchImage    = art.FetchImage
	decodeArt     = art.Decode
	dominantColor = art.DominantColor
	blurImage     = art.BlurImage
)
