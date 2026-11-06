//go:build !windows

package lyricsui

import "github.com/egoist/mygo"

func placeTaskbarLyricsCover(w *mygo.Window, p Placement) error { return nil }

func startTaskbarLyricsMonitor(deliver func(Placement)) func()   { return func() {} }
func placeTaskbarLyricsWindow(w *mygo.Window, p Placement) error { return nil }
func detachTaskbarLyricsWindow(w *mygo.Window)                   {}
