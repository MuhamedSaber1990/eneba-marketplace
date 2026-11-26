package catalog

import (
	"slices"
	"strings"
	"testing"
)

func TestWrapWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "short text stays on one line", text: "Ashen Crown", want: []string{"Ashen Crown"}},
		{name: "breaks between words", text: "Thunder Circuit Kart", want: []string{"Thunder", "Circuit", "Kart"}},
		{name: "fills a line when words fit", text: "Neon Drift 2089", want: []string{"Neon Drift", "2089"}},
		{name: "a long word gets its own line", text: "Witchlight Academy", want: []string{"Witchlight", "Academy"}},
		{name: "empty text gives no lines", text: "", want: []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := wrapWords(tt.text, 11)
			if !slices.Equal(got, tt.want) {
				t.Errorf("wrapWords(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

func TestRenderSVG(t *testing.T) {
	cover := Cover{Name: "Ashen Crown Steam Key GLOBAL", Title: "Ashen Crown", Platform: "Steam", ColorA: "#ff5f6d", ColorB: "#ffc371"}
	svg := cover.RenderSVG()

	for _, want := range []string{"<svg ", "</svg>", "ASHEN CROWN", "STEAM", "#ff5f6d", "#ffc371"} {
		if !strings.Contains(svg, want) {
			t.Errorf("SVG does not contain %q", want)
		}
	}
	// The full name is the image's description, but only the title is drawn.
	if !strings.Contains(svg, `aria-label="Ashen Crown Steam Key GLOBAL"`) {
		t.Errorf("the image has no description with the full name")
	}
	if strings.Contains(svg, "GLOBAL</text>") {
		t.Errorf("the region was drawn on the cover")
	}
}

// Text from the database must not be able to add markup to the image.
func TestRenderSVGEscapesText(t *testing.T) {
	cover := Cover{Name: `<script>alert(1)</script> Steam Key GLOBAL`, Title: `<script>alert(1)</script>`, Platform: "Steam", ColorA: `"><script>`, ColorB: "#000"}
	svg := cover.RenderSVG()

	if strings.Contains(svg, "<script>") || strings.Contains(svg, "<SCRIPT>") {
		t.Errorf("unescaped markup in SVG: %s", svg)
	}
}

func TestEscapeLike(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"plain text is unchanged", "crown", "crown"},
		{"percent is escaped", "100%", `100\%`},
		{"underscore is escaped", "a_b", `a\_b`},
		{"backslash is escaped", `a\b`, `a\\b`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := escapeLike(tt.in); got != tt.want {
				t.Errorf("escapeLike(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestTitleFontSize(t *testing.T) {
	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{"short lines get the full size", []string{"Ashen", "Crown"}, 30},
		{"eight letters still fit at full size", []string{"Frontier"}, 30},
		{"a ten-letter word is drawn smaller", []string{"Witchlight", "Academy"}, 25},
		{"the longest line decides", []string{"Kart", "Thunderstorm"}, 21},
		{"very long words stop shrinking at 14", []string{"Supercalifragilisticexpialidocious"}, 14},
		{"no lines does not divide by zero", []string{}, 30},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := titleFontSize(tt.lines); got != tt.want {
				t.Errorf("titleFontSize(%q) = %d, want %d", tt.lines, got, tt.want)
			}
		})
	}
}
