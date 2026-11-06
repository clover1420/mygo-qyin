package app

import (
	"errors"
	"image"
	"image/color"
	"image/png"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"

	"lilt/internal/media/qqmusic"
)

// An authorization error after scanning must offer retry instead of leaving the scanned overlay stuck.
func TestLoginErrorAfterScanCanRefresh(t *testing.T) {
	asked := make(chan struct{}, 1)
	old := beginLoginQR
	beginLoginQR = func() (string, error) {
		asked <- struct{}{}
		return "", errors.New("offline")
	}
	t.Cleanup(func() { beginLoginQR = old })

	a := newTestApp()
	a.settings = defaultSettings()
	a.login = loginState{Open: true, Gen: 7, Key: "old-key", Status: qqmusic.QRScanned, Scanner: "listener", Err: "authorization failed"}
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if err := tt.Click("刷新二维码"); err != nil {
		t.Fatal(err)
	}
	if !a.login.Open || a.login.Gen != 8 || a.login.Status != 0 || a.login.Key != "" || a.login.Err != "" || a.login.Scanner != "" {
		t.Fatalf("retry retained the failed login: %+v", a.login)
	}
	wait(t, asked, "a new QQ login QR")
	tt.Frame()
}

// A failed QR fetch must permit another refresh, and its successful response must replace the old state.
func TestLoginQRRefreshAfterFetchFailure(t *testing.T) {
	asked := make(chan struct{}, 2)
	old := beginLoginQR
	beginLoginQR = func() (string, error) { asked <- struct{}{}; return "", errors.New("offline") }
	t.Cleanup(func() { beginLoginQR = old })
	a := newTestApp()
	a.settings = defaultSettings()
	a.login = loginState{Open: true, Gen: 1, Err: "previous failure"}
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if err := tt.Click("刷新二维码"); err != nil {
		t.Fatal(err)
	}
	wait(t, asked, "the first QR refresh")
	first := a.login.Gen
	failure := &qqmusic.LoginError{Step: "qr-show", Err: &url.Error{Op: "Get", URL: "https://xui.ptlogin2.qq.com/ssl/ptqrshow?private=fake", Err: io.EOF}}
	if a.acceptLoginQR(first, "", nil, failure) || a.login.Err == "" || a.login.Key != "" || a.login.Code != nil {
		t.Fatal("failed fetch left a ready QR")
	}
	tt.Frame()
	if err := tt.Click("刷新二维码"); err != nil {
		t.Fatal(err)
	}
	wait(t, asked, "the second QR refresh")
	second := a.login.Gen
	bitmap := ui.NewBitmap(image.NewRGBA(image.Rect(0, 0, 37, 37)))
	if a.acceptLoginQR(first, "stale", bitmap, nil) {
		t.Fatal("old QR response replaced the new refresh")
	}
	if !a.acceptLoginQR(second, "fresh", bitmap, nil) {
		t.Fatal("fresh QR was discarded")
	}
	if a.login.Key != "fresh" || a.login.Code != bitmap || a.login.Status != qqmusic.QRWaiting || a.login.Err != "" {
		t.Fatal("fresh QR did not replace the failed fetch")
	}
	tt.Frame()
}

// Switching login channels replaces the QR generation, and refresh keeps the selected channel.
func TestLoginSwitchesToWechatAndRefreshes(t *testing.T) {
	asked := make(chan struct{}, 2)
	old := beginWechatLoginQR
	beginWechatLoginQR = func() (string, error) { asked <- struct{}{}; return "", errors.New("offline") }
	t.Cleanup(func() { beginWechatLoginQR = old })
	a := newTestApp()
	a.settings = defaultSettings()
	a.login = loginState{Open: true, Gen: 4, Key: "old-qq", Status: qqmusic.QRWaiting}
	tt := ui.NewTester(a.view, 1000, 700)
	tt.Frame()
	if err := tt.Click("微信"); err != nil {
		t.Fatal(err)
	}
	wait(t, asked, "a WeChat QR")
	if !a.login.Wechat || a.login.Gen != 5 || a.login.Key != "" {
		t.Fatal("switch retained QQ login")
	}
	if a.acceptLoginQR(4, "stale-qq", nil, nil) {
		t.Fatal("old QQ result replaced WeChat login")
	}
	a.acceptLoginQR(5, "", nil, errors.New("offline"))
	tt.Frame()
	if err := tt.Click("刷新二维码"); err != nil {
		t.Fatal(err)
	}
	wait(t, asked, "a refreshed WeChat QR")
	if !a.login.Wechat || a.login.Gen != 6 {
		t.Fatal("refresh changed the login channel")
	}
	tt.Frame()
}

// With NM_FRAMES set, saves the WeChat login layout using an inert QR-like image.
func TestWechatLoginFrames(t *testing.T) {
	dir := os.Getenv("NM_FRAMES")
	if dir == "" {
		t.Skip()
	}
	qr := image.NewRGBA(image.Rect(0, 0, 45, 45))
	for y := 0; y < 45; y++ {
		for x := 0; x < 45; x++ {
			c := color.White
			if x > 3 && x < 41 && y > 3 && y < 41 && ((x*11+y*7+x*y)%5 < 2) {
				c = color.Black
			}
			qr.Set(x, y, c)
		}
	}
	for _, dark := range []bool{false, true} {
		a := newTestApp()
		a.settings = defaultSettings()
		a.login = loginState{Open: true, Wechat: true, Status: qqmusic.QRWaiting, Code: ui.NewBitmap(qr)}
		tt := ui.NewTester(a.view, 1000, 700)
		tt.SetDark(dark)
		tt.Advance(time.Second)
		tt.Frame()
		name := "login-wechat.png"
		if dark {
			name = "login-wechat-dark.png"
		}
		f, err := os.Create(filepath.Join(dir, name))
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, tt.Image()); err != nil {
			t.Fatal(err)
		}
		f.Close()
	}
}
