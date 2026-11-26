package catalog

import (
	"fmt"
	"html"
	"strings"
)

// Cover size in pixels. The store card is 225 wide in the design reference.
const (
	coverWidth  = 225
	coverHeight = 300
)

// RenderSVG draws the cover: a diagonal gradient, the title in large letters
// and the platform on a dark strip along the bottom.
//
// The product is invented, so there is no artwork to use. Drawing the cover
// from data keeps every product looking finished without any image files.
func (c Cover) RenderSVG() string {
	// The product name ends with "<platform> Key <REGION>". Only the title
	// part goes on the cover, the platform has its own strip.
	title, _, _ := strings.Cut(c.Name, " "+c.Platform+" Key ")
	lines := wrapWords(title, 11)

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" role="img" aria-label="%s">`,
		coverWidth, coverHeight, coverWidth, coverHeight, html.EscapeString(c.Name))
	fmt.Fprintf(&b, `<defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="%s"/><stop offset="1" stop-color="%s"/></linearGradient></defs>`,
		html.EscapeString(c.ColorA), html.EscapeString(c.ColorB))
	fmt.Fprintf(&b, `<rect width="%d" height="%d" fill="url(#g)"/>`, coverWidth, coverHeight)

	// Two large translucent circles, so the cover is not a flat gradient.
	b.WriteString(`<circle cx="190" cy="50" r="90" fill="#ffffff" fill-opacity="0.12"/>`)
	b.WriteString(`<circle cx="30" cy="250" r="110" fill="#000000" fill-opacity="0.15"/>`)

	// The title, one <text> element per line, centered as a block.
	const lineHeight = 34
	firstLineY := 140 - (len(lines)-1)*lineHeight/2
	for i, line := range lines {
		fmt.Fprintf(&b, `<text x="20" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="30" font-weight="800" fill="#ffffff">%s</text>`,
			firstLineY+i*lineHeight, html.EscapeString(strings.ToUpper(line)))
	}

	// The platform strip.
	fmt.Fprintf(&b, `<rect y="%d" width="%d" height="34" fill="#000000" fill-opacity="0.55"/>`, coverHeight-34, coverWidth)
	fmt.Fprintf(&b, `<text x="20" y="%d" font-family="Arial, Helvetica, sans-serif" font-size="13" font-weight="800" fill="#ffffff">%s</text>`,
		coverHeight-12, html.EscapeString(strings.ToUpper(c.Platform)))

	b.WriteString(`</svg>`)
	return b.String()
}

// wrapWords splits text into lines of at most maxLen characters, breaking
// only between words. A single word longer than maxLen gets its own line.
func wrapWords(text string, maxLen int) []string {
	lines := []string{}
	current := ""
	for _, word := range strings.Fields(text) {
		switch {
		case current == "":
			current = word
		case len(current)+1+len(word) <= maxLen:
			current += " " + word
		default:
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}
