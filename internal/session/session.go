// Package session keeps where playback was left, in session.json in the user
// data directory: on quitting, and every little while as it plays, in case
// the app does not get to quit.
package session

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/egoist/mygo"

	"lilt/internal/player"
)

// Path returns where the session is kept.
func Path() (string, error) {
	dir, err := mygo.App.Path(mygo.PathUserData)
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// Load reads the session last kept, its raw JSON, and whether there was one.
func Load() (player.Session, string, bool) {
	path, err := Path()
	if err != nil {
		return player.Session{}, "", false
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return player.Session{}, "", false
	}
	var s player.Session
	if json.Unmarshal(b, &s) != nil {
		return player.Session{}, "", false
	}
	return s, string(b), true
}

// Encode returns the session as the JSON kept on disk.
func Encode(s player.Session) (string, error) {
	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// Write keeps the session's raw JSON atomically.
func Write(raw string) error {
	path, err := Path()
	if err != nil {
		return err
	}
	if err := os.WriteFile(path+".part", []byte(raw), 0o600); err != nil {
		return err
	}
	return os.Rename(path+".part", path)
}
