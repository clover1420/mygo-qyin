package app

import (
	"errors"
	"fmt"
	"net"
	"strings"

	"lilt/internal/media/qqmusic"
)

// searchResults and searchSuggest search QQ Music; tests replace them.
var (
	searchResults     = qqmusic.SearchPage
	searchResultsFrom = qqmusic.SearchPageSource
	searchSuggest     = qqmusic.SearchSuggestContext
)

// searchLimit is how many results load at a time.
const searchLimit = 20

// searchData is what a search found, as far as its pages have loaded.
type searchData struct {
	res  qqmusic.SearchResult
	page paged
}

// search searches for q, of the kind last searched for.
func (a *app) search(q string) {
	q = strings.TrimSpace(q)
	if q == "" {
		return
	}
	kind := a.searched.Kind
	if kind == 0 {
		kind = qqmusic.SearchSongs
	}
	a.searchFor(route{Page: pageSearch, Query: q, Kind: kind})
}

// searchFor opens the page of a search, remembering its query. Searched
// for again, it loads anew.
func (a *app) searchFor(r route) {
	a.rememberSearch(r.Query)
	a.dismissSuggestions()
	if r == a.route {
		a.nowPlaying = false
		a.loadSearch(r, true)
		return
	}
	a.open(r)
}

// loadSearch loads the first page of a search, unless it has, or again.
func (a *app) loadSearch(r route, again bool) {
	if d := a.searches[r]; d != nil && !again {
		// A failed first page is not a cached result. Reopening it retries,
		// while loaded pages (including failed pagination) and active loads stay.
		if d.page.loaded || d.page.loading || d.page.err == "" {
			return
		}
	}
	d := &searchData{}
	a.searches[r] = d
	a.moreResults(r, d)
}

// moreResults loads the next page of a search. A search done again has
// new data, so the last one's pages land where nothing shows them.
func (a *app) moreResults(r route, d *searchData) {
	if d.page.loading {
		return
	}
	d.page.loading, d.page.err = true, ""
	offset := d.res.NextOffset
	if offset <= 0 {
		offset = d.res.Len()
	}
	source, fetch, fetchFrom := d.res.Source, searchResults, searchResultsFrom
	go func() {
		var res qqmusic.SearchResult
		var err error
		if source == qqmusic.SearchAuto {
			res, err = fetch(r.Query, r.Kind, searchLimit, offset)
		} else {
			res, err = fetchFrom(r.Query, r.Kind, searchLimit, offset, source)
		}
		a.post(func() {
			d.finishPage(res, err)
		})
	}()
}

func (d *searchData) finishPage(res qqmusic.SearchResult, err error) {
	d.page.loading = false
	if err != nil {
		d.page.err = searchErrorMessage(err)
		return
	}
	d.res.Append(res)
	offset := d.res.NextOffset
	if offset <= 0 {
		offset = d.res.Len()
	}
	// Follow the upstream page boundary when unusable entries were filtered.
	// A server can still claim more while sending none.
	d.page.loaded, d.page.more = true, res.Len() > 0 && offset < d.res.Total
}

func searchErrorMessage(err error) string {
	var cooldown *qqmusic.SearchCooldownError
	var protocol *qqmusic.ProtocolError
	var filtered *qqmusic.SearchFilterError
	var networkError net.Error
	switch {
	case errors.As(err, &cooldown):
		return "搜索暂时不可用，请稍后重试"
	case errors.As(err, &protocol) && protocol.Code == 2001:
		return "QQ 音乐拒绝搜索（2001），请确认登录后重试"
	case errors.As(err, &filtered):
		return fmt.Sprintf("QQ 音乐过滤了搜索结果（%d），请稍后重试", filtered.Filter)
	case errors.As(err, &networkError) && networkError.Timeout():
		return "搜索超时，请检查网络后重试"
	default:
		return "搜索失败，请稍后再试"
	}
}

// searchHistoryLimit is how many searches the history keeps.
const searchHistoryLimit = 50

// rememberSearch puts q first in the history of searches, once.
func (a *app) rememberSearch(q string) {
	history := []string{q}
	for _, s := range a.settings.SearchHistory {
		if s != q && len(history) < searchHistoryLimit {
			history = append(history, s)
		}
	}
	a.settings.SearchHistory = history
	a.settings.Save()
}

func (a *app) clearSearchHistory() {
	a.settings.SearchHistory = nil
	a.settings.Save()
}
