package theme

import "github.com/egoist/mygo/ui"

// Palette holds the app's colors, for the light and the dark appearance.
type Palette struct {
	Dark      bool
	Bg        ui.Color // the content's background
	Sidebar   ui.Color // under the sidebar where no material shows
	Text      ui.Color
	Muted     ui.Color
	Faint     ui.Color
	Hairline  ui.Color
	Hover     ui.Color
	Press     ui.Color
	Accent    ui.Color
	OnAccent  ui.Color
	CardShade ui.Color
}

// For returns the palette of a theme.
func For(t *ui.Theme) Palette {
	if t.Dark {
		return Palette{
			Dark:      true,
			Bg:        ui.Hex("#141416"),
			Sidebar:   ui.Hex("#1c1c1f"),
			Text:      ui.Hex("#f5f5f7"),
			Muted:     ui.RGBA(235, 235, 245, 0.56),
			Faint:     ui.RGBA(235, 235, 245, 0.3),
			Hairline:  ui.RGBA(255, 255, 255, 0.07),
			Hover:     ui.RGBA(255, 255, 255, 0.06),
			Press:     ui.RGBA(255, 255, 255, 0.1),
			Accent:    AccentBase(true),
			OnAccent:  Paper,
			CardShade: ui.RGBA(0, 0, 0, 0.45),
		}
	}
	return Palette{
		Bg:        ui.Hex("#ffffff"),
		Sidebar:   ui.Hex("#f4f4f6"),
		Text:      ui.Hex("#1d1d1f"),
		Muted:     ui.RGBA(60, 60, 67, 0.62),
		Faint:     ui.RGBA(60, 60, 67, 0.32),
		Hairline:  ui.RGBA(0, 0, 0, 0.07),
		Hover:     ui.RGBA(0, 0, 0, 0.04),
		Press:     ui.RGBA(0, 0, 0, 0.08),
		Accent:    AccentBase(false),
		OnAccent:  Paper,
		CardShade: ui.RGBA(0, 0, 0, 0.16),
	}
}
