// Package captions writes down an episode as it plays, on the device, and
// keeps the Transcript on disk. It owns the model and the transcription
// running; the app owns the pane that shows it.
package captions

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"sync"
	"time"

	"github.com/egoist/mygo"

	"lilt/internal/core/settings"
	"lilt/internal/media/lyrics"
	"lilt/internal/media/player"
	"lilt/internal/media/speech"
)

// Deps are what the model needs from the app: what plays, the choices, and
// where to publish changes.
type Deps struct {
	Player   *player.Player
	Settings *settings.Settings
	Post     func(func())
	Redraw   func()
}

// An episode's captions are written down on the device, as it plays:
// from where the listener is, ahead of them, and kept on disk, so that
// what was written down once is there the next time. A seek past what is
// written down starts over from there, leaving what was for later.

// These find, download and run the model; tests replace them.
var (
	Available      = speech.Available
	Installed      = speech.Installed
	Download       = speech.Download
	Remove         = speech.Remove
	NewTranscriber = func(dir string) (transcriber, error) { return speech.NewRecognizer(dir) }
)

// A transcriber writes down what an episode says (speech.Recognizer).
type transcriber interface {
	Transcribe(ctx context.Context, src speech.Source, from, until time.Duration, said func(speech.Segment), heard func(time.Duration)) error
	Close()
}

// The model loaded, held while an episode is written down, and the lock
// that one transcription at a time holds it under: a new one waits for
// the one it replaces to stop.
var recognizer struct {
	sync.Mutex
	t transcriber
}

// captionLead is how far ahead of the listener the transcription may be
// for a seek to wait for it, rather than start another from there.
const captionLead = 30 * time.Second

// Model is what the app has of captions: the model, and the episode
// playing's Transcript.
type Model struct {
	Deps

	Model ModelState

	Program int64       // the episode the Transcript is of, 0 for none
	Tr      *Transcript // nil while it loads
	Lines   []lyrics.Line
	Work    *Work         // the transcription running, or nil
	Failed  int64         // the episode the transcription failed for
	Saved   time.Duration // how far the Transcript on disk is written down
}

// ModelState is the model's: whether it is installed, or downloading.
type ModelState struct {
	Checked, Installed bool
	Downloading        bool
	Done, Total        int64
	Err                string
	Cancel             context.CancelFunc
}

// A Transcript is what is written down of an episode, and which stretches
// of it all is written down for.
type Transcript struct {
	Model    string           `json:"model"`
	Segments []speech.Segment `json:"segments"`
	Spans    []Span           `json:"spans"`
}

// A Span is a stretch of an episode, from its first to its second.
type Span [2]time.Duration

// Work is a transcription running from From up to Until, 0 for the
// episode's end, written down as far as At.
type Work struct {
	Program     int64
	From, Until time.Duration
	At          time.Duration
	Cancel      context.CancelFunc
}

// modelDir is where the model and the libraries that run it download to.
func modelDir() (string, error) {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "models"), nil
}

func transcriptPath(program int64) (string, error) {
	dir, err := mygo.App.Path(mygo.PathCache)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "transcripts", strconv.FormatInt(program, 10)+".json"), nil
}

// Shown reports whether the app can make captions on this
// platform. The listener may still have to download the model.
func (m *Model) Shown() bool { return Available() }

// Ready reports whether the model is installed, looking once.
func (m *Model) Ready() bool {
	st := &m.Model
	if !st.Checked {
		st.Checked = true
		if dir, err := modelDir(); err == nil {
			st.Installed = Installed(dir)
		}
	}
	return st.Installed
}

// On reports whether episodes are written down as they play.
func (m *Model) On() bool {
	return m.Settings.Captions && m.Shown() && m.Ready()
}

// StartDownload downloads the model, its progress shown as it goes.
func (m *Model) StartDownload() {
	st := &m.Model
	if st.Downloading || st.Installed {
		return
	}
	dir, err := modelDir()
	if err != nil {
		st.Err = "无法保存模型"
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	st.Downloading, st.Err, st.Cancel = true, "", cancel
	st.Done, st.Total = 0, speech.ModelSize()
	go func() {
		// Shown in steps of a megabyte: the download reads far smaller
		// pieces.
		var shown int64
		err := Download(ctx, dir, func(done, total int64) {
			if done-shown < 1<<20 && done < total {
				return
			}
			shown = done
			m.Post(func() {
				st.Done, st.Total = done, total
				m.Redraw()
			})
		})
		m.Post(func() {
			m.Redraw()
			st.Downloading, st.Cancel = false, nil
			switch {
			case err == nil:
				st.Installed, st.Checked = true, true
				m.Settings.Captions = true
				m.Settings.Save()
				m.Sync()
			case ctx.Err() != nil:
			default:
				st.Err = "下载失败，请检查网络后重试"
			}
		})
	}()
}

// CancelDownload stops the download, keeping what came in for the
// next.
func (m *Model) CancelDownload() {
	if st := &m.Model; st.Cancel != nil {
		st.Cancel()
	}
}

// Delete stops the transcription, lets the model go and removes it.
func (m *Model) Delete() {
	st := &m.Model
	if !st.Installed {
		return
	}
	m.Stop()
	st.Installed = false
	dir, err := modelDir()
	if err != nil {
		return
	}
	go func() {
		recognizer.Lock()
		if recognizer.t != nil {
			recognizer.t.Close()
			recognizer.t = nil
		}
		Remove(dir)
		recognizer.Unlock()
	}()
}

// Sync keeps the episode playing written down ahead of the
// listener: it loads its Transcript, and starts a transcription from the
// first of what is not written down from where they are, unless one
// running gets there soon. It runs whenever the player changes.
func (m *Model) Sync() {
	cs := m
	var program int64
	if tr := m.Player.Current(); tr != nil {
		program = tr.ProgramID
	}
	if program != cs.Program {
		m.Pause()
		m.saveTranscript()
		cs.Program, cs.Tr, cs.Lines = program, nil, nil
		if program != 0 && m.Shown() {
			m.loadTranscript(program)
		}
	}
	if program == 0 || !m.On() || cs.Failed == program {
		m.Stop()
		return
	}
	if cs.Tr == nil || !m.Player.Ready() {
		// The model stays for once the episode or its Transcript loads.
		m.Pause()
		return
	}
	from, until, ok := CaptionStart(cs.Tr.Spans, m.Player.Position(), m.Player.Duration())
	if w := cs.Work; w != nil {
		if ok && w.From <= from && from <= w.At+captionLead {
			return
		}
		w.Cancel()
		cs.Work = nil
	}
	if !ok {
		releaseRecognizer()
		return
	}
	m.transcribe(program, from, until)
}

// Stop stops the transcription, if one runs, and lets the model
// go once it has.
func (m *Model) Stop() {
	m.Pause()
	releaseRecognizer()
}

// Pause stops the transcription, if one runs, keeping the model.
func (m *Model) Pause() {
	if w := m.Work; w != nil {
		w.Cancel()
		m.Work = nil
	}
}

// releaseRecognizer lets the model go, once no transcription holds it.
func releaseRecognizer() {
	go func() {
		recognizer.Lock()
		defer recognizer.Unlock()
		if recognizer.t != nil {
			recognizer.t.Close()
			recognizer.t = nil
		}
	}()
}

// transcribe writes down the episode playing from from up to until, on a
// goroutine.
func (m *Model) transcribe(program int64, from, until time.Duration) {
	src, err := m.Player.OpenAudio()
	if err != nil {
		return
	}
	if err := src.SetPosition(int64(from.Seconds() * float64(src.SampleRate()))); err != nil {
		src.Close()
		return
	}
	dir, err := modelDir()
	if err != nil {
		src.Close()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	w := &Work{Program: program, From: from, Until: until, At: from, Cancel: cancel}
	m.Work = w
	go func() {
		defer src.Close()
		recognizer.Lock()
		defer recognizer.Unlock()
		if ctx.Err() != nil {
			return
		}
		var err error
		if recognizer.t == nil {
			recognizer.t, err = NewTranscriber(dir)
		}
		if err == nil {
			err = recognizer.t.Transcribe(ctx, src, from, until,
				func(s speech.Segment) { m.Post(func() { m.captionSaid(w, s) }) },
				func(at time.Duration) { m.Post(func() { m.captionHeard(w, at) }) })
		}
		m.Post(func() { m.captionDone(w, err, ctx.Err() != nil) })
	}()
}

// captionSaid adds a segment written down to the Transcript of its
// episode, unless the Transcript has it: a transcription started again
// from where one stopped writes down again what that one did last.
func (m *Model) captionSaid(w *Work, s speech.Segment) {
	cs := m
	if w.Program != cs.Program || cs.Tr == nil {
		return
	}
	cs.Tr.Segments = AddSegment(cs.Tr.Segments, s)
	cs.Lines = speech.Lines(cs.Tr.Segments)
}

// captionHeard marks the episode written down from where w started to at.
func (m *Model) captionHeard(w *Work, at time.Duration) {
	cs := m
	if w.Program != cs.Program || cs.Tr == nil {
		return
	}
	w.At = max(w.At, at)
	cs.Tr.Spans = AddSpan(cs.Tr.Spans, Span{w.From, at})
	// Kept a minute at a time, should the app not quit as it should.
	if at > cs.Saved+time.Minute || at < cs.Saved {
		cs.Saved = at
		m.saveTranscript()
	}
}

// captionDone ends w: all it was to write down is, unless it was stopped
// or failed. Then the next of what is not written down follows.
func (m *Model) captionDone(w *Work, err error, stopped bool) {
	cs := m
	if cs.Work == w {
		cs.Work = nil
	}
	if w.Program != cs.Program || cs.Tr == nil || stopped {
		return
	}
	if err != nil {
		cs.Failed = w.Program
		m.saveTranscript()
		return
	}
	to := w.Until
	if to == 0 {
		to = max(w.At, m.Player.Duration())
	}
	cs.Tr.Spans = AddSpan(cs.Tr.Spans, Span{w.From, to})
	m.saveTranscript()
	m.Sync()
}

// Written reports whether all of the episode playing is written
// down.
func (m *Model) Written() bool {
	cs := m
	if cs.Tr == nil {
		return false
	}
	_, _, ok := CaptionStart(cs.Tr.Spans, 0, m.Player.Duration())
	return !ok
}

// Writing reports whether what is at pos is being written down
// and not yet there.
func (m *Model) Writing(pos time.Duration) bool {
	cs := m
	if cs.Tr == nil {
		return m.On() && cs.Program != 0
	}
	if _, ok := Covered(cs.Tr.Spans, pos); ok {
		return false
	}
	return cs.Work != nil
}

// loadTranscript reads the Transcript of an episode on a goroutine, or
// starts one with nothing in it.
func (m *Model) loadTranscript(program int64) {
	path, err := transcriptPath(program)
	if err != nil {
		m.Tr = &Transcript{Model: speech.ModelName}
		return
	}
	go func() {
		tr := &Transcript{}
		if b, err := os.ReadFile(path); err != nil || json.Unmarshal(b, tr) != nil || tr.Model != speech.ModelName {
			tr = &Transcript{Model: speech.ModelName}
		}
		m.Post(func() {
			cs := m
			if cs.Program == program && cs.Tr == nil {
				cs.Tr = tr
				cs.Lines = speech.Lines(tr.Segments)
				m.Sync()
			}
		})
	}()
}

// saveTranscript keeps the Transcript of the episode on disk, on a
// goroutine.
func (m *Model) saveTranscript() {
	if write := m.TranscriptWrite(); write != nil {
		go write()
	}
}

// transcriptWrite returns what writes the Transcript of the episode as
// it is now, or nil with nothing to write.
func (m *Model) TranscriptWrite() func() {
	cs := m
	if cs.Program == 0 || cs.Tr == nil || len(cs.Tr.Spans) == 0 {
		return nil
	}
	path, err := transcriptPath(cs.Program)
	if err != nil {
		return nil
	}
	b, err := json.Marshal(cs.Tr)
	if err != nil {
		return nil
	}
	return func() {
		if os.MkdirAll(filepath.Dir(path), 0o755) == nil {
			tmp := path + ".tmp"
			if os.WriteFile(tmp, b, 0o644) == nil {
				os.Rename(tmp, path)
			}
		}
	}
}

// CaptionStart returns where to write down from for a listener at pos of
// an episode length long, of which spans are written down: pos, or the
// end of the Span it is in, or once all from there on is written down,
// the first of what is not before it; and up to where: the start of the
// next Span, or 0 for the end. It reports false once all of it is
// written down.
func CaptionStart(spans []Span, pos, length time.Duration) (from, until time.Duration, ok bool) {
	if from, until, ok = CaptionAfter(spans, pos, length); ok || pos == 0 {
		return from, until, ok
	}
	return CaptionAfter(spans, 0, length)
}

// CaptionAfter is CaptionStart for what is from pos on alone.
func CaptionAfter(spans []Span, pos, length time.Duration) (from, until time.Duration, ok bool) {
	from = pos
	if end, in := Covered(spans, pos); in {
		from = end
	}
	if length > 0 && from >= length-time.Second {
		return 0, 0, false
	}
	for _, s := range spans {
		if s[0] > from {
			return from, s[0], true
		}
	}
	return from, 0, true
}

// Covered returns the end of the Span holding t, if one does.
func Covered(spans []Span, t time.Duration) (time.Duration, bool) {
	for _, s := range spans {
		if s[0] <= t && t < s[1] {
			return s[1], true
		}
	}
	return 0, false
}

// AddSpan adds s to spans, sorted, merging those it meets.
func AddSpan(spans []Span, s Span) []Span {
	if s[1] <= s[0] {
		return spans
	}
	out := make([]Span, 0, len(spans)+1)
	for _, o := range spans {
		if o[1] < s[0] || o[0] > s[1] {
			out = append(out, o)
			continue
		}
		s = Span{min(s[0], o[0]), max(s[1], o[1])}
	}
	out = append(out, s)
	slices.SortFunc(out, func(x, y Span) int { return int(x[0] - y[0]) })
	return out
}

// AddSegment adds s to segs, sorted, unless it is mostly over one there.
func AddSegment(segs []speech.Segment, s speech.Segment) []speech.Segment {
	i, _ := slices.BinarySearchFunc(segs, s.Start, func(x speech.Segment, t time.Duration) int { return int(x.Start - t) })
	for _, o := range segs[max(i-1, 0):min(i+1, len(segs))] {
		if overlap := min(o.End, s.End) - max(o.Start, s.Start); overlap > (s.End-s.Start)/2 {
			return segs
		}
	}
	return slices.Insert(segs, i, s)
}
