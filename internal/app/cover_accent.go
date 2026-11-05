package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/qqmusic"
)

// followCover keeps the cover of the song playing, so its color can accent
// every window. While a new cover loads, the last one's color stays.
func (a *app) followCover() {
	tr := a.pl.Current()
	if tr == nil {
		return
	}
	if tr.Album.PicURL == "" {
		a.coverArt = nil
		return
	}
	if art := a.art.get(qqmusic.Sized(tr.Album.PicURL, 600), false); art != nil && !art.failed {
		a.coverArt = art
	}
}

// coverAccent turns the cover's most vivid color into an accent that keeps
// white text legible. The hue stays, while saturation and value are
// constrained, as Cloudy does. A gray cover does not follow.
func (a *app) coverAccent() (ui.Color, bool) {
	art := a.coverArt
	if art == nil || art.failed {
		return ui.Color{}, false
	}
	h, s, _ := hsv(float64(art.tint.R)/255, float64(art.tint.G)/255, float64(art.tint.B)/255)
	if s < 0.12 {
		return ui.Color{}, false
	}
	return hsvColor(h, min(max(s, 0.5), 0.8), 0.8), true
}

// accentVariants are the interactive shades of an accent.
func accentVariants(base ui.Color) (hover, pressed ui.Color) {
	return base.Mix(paper, 0.12), base.Mix(ink, 0.12)
}

// applyTheme sets a window's theme for this frame. Every window calls it, so
// none keeps the fixed red while another follows the cover.
func (a *app) applyTheme(c *ui.Context) palette {
	t := *c.Theme()
	if a.settings.AccentCover {
		a.followCover()
	}
	p := paletteFor(&t)
	base := accentBase(t.Dark)
	if a.settings.AccentCover {
		if accent, ok := a.coverAccent(); ok {
			base = accent
		}
	}
	p.accent = base
	a.pal = p
	hover, pressed := accentVariants(base)
	t.Accent, t.AccentHover, t.AccentPressed = base, hover, pressed
	t.AccentText = p.onAccent
	t.Selection = base.Alpha(0.3)
	t.Focus = base.Alpha(0.55)
	t.Font = a.fontFamily()
	c.SetTheme(&t)
	return p
}

// progressFill is the played part of the player bar's progress. It takes the
// cover's accent with the rest of the controls, and keeps its neutral look
// while the setting is off.
func (a *app) progressFill(p palette) ui.Color {
	if a.settings.AccentCover {
		if accent, ok := a.coverAccent(); ok {
			return accent
		}
	}
	return p.text.Alpha(0.75)
}
