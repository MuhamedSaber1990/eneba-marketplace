package catalog

import (
	"context"
	"slices"
	"strings"
	"testing"
)

func TestSearchWords(t *testing.T) {
	tests := []struct {
		name string
		text string
		want []string
	}{
		{"one word is lowercased", "Kart", []string{"kart"}},
		{"several words are split", "  Thunder   Circuit ", []string{"thunder", "circuit"}},
		{"punctuation separates words", "kart,xbox-key", []string{"kart", "xbox", "key"}},
		{"digits are kept", "Drift 2089", []string{"drift", "2089"}},
		{"wildcard characters are dropped", "100% _off_", []string{"100", "off"}},
		{"only symbols gives no words", "% _ !", []string{}},
		{"empty text gives no words", "", []string{}},
		{"at most five words", "a b c d e f g", []string{"a", "b", "c", "d", "e"}},
		{"a very long word is cut to 40 letters", strings.Repeat("x", 60), []string{strings.Repeat("x", 40)}},
		{"letters with accents count as letters", "Pokémon", []string{"pokémon"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := SearchWords(tt.text)
			if !slices.Equal(got, tt.want) {
				t.Errorf("SearchWords(%q) = %q, want %q", tt.text, got, tt.want)
			}
		})
	}
}

// titleOfProduct returns the game's title from a product name: the part
// before its platform.
func titleOfProduct(p ProductSummary) string {
	title, _, _ := strings.Cut(p.Name, " "+p.Platform+" Key ")
	return title
}

// searchTitles runs a search over all pages and returns the distinct game
// titles found, sorted.
func searchTitles(t *testing.T, store *Store, query string) []string {
	t.Helper()
	titles := []string{}
	for _, p := range listAll(t, store, ListParams{Query: query, Sort: SortRelevance}) {
		if title := titleOfProduct(p); !slices.Contains(titles, title) {
			titles = append(titles, title)
		}
	}
	slices.Sort(titles)
	return titles
}

// The searches promised in the plan, against the seeded catalog. Each case
// lists exactly which games must be found: nothing missing, nothing extra.
func TestSearch(t *testing.T) {
	store := testStore(t)

	tests := []struct {
		group string
		query string
		want  []string // game titles, sorted
	}{
		// Partial words.
		{"partial", "thund", []string{"Thunder Circuit Kart"}},
		{"partial", "forge", []string{"Mythforge Legends"}},
		{"partial", "circ", []string{"Circuit Breakers", "Thunder Circuit Kart"}},
		{"partial", "harb", []string{"Quiet Harbor"}},
		{"partial", "kart", []string{"Thunder Circuit Kart"}},

		// Typos.
		{"typo", "witchlite", []string{"Witchlight Academy"}},
		{"typo", "circut", []string{"Circuit Breakers", "Thunder Circuit Kart"}},
		{"typo", "obsidain", []string{"Obsidian Vow"}},
		{"typo", "starfal odyssey", []string{"Starfall Odyssey"}},
		{"typo", "emberwach", []string{"Emberwatch"}},
		{"typo", "lanternwod", []string{"Lanternwood"}},
		{"typo", "mythforge legnds", []string{"Mythforge Legends"}},
		{"typo", "frostlin raly", []string{"Frostline Rally"}},

		// Several words, upper case.
		{"words", "QUIET HARBOR", []string{"Quiet Harbor"}},
		{"words", "kart xbox", []string{"Thunder Circuit Kart"}},
		{"words", "ashen steam", []string{}}, // Ashen Crown has no Steam edition in the seed

		// Must find nothing.
		{"nothing", "zzzzqq", []string{}},
		{"nothing", "%", []string{}},
		{"nothing", "thunder zzzz", []string{}},

		// Known limits: these do NOT work, and the test pins that down so a
		// change in behavior is noticed.
		{"limit", "vwo", []string{}},    // typo in a three-letter word (vow)
		{"limit", "krat", []string{}},   // two letters swapped in a four-letter word (kart)
		{"limit", "rouge", []string{}},  // swapped letters in a five-letter word (rogue)
		{"limit", "glbal", []string{}},  // scores 0.44 against "GLOBAL", just under the threshold
		{"limit", "racing", []string{}}, // a genre, not part of any name
	}

	for _, tt := range tests {
		t.Run(tt.group+"/"+tt.query, func(t *testing.T) {
			got := searchTitles(t, store, tt.query)
			if !slices.Equal(got, tt.want) {
				t.Errorf("search %q found %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestSearchByPlatformWord(t *testing.T) {
	store := testStore(t)

	// "switch" is part of the platform in every Nintendo Switch product name.
	found := listAll(t, store, ListParams{Query: "switch", Sort: SortRelevance})
	onSwitch := listAll(t, store, ListParams{Platforms: []string{"Nintendo Switch"}, Sort: "name"})
	if len(found) != len(onSwitch) {
		t.Fatalf(`"switch" found %d products, there are %d Nintendo Switch products`, len(found), len(onSwitch))
	}
	for _, p := range found {
		if p.Platform != "Nintendo Switch" {
			t.Errorf(`"switch" found %s, which is on %s`, p.Slug, p.Platform)
		}
	}

	// "kart xbox" must give only Xbox editions.
	for _, p := range listAll(t, store, ListParams{Query: "kart xbox", Sort: SortRelevance}) {
		if p.Platform != "Xbox" {
			t.Errorf(`"kart xbox" found %s, which is on %s`, p.Slug, p.Platform)
		}
	}
}

func TestSearchRanking(t *testing.T) {
	store := testStore(t)

	// "circuit" starts the name of Circuit Breakers and sits in the middle
	// of Thunder Circuit Kart. Names that start with the text come first.
	items := listAll(t, store, ListParams{Query: "circuit", Sort: SortRelevance})
	seenOther := false
	for _, p := range items {
		startsWith := strings.HasPrefix(strings.ToLower(p.Name), "circuit")
		if !startsWith {
			seenOther = true
		}
		if startsWith && seenOther {
			t.Errorf("%s starts with the search text but is ranked below a product that only contains it", p.Slug)
		}
	}
	if len(items) == 0 || !seenOther {
		t.Fatalf("expected both kinds of match for %q, got %d products", "circuit", len(items))
	}
}

func TestSearchWithFilters(t *testing.T) {
	store := testStore(t)

	all := listAll(t, store, ListParams{Query: "circut", Sort: SortRelevance})
	onXbox := listAll(t, store, ListParams{Query: "circut", Platforms: []string{"Xbox"}, Sort: SortRelevance})

	want := 0
	for _, p := range all {
		if p.Platform == "Xbox" {
			want++
		}
	}
	if len(onXbox) != want {
		t.Errorf("search with a platform filter gave %d products, want %d", len(onXbox), want)
	}
	for _, p := range onXbox {
		if p.Platform != "Xbox" {
			t.Errorf("%s is on %s", p.Slug, p.Platform)
		}
	}

	// Any other sort order still works together with a search.
	byPrice := listAll(t, store, ListParams{Query: "circut", Sort: "price_asc"})
	if len(byPrice) != len(all) {
		t.Errorf("sorting by price changed the number of results: %d, want %d", len(byPrice), len(all))
	}
}

func TestSuggest(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	t.Run("too short gives nothing", func(t *testing.T) {
		for _, text := range []string{"", "a", " a ", "%"} {
			items, err := store.Suggest(ctx, text)
			if err != nil {
				t.Fatalf("Suggest(%q): %v", text, err)
			}
			if len(items) != 0 {
				t.Errorf("Suggest(%q) gave %d items, want 0", text, len(items))
			}
		}
	})

	t.Run("a typo still suggests the game", func(t *testing.T) {
		items, err := store.Suggest(ctx, "thundr")
		if err != nil {
			t.Fatalf("Suggest: %v", err)
		}
		if len(items) == 0 || titleOfProduct(items[0]) != "Thunder Circuit Kart" {
			t.Errorf(`Suggest("thundr") = %d items, first %+v`, len(items), items)
		}
	})

	t.Run("never more than the limit", func(t *testing.T) {
		items, err := store.Suggest(ctx, "key") // "key" is in every product name
		if err != nil {
			t.Fatalf("Suggest: %v", err)
		}
		if len(items) != SuggestLimit {
			t.Errorf("got %d suggestions, want %d", len(items), SuggestLimit)
		}
	})
}
