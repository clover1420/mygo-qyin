package app

import (
	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
	"lilt/internal/ui/base/skeleton"
	accountpkg "lilt/internal/ui/features/account"
)

// The account and its login live in lilt/internal/ui/features/account; the app keeps
// the login's state and these names while its views still live here.
type loginState = accountpkg.State

// beginLoginQR is replaced by UI tests so retry actions never contact QQ.
var beginLoginQR = qqmusic.QRKey
var beginWechatLoginQR = qqmusic.WechatQRKey

// boot signs in with the session in the keychain, if any, then loads the
// home page for whoever is listening.
func (a *app) boot() {
	a.homeLoading = true
	go func() {
		var prof *qqmusic.Profile
		if qqmusic.RestoreSession() {
			p, err := qqmusic.Account()
			switch {
			case err == nil && p == nil:
				qqmusic.ClearSession() // expired
			case err == nil:
				prof = p
			}
		}
		a.post(func() {
			a.signedIn(prof)
		})
	}()
}

// signedIn shows what is personal to prof, or to nobody.
func (a *app) signedIn(prof *qqmusic.Profile) {
	a.profile = prof
	// Playback picks up where it was left, once the session that lets
	// the listener play their songs is back.
	a.restoreSession()
	a.mine, a.saved, a.recPlaylists, a.liked = nil, nil, nil, nil
	a.clearSongShelves()
	a.likedPrograms, a.askedPrograms = nil, nil
	a.likedAlbums = nil
	a.likedScroll = likedScrollState{}
	// What the listener may play, and follows, changed: load the pages
	// again.
	a.playlists = map[int64]*qqmusic.Playlist{}
	a.plErr = map[int64]string{}
	a.artists, a.albums, a.users, a.podcasts = map[int64]*artistData{}, map[int64]*albumData{}, map[int64]*userData{}, map[int64]*podcastData{}
	a.podcastTab = &podcastTab{}
	a.purchased = [2]*purchasedData{}
	a.purchasedScroll = [2]listScrollState{}
	a.searches = map[route]*searchData{}
	a.load(a.route)
	a.loadHome()
	if prof == nil {
		return
	}
	a.loadLiked(prof.UserID)
	a.loadMine(prof.UserID)
}

// accountDeps gathers what the account needs, once a frame.
func (a *app) accountDeps() accountpkg.Deps {
	return accountpkg.Deps{
		Palette:     a.pal,
		Art:         a.art,
		Login:       &a.login,
		AccountOpen: &a.accountOpen,
		Profile:     func() *qqmusic.Profile { return a.profile },
		Post:        a.post,
		SignedIn:    a.signedIn,
		BeginQR:     func() (string, error) { return beginLoginQR() },
		BeginWechatQR: func() (string, error) {
			return beginWechatLoginQR()
		},
		Pane: a.pane,
		Line: func() ui.Color { return a.inkOn(false).line },
		QRSkeleton: func(c *ui.Context, pad float32) ui.Element {
			return skeleton.QR(c, shimmer{Base: shade(3), Sheen: shade(1)}.In(c), pad)
		},
		Sidebar:   a.sidebarSurface,
		Avatar:    a.avatar,
		OpenLogin: a.openLogin,
		OpenUser:  func(id int64) { a.open(route{Page: pageUser, ID: id}) },
		SignOut:   a.signOut,
	}
}

func (a *app) openLogin() { a.openLoginChannel(false) }

func (a *app) openLoginChannel(wechat bool) {
	accountpkg.Start(a.accountDeps(), wechat)
}

// acceptLoginQR applies a worker result on the main thread before polling starts.
func (a *app) acceptLoginQR(gen int, key string, code *ui.Bitmap, err error) bool {
	return accountpkg.Accept(a.accountDeps(), gen, key, code, err)
}

func (a *app) signOut() {
	accountpkg.Close(a.accountDeps())
	qqmusic.Logout()
	a.signedIn(nil)
}

func (a *app) loginDialog(c *ui.Context) {
	accountpkg.Dialog(c, a.accountDeps())
}

func (a *app) accountRow(c *ui.Context) {
	accountpkg.Row(c, a.accountDeps())
}
