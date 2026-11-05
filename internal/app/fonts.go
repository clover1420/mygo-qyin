package app

import (
	"log"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/egoist/mygo"
	_ "github.com/egoist/mygo/locales/zh-Hans"
	"github.com/egoist/mygo/ui"
)

// appLanguage is the language the app is written in. Text is in it, and
// Windows picks the fonts for Han characters by it: in English, as MyGo
// takes an app without catalogs to be, they came from Japanese fonts, which
// lack many simplified characters, so that a line changed font and weight
// from one character to the next.
const appLanguage = "zh-Hans"

// A uiFont is a font the listener can read the app in.
type uiFont struct {
	id    string // kept in settings; empty for the default
	label string
	// family is the list of families text takes, the first one there is
	// drawing it.
	family string
	// files are those of the font in the fonts directory of the app's
	// resources, for a font the system doesn't have.
	files []string
}

// uiFonts are the fonts on offer, the default first. Those in the system
// are the ones Windows comes with in Chinese.
var uiFonts = []uiFont{
	{id: "", label: "默认"},
	{id: "yahei", label: "微软雅黑", family: "Microsoft YaHei UI"},
	{id: "dengxian", label: "等线", family: "DengXian"},
	{id: "simhei", label: "黑体", family: "SimHei"},
	{id: "noto-serif", label: "思源宋体", family: "Noto Serif SC"},
	{id: "simsun", label: "宋体", family: "SimSun"},
	{id: "wenkai", label: "霞鹜文楷", family: "LXGW WenKai", files: []string{"LXGWWenKai-Regular.ttf", "LXGWWenKai-Medium.ttf"}},
	{id: "kaiti", label: "楷体", family: "KaiTi"},
	{id: "fangsong", label: "仿宋", family: "FangSong"},
}

// defaultFontFamily is the family of the default font. Elsewhere the
// system's reads Chinese well; on Windows Segoe UI has no Han characters.
// Noto Sans SC has Medium and Black as well, where Microsoft YaHei
// has Light, Regular and Bold alone.
func defaultFontFamily() string {
	if runtime.GOOS == "windows" {
		return "Noto Sans SC, Microsoft YaHei UI"
	}
	return ""
}

// fontOf returns the preset of an id, and whether it is one: the listener
// may choose any family this computer has, kept by its name.
func fontOf(id string) (uiFont, bool) {
	for _, f := range uiFonts {
		if f.id == id {
			return f, true
		}
	}
	return uiFont{}, false
}

// fontFamily returns the family list of the font chosen, for the theme.
func (a *app) fontFamily() string {
	id := a.settings.Font
	if id == "" {
		return defaultFontFamily()
	}
	if f, ok := fontOf(id); ok {
		return f.family
	}
	// A family of the system, kept by its name.
	return id
}

// loadFont registers the files of the font chosen, the first time it is
// chosen: their tens of MB stay on disk for those who read another font.
// A font that won't load leaves its family to the system's fallback.
func (a *app) loadFont() {
	f, ok := fontOf(a.settings.Font)
	if !ok || len(f.files) == 0 || a.fontsLoaded[f.id] {
		return
	}
	dir, err := mygo.App.Path(mygo.PathResources)
	if err != nil {
		log.Printf("font %s: %v", f.id, err)
		return
	}
	for _, name := range f.files {
		b, err := os.ReadFile(filepath.Join(dir, "fonts", name))
		if err == nil {
			err = ui.RegisterFont(b, "")
		}
		if err != nil {
			log.Printf("font %s: %v", f.id, err)
			return
		}
	}
	if a.fontsLoaded == nil {
		a.fontsLoaded = map[string]bool{}
	}
	a.fontsLoaded[f.id] = true
}

// fontChoices returns the fonts this computer can show, read once: the
// presets it has, the ones that come as files, then every other family the
// system lists, so that what is chosen is always a font that is there.
func (a *app) fontChoices() []uiFont {
	if a.fontList != nil {
		return a.fontList
	}
	// The system's families, by name, one of each name however it is cased.
	families := make(map[string]string, len(uiFonts))
	for _, f := range ui.FontFamilies() {
		name := strings.TrimSpace(f.Name)
		if name == "" {
			continue
		}
		key := strings.ToLower(name)
		if _, ok := families[key]; !ok {
			families[key] = name
		}
	}
	dir, _ := mygo.App.Path(mygo.PathResources)
	choices := make([]uiFont, 0, len(families)+1)
	choices = append(choices, uiFonts[0]) // the default first
	used := map[string]bool{"": true}
	for _, f := range uiFonts[1:] {
		key := strings.ToLower(f.family)
		name, installed := families[key]
		shipped := len(f.files) > 0 && fileExists(filepath.Join(dir, "fonts", f.files[0]))
		if !installed && !shipped {
			continue // not on this computer, and not shipped with the app
		}
		choices = append(choices, f)
		used[key] = true
		if installed {
			used[strings.ToLower(name)] = true
		}
	}
	rest := make([]string, 0, len(families))
	for key, name := range families {
		if !used[key] {
			rest = append(rest, name)
		}
	}
	slices.SortFunc(rest, func(x, y string) int { return strings.Compare(strings.ToLower(x), strings.ToLower(y)) })
	for _, name := range rest {
		choices = append(choices, uiFont{id: name, label: name, family: name})
	}
	a.fontList = choices
	return choices
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
