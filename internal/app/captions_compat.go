package app

import "lilt/internal/media/captions"

// The captions model lives in lilt/internal/media/captions; the app keeps these
// names while the pane that shows it still lives here.
type captionState = captions.Model
type modelState = captions.ModelState
type ModelState = captions.ModelState
type transcript = captions.Transcript
type span = captions.Span

var (
	captionStart = captions.CaptionStart
	captionAfter = captions.CaptionAfter
	addSpan      = captions.AddSpan
	addSegment   = captions.AddSegment
	covered      = captions.Covered
)

// initCaptions gives the model what it needs from the app.
func (a *app) initCaptions() {
	a.captions.Deps = captions.Deps{
		Player:   a.pl,
		Settings: &a.settings,
		Post:     a.post,
		Redraw:   a.redraw,
	}
}

// setCaptionState replaces the model's state, keeping its deps, as tests do.
func (a *app) setCaptionState(state captionState) {
	a.captions = state
	a.initCaptions()
}

func (a *app) captionsShown() bool { return a.captions.Shown() }
func (a *app) captionsOn() bool    { return a.captions.On() }

func (a *app) startModelDownload()     { a.captions.StartDownload() }
func (a *app) cancelModelDownload()    { a.captions.CancelDownload() }
func (a *app) syncCaptions()           { a.captions.Sync() }
func (a *app) transcriptWrite() func() { return a.captions.TranscriptWrite() }
