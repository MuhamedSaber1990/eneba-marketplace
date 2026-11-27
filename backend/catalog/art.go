package catalog

import "strings"

// Cover art. A game title can have an illustration, stored as a static file
// in frontend/public/covers and served by the site itself, not by this API.
// All editions of a title (Steam, Xbox, each region) share one illustration.
//
// The illustrations are AI-generated. docs/cover-art.md says with which tool
// and gives the prompt for each.

// artTitles lists the titles that have an illustration, by file name without
// the extension. A title missing from this list gets the drawn SVG cover.
//
// TestArtTitlesMatchTheFiles checks this list against the folder in both
// directions, so a file cannot be forgotten and a name cannot be misspelled.
var artTitles = map[string]bool{
	"ashen-crown":          true,
	"bladeborn-arena":      true,
	"circuit-breakers":     true,
	"dune-runners":         true,
	"emberwatch":           true,
	"frostline-rally":      true,
	"gearheart-garage":     true,
	"hollow-tide":          true,
	"ironvale-tactics":     true,
	"lanternwood":          true,
	"mythforge-legends":    true,
	"neon-drift-2089":      true,
	"obsidian-vow":         true,
	"paper-planets":        true,
	"pocket-kingdoms":      true,
	"quiet-harbor":         true,
	"rogue-meridian":       true,
	"saltwind-pirates":     true,
	"skybound-couriers":    true,
	"starfall-odyssey":     true,
	"thunder-circuit-kart": true,
	"tiny-titans-league":   true,
	"verdant-protocol":     true,
	"witchlight-academy":   true,
}

// artVersion is part of every illustration's address. The files are cached
// by browsers for a year. Replacing a file means raising this number, which
// gives every illustration a new address.
const artVersion = "3"

// CoverFor returns the address of a product's cover and whether it is an
// illustration.
//
// The second value matters to the page: an illustration has no text in it, so
// the page writes the title over it. The SVG fallback already has the title
// drawn in, and must not get it a second time.
func CoverFor(title, productSlug string) (url string, isArt bool) {
	file := artFileName(title)
	if artTitles[file] {
		return "/covers/" + file + ".webp?v=" + artVersion, true
	}
	return CoverURL(productSlug), false
}

// artFileName turns a title into its file name: "Neon Drift 2089" becomes
// "neon-drift-2089".
func artFileName(title string) string {
	return strings.ReplaceAll(strings.ToLower(strings.TrimSpace(title)), " ", "-")
}
