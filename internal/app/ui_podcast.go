package app

import (
	"slices"

	"lilt/internal/media/qqmusic"
)

// These load a podcast and its episodes from NetEase, and subscribe to it,
// and the podcasts the tab picks and the most popular; tests replace them.
var (
	podcastDetail    = qqmusic.PodcastDetail
	podcastPrograms  = qqmusic.PodcastPrograms
	subscribePodcast = qqmusic.SubscribePodcast

	recommendedPodcasts = qqmusic.RecommendedPodcasts
	hotPodcasts         = qqmusic.HotPodcasts
)

// programsPage is how many of a podcast's episodes load at a time.
const programsPage = 30

// podcastData is what the app has of a podcast: what it is, and its
// episodes, the newest first, as far as they have loaded.
type podcastData struct {
	podcast     *qqmusic.Podcast
	loading     bool
	err         string
	programs    []qqmusic.Program
	page        paged
	subscribing bool // a subscription or its end is on its way
	thenPlay    bool // its card's play button waits for it to load
}

func (a *app) loadPodcast(id int64) {
	d := a.podcasts[id]
	if d == nil {
		d = &podcastData{}
		a.podcasts[id] = d
	}
	if !d.page.loaded {
		a.loadPrograms(id)
	}
	if d.podcast != nil || d.loading {
		return
	}
	d.loading, d.err = true, ""
	go func() {
		pod, err := podcastDetail(id)
		a.post(func() {
			d.loading = false
			if err != nil {
				d.err, d.thenPlay = "播客加载失败", false
				return
			}
			d.podcast = pod
			a.playLoaded(d)
		})
	}()
}

// loadPrograms loads the next page of a podcast's episodes.
func (a *app) loadPrograms(id int64) {
	d := a.podcasts[id]
	if d.page.loading {
		return
	}
	d.page.loading, d.page.err = true, ""
	offset := len(d.programs)
	go func() {
		programs, more, err := podcastPrograms(id, programsPage, offset)
		a.post(func() {
			d.page.loading = false
			if err != nil {
				d.page.err, d.thenPlay = "节目加载失败", false
				return
			}
			d.programs = append(d.programs, programs...)
			d.page.loaded, d.page.more = true, more && len(programs) > 0
			// Which the listener likes, for their hearts.
			ids := make([]int64, len(programs))
			for i := range programs {
				ids[i] = programs[i].ID
			}
			a.askProgramLikes(ids)
			a.playLoaded(d)
		})
	}()
}

// playProgram plays the episodes loaded, from the one at i: those with
// audio to play.
func (a *app) playProgram(d *podcastData, i int) {
	var queue []qqmusic.Track
	at := 0
	for j := range d.programs {
		if d.programs[j].MainSong.ID == 0 {
			continue
		}
		if j <= i {
			at = len(queue)
		}
		queue = append(queue, d.programs[j].Track(d.podcast))
	}
	if len(queue) > 0 {
		a.pl.PlayQueue(queue, at)
	}
}

// playPodcast plays a podcast's episodes from the newest, from its card,
// once it and they have loaded.
func (a *app) playPodcast(id int64) {
	a.loadPodcast(id)
	d := a.podcasts[id]
	d.thenPlay = true
	a.playLoaded(d)
}

// playLoaded plays a podcast its card asked to play, once it can.
func (a *app) playLoaded(d *podcastData) {
	if d.thenPlay && d.podcast != nil && d.page.loaded {
		d.thenPlay = false
		a.playProgram(d, 0)
	}
}

// toggleSubscribe subscribes the listener to a podcast, or unsubscribes
// them, at once, going back if NetEase refuses. Signed out, it asks the
// listener to sign in.
func (a *app) toggleSubscribe(d *podcastData) {
	if d.podcast == nil || d.subscribing {
		return
	}
	if a.profile == nil {
		a.openLogin()
		return
	}
	pod := d.podcast
	tab := a.podcastTab
	sub := !pod.Subed
	pod.Subed, d.subscribing = sub, true
	if sub {
		pod.SubCount++
	} else {
		pod.SubCount--
	}
	go func() {
		err := subscribePodcast(pod.ID, sub)
		a.post(func() {
			d.subscribing = false
			switch {
			case err != nil:
				pod.Subed = !sub
				if sub {
					pod.SubCount--
				} else {
					pod.SubCount++
				}
				a.toast = "操作失败，请稍后再试"
			case sub:
				tab.subscribed(pod)
				a.toast = "已订阅"
			default:
				tab.unsubscribed(pod.ID)
				a.toast = "已取消订阅"
			}
		})
	}()
}

// podcastTab is what the podcasts tab lists: the podcasts the listener
// subscribes to, NetEase's picks, and the most popular, as far as they have
// loaded.
type podcastTab struct {
	subs    []qqmusic.Podcast // the latest subscribed first
	picks   []qqmusic.Podcast
	hot     []qqmusic.Podcast
	hotPage paged
	hotAt   int // how far into NetEase's list the next page starts
	loading bool
	err     string
}

func (a *app) loadPodcastTab() {
	t := a.podcastTab
	if t.hotPage.loaded || t.loading {
		return
	}
	t.loading, t.err = true, ""
	signedIn := a.profile != nil
	go func() {
		var subs []qqmusic.Podcast
		if signedIn {
			subs, _ = subscribedPodcasts()
		}
		picks, err1 := recommendedPodcasts()
		hot, more, err2 := hotPodcasts(programsPage, 0)
		a.post(func() {
			t.loading = false
			if err1 != nil && err2 != nil {
				t.err = "无法连接 QQ 音乐，请检查网络"
				return
			}
			t.subs, t.picks = subs, picks
			t.addHot(hot)
			t.hotPage = paged{loaded: true, more: more && len(hot) > 0}
			if err2 != nil {
				t.hotPage.err = "热门播客加载失败"
			}
		})
	}()
}

// loadHotPodcasts loads the next page of the most popular podcasts.
func (a *app) loadHotPodcasts() {
	t := a.podcastTab
	if t.hotPage.loading {
		return
	}
	t.hotPage.loading, t.hotPage.err = true, ""
	offset := t.hotAt
	go func() {
		hot, more, err := hotPodcasts(programsPage, offset)
		a.post(func() {
			t.hotPage.loading = false
			if err != nil {
				t.hotPage.err = "热门播客加载失败"
				return
			}
			t.addHot(hot)
			t.hotPage.more = more && len(hot) > 0
		})
	}()
}

// addHot lists a page of the most popular podcasts after those listed,
// but those already listed: NetEase's pages overlap, a podcast often
// coming again a page or two later.
func (t *podcastTab) addHot(page []qqmusic.Podcast) {
	t.hotAt += len(page)
	for _, pod := range page {
		if !slices.ContainsFunc(t.hot, func(p qqmusic.Podcast) bool { return p.ID == pod.ID }) {
			t.hot = append(t.hot, pod)
		}
	}
}

// subscribed lists a podcast the listener just subscribed to first, once
// the tab has loaded what they subscribe to.
func (t *podcastTab) subscribed(pod *qqmusic.Podcast) {
	if !t.hotPage.loaded {
		return
	}
	t.unsubscribed(pod.ID)
	t.subs = append([]qqmusic.Podcast{*pod}, t.subs...)
}

// unsubscribed takes a podcast out of the listener's subscriptions.
func (t *podcastTab) unsubscribed(id int64) {
	t.subs = slices.DeleteFunc(t.subs, func(p qqmusic.Podcast) bool { return p.ID == id })
}
