package catalog

import (
	"context"
	"errors"
	"os"
	"slices"
	"strings"
	"testing"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

// These tests run real queries against the dev database (DATABASE_URL in
// .env), filled by cmd/seed. They check properties that hold for any seed
// data ("every result is a Steam product", "prices never go down") rather
// than exact names. Without a database they are skipped.

func testStore(t *testing.T) *Store {
	t.Helper()
	config.LoadDotEnv("../../.env") // tests run in this package's folder
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL is not set, skipping database tests")
	}
	pool, err := db.Connect(context.Background())
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewStore(pool)
}

// listAll walks every page of a listing and returns all its items.
func listAll(t *testing.T, store *Store, params ListParams) []ProductSummary {
	t.Helper()
	all := []ProductSummary{}
	for page := 1; ; page++ {
		params.Page = page
		result, err := store.List(context.Background(), params)
		if err != nil {
			t.Fatalf("List page %d: %v", page, err)
		}
		all = append(all, result.Items...)
		if page >= result.TotalPages {
			if len(all) != result.Total {
				t.Fatalf("got %d items over all pages, but total says %d", len(all), result.Total)
			}
			return all
		}
	}
}

func intPtr(n int) *int { return &n }

func TestListPagination(t *testing.T) {
	store := testStore(t)
	all := listAll(t, store, ListParams{Sort: "name"})

	if len(all) <= PageSize {
		t.Fatalf("need more than one page of seed data, got %d products", len(all))
	}
	// No product may appear on two pages.
	seen := map[string]bool{}
	for _, p := range all {
		if seen[p.Slug] {
			t.Errorf("%s appears on more than one page", p.Slug)
		}
		seen[p.Slug] = true
	}

	// A page past the end is empty but still reports the total.
	result, err := store.List(context.Background(), ListParams{Sort: "name", Page: 999})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(result.Items) != 0 || result.Total != len(all) {
		t.Errorf("page 999: %d items, total %d; want 0 items, total %d", len(result.Items), result.Total, len(all))
	}
}

func TestListFilters(t *testing.T) {
	store := testStore(t)
	everything := listAll(t, store, ListParams{Sort: "name"})

	tests := []struct {
		name   string
		params ListParams
		keep   func(p ProductSummary) bool // what a correct result looks like
	}{
		{
			name:   "one platform",
			params: ListParams{Platforms: []string{"Steam"}},
			keep:   func(p ProductSummary) bool { return p.Platform == "Steam" },
		},
		{
			name:   "two platforms",
			params: ListParams{Platforms: []string{"Steam", "Xbox"}},
			keep:   func(p ProductSummary) bool { return p.Platform == "Steam" || p.Platform == "Xbox" },
		},
		{
			name:   "platform and region together",
			params: ListParams{Platforms: []string{"PlayStation"}, Regions: []string{"Europe"}},
			keep:   func(p ProductSummary) bool { return p.Platform == "PlayStation" && p.Region == "Europe" },
		},
		{
			name:   "price range leaves out sold-out products",
			params: ListParams{MinPriceCents: intPtr(2000), MaxPriceCents: intPtr(4000)},
			keep: func(p ProductSummary) bool {
				return p.FromPriceCents != nil && *p.FromPriceCents >= 2000 && *p.FromPriceCents <= 4000
			},
		},
		{
			name:   "name search ignores case",
			params: ListParams{Query: "CROWN"},
			keep:   func(p ProductSummary) bool { return strings.Contains(strings.ToLower(p.Name), "crown") },
		},
		{
			name:   "a percent sign is not a wildcard",
			params: ListParams{Query: "%"},
			keep:   func(p ProductSummary) bool { return strings.Contains(p.Name, "%") },
		},
		{
			name:   "unknown platform matches nothing",
			params: ListParams{Platforms: []string{"Dreamcast"}},
			keep:   func(p ProductSummary) bool { return false },
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.params.Sort = "name"
			got := listAll(t, store, tt.params)

			// The expected result: filter the full list in Go with the same rule.
			want := []string{}
			for _, p := range everything {
				if tt.keep(p) {
					want = append(want, p.Slug)
				}
			}
			gotSlugs := []string{}
			for _, p := range got {
				gotSlugs = append(gotSlugs, p.Slug)
			}
			slices.Sort(want)
			slices.Sort(gotSlugs)
			if !slices.Equal(gotSlugs, want) {
				t.Errorf("got %d products %v\nwant %d products %v", len(gotSlugs), gotSlugs, len(want), want)
			}
		})
	}
}

func TestListSorting(t *testing.T) {
	store := testStore(t)

	// inOrder says whether b may come right after a.
	tests := []struct {
		sort    string
		inOrder func(a, b ProductSummary) bool
	}{
		{"name", func(a, b ProductSummary) bool { return a.Name <= b.Name }},
		{"newest", func(a, b ProductSummary) bool { return a.ReleaseDate >= b.ReleaseDate }},
		{"price_asc", func(a, b ProductSummary) bool {
			if a.FromPriceCents == nil { // sold out: only sold-out products may follow
				return b.FromPriceCents == nil
			}
			return b.FromPriceCents == nil || *a.FromPriceCents <= *b.FromPriceCents
		}},
		{"price_desc", func(a, b ProductSummary) bool {
			if a.FromPriceCents == nil {
				return b.FromPriceCents == nil
			}
			return b.FromPriceCents == nil || *a.FromPriceCents >= *b.FromPriceCents
		}},
	}

	for _, tt := range tests {
		t.Run(tt.sort, func(t *testing.T) {
			items := listAll(t, store, ListParams{Sort: tt.sort})
			for i := 1; i < len(items); i++ {
				if !tt.inOrder(items[i-1], items[i]) {
					t.Fatalf("out of order at position %d: %q then %q", i, items[i-1].Name, items[i].Name)
				}
			}
		})
	}

	t.Run("unknown sort is an error", func(t *testing.T) {
		if _, err := store.List(context.Background(), ListParams{Sort: "release_date; DROP TABLE products", Page: 1}); err == nil {
			t.Fatal("expected an error")
		}
	})
}

func TestGet(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	t.Run("unknown slug is ErrNotFound", func(t *testing.T) {
		_, err := store.Get(ctx, "no-such-product")
		if !errors.Is(err, ErrNotFound) {
			t.Fatalf("error = %v, want ErrNotFound", err)
		}
	})

	t.Run("offers are in stock first, then by price, and agree with the list", func(t *testing.T) {
		for _, summary := range listAll(t, store, ListParams{Sort: "name"}) {
			detail, err := store.Get(ctx, summary.Slug)
			if err != nil {
				t.Fatalf("Get(%s): %v", summary.Slug, err)
			}
			if len(detail.Offers) != summary.OfferCount {
				t.Errorf("%s: %d offers, list says %d", summary.Slug, len(detail.Offers), summary.OfferCount)
			}
			for i := 1; i < len(detail.Offers); i++ {
				before, after := detail.Offers[i-1], detail.Offers[i]
				bothInStock := before.Stock > 0 && after.Stock > 0
				if before.Stock == 0 && after.Stock > 0 {
					t.Errorf("%s: a sold-out offer comes before one in stock", summary.Slug)
				}
				if bothInStock && before.PriceCents > after.PriceCents {
					t.Errorf("%s: in-stock offers are not cheapest first", summary.Slug)
				}
			}
			// The "from" price on the card must be the price the page leads with.
			if (detail.FromPriceCents == nil) != (summary.FromPriceCents == nil) ||
				(detail.FromPriceCents != nil && *detail.FromPriceCents != *summary.FromPriceCents) {
				t.Errorf("%s: list and detail disagree on the from price", summary.Slug)
			}
		}
	})
}

func TestFilters(t *testing.T) {
	store := testStore(t)
	filters, err := store.Filters(context.Background())
	if err != nil {
		t.Fatalf("Filters: %v", err)
	}
	if len(filters.Platforms) == 0 || len(filters.Regions) == 0 {
		t.Errorf("no platforms or regions: %+v", filters)
	}
	if !slices.IsSorted(filters.Platforms) || !slices.IsSorted(filters.Regions) {
		t.Errorf("filter options are not sorted: %+v", filters)
	}
	if filters.MinPriceCents <= 0 || filters.MaxPriceCents < filters.MinPriceCents {
		t.Errorf("bad price range: %d to %d", filters.MinPriceCents, filters.MaxPriceCents)
	}
}
