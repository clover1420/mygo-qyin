// Package skeleton draws the gray blocks that stand in for what is loading:
// a lighter band sweeps over them, across all the blocks of a view in step.
package skeleton

import (
	"time"

	"github.com/egoist/mygo/ui"
)

// Shimmer is how the blocks of a skeleton look.
type Shimmer struct {
	Base, Sheen ui.Color
	span        float32 // the width the band crosses: the window's
	still       bool    // the desktop asks for less motion: no band
}

const (
	period = 1600 * time.Millisecond
	sweep  = 1100 * time.Millisecond // of the period, the band crossing; then it rests
	band   = 260                     // DIPs
)

// In fits the shimmer to the window it is drawn in.
func (s Shimmer) In(c *ui.Context) Shimmer {
	s.span, _ = c.Size()
	s.still = c.Preferences().ReduceMotion
	return s
}

// Paint fills r, its corners rounded by radius, as a block of the skeleton.
func (s Shimmer) Paint(p *ui.Painter, r ui.Rect, radius float32) {
	if s.still || r.W <= 0 {
		p.Fill(r, s.Base, radius)
		return
	}
	at := time.Duration(p.Now().UnixNano() % int64(period))
	if at >= sweep {
		p.Fill(r, s.Base, radius)
		p.After(period - at)
		return
	}
	p.AnimationFrameRate(60)
	t := float32(at) / float32(sweep)
	x := -band/2 + t*(s.span+band)
	if x+band/2 < r.X || x-band/2 > r.X+r.W {
		p.Fill(r, s.Base, radius)
		return
	}
	// The band's place along the block, which it may lie partly beyond.
	along := func(x float32) float32 { return (x - r.X) / r.W }
	p.FillGradient(r, ui.LinearGradient{Angle: 90, Stops: []ui.GradientStop{
		{Color: s.Base, At: along(x - band/2)},
		{Color: s.Sheen, At: along(x)},
		{Color: s.Base, At: along(x + band/2)},
	}}, radius)
}

// Bone is a block of the skeleton, as big as it is laid out.
func (s Shimmer) Bone(c *ui.Context, radius float32) ui.Element {
	return ui.Box(c).Shrink(0).Draw(func(p *ui.Painter, r ui.Rect) { s.Paint(p, r, radius) })
}

// Text is a line of text at size, as tall as the line and as wide as it is
// laid out, its block the height of the letters.
func (s Shimmer) Text(c *ui.Context, size, lineHeight float32) ui.Element {
	return ui.Box(c).Shrink(0).Height(size * lineHeight).Draw(func(p *ui.Painter, r ui.Rect) {
		h := size * 0.78
		s.Paint(p, ui.Rect{X: r.X, Y: r.Y + (r.H-h)/2, W: r.W, H: h}, h*0.32)
	})
}

// boneWidths vary the lines of a skeleton, as the lengths of what they
// stand for do, as percentages.
var boneWidths = [...]float32{62, 44, 78, 53, 70, 38, 66, 49, 82, 57, 41, 73}

// BoneWidth returns the width of line i, as a percentage.
func BoneWidth(i int) float32 { return boneWidths[i%len(boneWidths)] }

// Comment is n of commentRow's comments.
func Comment(c *ui.Context, s Shimmer, n int) {
	for i := range n {
		ui.Row(c.Key(i)).Padding(12).Gap(12).AlignItems(ui.Start).Children(func() {
			s.Bone(c, 17).Size(34, 34)
			ui.Column(c).Grow(1).Shrink(1).MinWidth(0).Gap(6).Children(func() {
				s.Text(c, 13, 1.3).Width(60 + BoneWidth(i))
				ui.Column(c).Children(func() {
					s.Text(c, 15, 1.55).WidthPercent(94)
					if i%3 != 2 {
						s.Text(c, 15, 1.55).WidthPercent(BoneWidth(i + 2))
					}
				})
				s.Text(c, 11.5, 1.3).Width(96)
			})
		})
	}
}

// Lyrics is lyricsPane's lines while they load.
func Lyrics(c *ui.Context, s Shimmer) {
	ui.Column(c).Gap(6).Padding(0, 8, 0, 0).Children(func() {
		for i := range 7 {
			ui.Box(c.Key(i)).Padding(10, 12).Children(func() {
				s.Text(c, 30, 1.25).WidthPercent(BoneWidth(i+1) * 0.9)
			})
		}
	})
}

// Stage is stageLine's words while the lyrics load.
func Stage(c *ui.Context, s Shimmer, f, w float32) {
	ui.Row(c).Wrap().Justify(ui.Center).AlignItems(ui.Center).Padding(0, w*0.1).Gap(f * 0.35).Children(func() {
		for i, n := range []float32{2.2, 1.3, 3.1, 1.8} {
			s.Bone(c.Key(i), f*0.22).Size(f*n, f*0.85)
		}
	})
}

// QR fills a QR code's room with what looks like one, at its size, its
// quiet zone pad: three finders, and modules between them.
func QR(c *ui.Context, s Shimmer, pad float32) ui.Element {
	return ui.Box(c).Draw(func(p *ui.Painter, r ui.Rect) {
		const n = 25
		cell := (min(r.W, r.H) - 2*pad) / n
		x0, y0 := r.X+(r.W-cell*n)/2, r.Y+(r.H-cell*n)/2
		at := func(x, y, w, h int) ui.Rect {
			return ui.Rect{X: x0 + float32(x)*cell, Y: y0 + float32(y)*cell, W: float32(w) * cell, H: float32(h) * cell}
		}
		finder := func(x, y int) bool { return (x < 8 || x >= n-8) && y < 8 || x < 8 && y >= n-8 }
		for _, f := range [][2]int{{0, 0}, {n - 7, 0}, {0, n - 7}} {
			x, y := f[0], f[1]
			for _, e := range []ui.Rect{at(x, y, 7, 1), at(x, y+6, 7, 1), at(x, y+1, 1, 5), at(x+6, y+1, 1, 5)} {
				s.Paint(p, e, 0)
			}
			s.Paint(p, at(x+2, y+2, 3, 3), cell*0.4)
		}
		// Modules, a run of them at a time, in a fixed pattern.
		for y := range n {
			run := -1
			for x := 0; x <= n; x++ {
				h := uint32(x*73856093^y*19349663) * 2654435761
				on := x < n && !finder(x, y) && h>>29 < 4
				switch {
				case on && run < 0:
					run = x
				case !on && run >= 0:
					s.Paint(p, at(run, y, x-run, 1), cell*0.3)
					run = -1
				}
			}
		}
	})
}
