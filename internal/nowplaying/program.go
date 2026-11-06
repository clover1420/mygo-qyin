package nowplaying

import (
	"strings"

	"github.com/egoist/mygo/ui"

	"lilt/internal/qqmusic"
)

// ProgramPane is what the episode playing is about, in Now Playing, in
// place of a song's lyrics.
func ProgramPane(c *ui.Context, tr *qqmusic.Track) {
	ui.Scroll(c.Key(tr.ProgramID)).Grow(1).Padding(4, 12, 48, 0).Children(func() {
		desc := strings.TrimSpace(tr.Description)
		if desc == "" {
			desc = "这期节目没有简介"
		}
		ui.Text(c, desc).FontSize(15).LineHeight(1.7).TextColor(white.Alpha(0.75))
	})
}
