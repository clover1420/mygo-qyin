package app

import "github.com/egoist/mygo/ui"

// The palette's scales, as Cloudy draws them: one neutral grey, from the
// page's white (or near black) up to its text, and one red accent, from the
// faintest wash of it to the red of a heart taken.
//
// Steps keep the roles the app reads them by: 1–2 backgrounds, 3–5 controls,
// 6–8 borders, 9–10 solid accents, 11–12 text. Where Cloudy draws a wash in
// the accent it uses the accent at a low alpha, so the steps below the solid
// one are translucent: they lie over whatever is under them, as Cloudy's do.

// neutralLight and neutralDark are the greys Cloudy's surfaces and text
// resolve to, composited: its white page, its #f4f4f6 sidebar, its hairlines
// and its muted greys.
var (
	neutralLight = [...]string{
		"#ffffff", "#f4f4f6", "#f2f2f4", "#ededf0", "#e7e7ea", "#e3e3e6",
		"#dcdce0", "#cdcdd2", "#b4b4ba", "#9a9aa0", "#6e6e73", "#1d1d1f",
	}
	neutralDark = [...]string{
		"#141416", "#1c1c1f", "#232326", "#292930", "#2f2f35", "#35353b",
		"#45454c", "#55555c", "#6c6c74", "#8f8f96", "#9a9aa2", "#f5f5f7",
	}

	// The gold of a membership badge, which neither palette owns.
	warmLight = [...]string{
		"#fefdfb", "#fefbe9", "#fff7c2", "#ffee9c", "#fbe577", "#f3d673",
		"#e9c162", "#e2a336", "#ffc53d", "#ffba18", "#ab6400", "#4f3422",
	}
	warmDark = [...]string{
		"#16120c", "#1d180f", "#302008", "#3f2700", "#4d3000", "#5c3d05",
		"#714f19", "#8f6424", "#ffc53d", "#ffd60a", "#ffca16", "#ffe7b3",
	}
)

// neutral is a grey of the page: its background, its controls, its
// hairlines, its text.
func neutral(dark bool, step int) ui.Color {
	if dark {
		return ui.Hex(neutralDark[step-1])
	}
	return ui.Hex(neutralLight[step-1])
}

// accent is the app's red, washed at the lower steps and solid at the ninth,
// where a button takes it whole.
func accent(dark bool, step int) ui.Color {
	base := accentBase(dark)
	switch step {
	case 1:
		return base.Alpha(0.05)
	case 2:
		return base.Alpha(0.08)
	case 3:
		return base.Alpha(0.12)
	case 4:
		return base.Alpha(0.17)
	case 5:
		return base.Alpha(0.24)
	case 6:
		return base.Alpha(0.4)
	case 7:
		return base.Alpha(0.6)
	case 8:
		return base.Alpha(0.8)
	case 9:
		return base
	case 10:
		if dark {
			return base.Mix(ui.Hex("#ffffff"), 0.08)
		}
		return base.Mix(ui.Hex("#000000"), 0.12)
	case 11:
		// Text: deeper on white, lighter on black, as a link reads.
		if dark {
			return base.Mix(ui.Hex("#ffffff"), 0.32)
		}
		return base.Mix(ui.Hex("#000000"), 0.18)
	default:
		if dark {
			return base.Mix(ui.Hex("#ffffff"), 0.46)
		}
		return base.Mix(ui.Hex("#000000"), 0.3)
	}
}

// accentBase is the app's red itself, the one Cloudy's buttons and hearts take.
func accentBase(dark bool) ui.Color {
	if dark {
		return ui.Oklch(0.66, 0.2, 22)
	}
	return ui.Oklch(0.6, 0.21, 24)
}

// warm is the membership's gold.
func warm(dark bool, step int) ui.Color {
	if dark {
		return ui.Hex(warmDark[step-1])
	}
	return ui.Hex(warmLight[step-1])
}

// veil and shade are white and black at an alpha, for what lies over artwork.
var veilAlpha = [...]float32{0.05, 0.1, 0.15, 0.2, 0.3, 0.4, 0.5, 0.6, 0.7, 0.8, 0.9, 0.95}

func veil(step int) ui.Color { return paper.Alpha(veilAlpha[step-1]) }
func shade(step int) ui.Color { return ink.Alpha(veilAlpha[step-1]) }

// Black and white are the opaque endpoints of the overlay scales; they also
// give the fixed contrast of media controls and QR codes.
var (
	ink   = ui.Hex("#000000")
	paper = ui.Hex("#ffffff")
)

// The colors of the confetti a like throws, around the heart's red.
var (
	burstOrange = ui.Hex("#f76b15")
	burstPurple = ui.Hex("#8e4ec6")
	burstCyan   = ui.Hex("#00a2c7")
	burstPink   = ui.Hex("#d6409f")
)
