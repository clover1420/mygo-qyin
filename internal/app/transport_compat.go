package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/ui/base/transport"
)

// The playback controls live in lilt/internal/ui/base/transport; the app keeps
// these names while its views still live here.
type scrubState = transport.State
type barState = transport.Bar

var (
	secs         = transport.Secs
	drawBars     = transport.DrawBars
	drawSquiggly = transport.DrawSquiggly
	drawSmoke    = transport.DrawSmoke
	lineRandom   = transport.LineRandom
)

func (a *app) scrubber(c *ui.Context, st *scrubState, track, fill ui.Color, height float32, smoke ui.Color) ui.Element {
	return transport.Scrubber(c, a.pl, a.settings.ProgressEffectMode(), st, track, fill, height, smoke)
}

func (a *app) tickTimes(c *ui.Context) { transport.TickTimes(c, a.pl) }
