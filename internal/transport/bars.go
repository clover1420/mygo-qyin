package transport

import (
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/player"
)

// Bar draws n bars of an equalizer in r, from the levels of what plays,
// smoothed in st between frames; idle bars rest low.
type Bar struct {
	h    [player.Bands]float32
	last time.Time
}

func (st *Bar) Step(now time.Time, target [player.Bands]float32, live bool) {
	dt := float32(now.Sub(st.last).Seconds())
	if st.last.IsZero() || dt > 0.2 {
		dt = 0.016
	}
	st.last = now
	for i := range st.h {
		tg := target[i]
		if !live {
			tg = 0
		}
		k := float32(18) // rises fast
		if tg < st.h[i] {
			k = 6 // falls slowly
		}
		st.h[i] += (tg - st.h[i]) * min(1, k*dt)
	}
}

// Group returns the mean of bands [from, to).
func (st *Bar) Group(from, to int) float32 {
	var s float32
	for i := from; i < to; i++ {
		s += st.h[i]
	}
	return s / float32(to-from)
}

func DrawBars(p *ui.Painter, r ui.Rect, st *Bar, n int, color ui.Color, minH float32) {
	gap := r.W / float32(n) * 0.38
	w := (r.W - gap*float32(n-1)) / float32(n)
	per := player.Bands / n
	for i := range n {
		// Shape the response: low bands are loud already.
		v := st.Group(i*per, (i+1)*per)
		v = v * v * 1.15
		h := max(minH, min(r.H, r.H*v))
		x := r.X + float32(i)*(w+gap)
		p.Fill(ui.Rect{X: x, Y: r.Y + (r.H-h)/2, W: w, H: h}, color, w/2)
	}
}
