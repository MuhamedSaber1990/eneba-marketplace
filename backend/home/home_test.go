package home

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"eneba-marketplace/backend/catalog"
	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

// These tests run each block's query against the seeded dev database and
// check the rule the block promises. Without a database they are skipped.

func testStores(t *testing.T) (*Store, *catalog.Store) {
	t.Helper()
	config.LoadDotEnv("../../.env")
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL is not set, skipping database tests")
	}
	pool, err := db.Connect(context.Background())
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}
	t.Cleanup(pool.Close)
	return NewStore(pool), catalog.NewStore(pool)
}

func TestLayoutFromDatabase(t *testing.T) {
	store, _ := testStores(t)
	blocks, err := store.Layout(context.Background())
	if err != nil {
		t.Fatalf("Layout: %v", err)
	}
	if len(blocks) == 0 {
		t.Fatal("the seeded layout has no blocks")
	}
	for _, block := range blocks {
		if !strings.HasPrefix(block.Endpoint, "/api/home/") {
			t.Errorf("block %q has endpoint %q", block.Type, block.Endpoint)
		}
	}
}

func TestBanners(t *testing.T) {
	store, _ := testStores(t)
	banners, err := store.Banners(context.Background())
	if err != nil {
		t.Fatalf("Banners: %v", err)
	}
	if len(banners) == 0 || len(banners) > maxBanners {
		t.Fatalf("got %d banners, want 1 to %d", len(banners), maxBanners)
	}
	for _, b := range banners {
		if b.Title == "" || b.CTALabel == "" || !strings.HasPrefix(b.Link, "/") {
			t.Errorf("incomplete banner: %+v", b)
		}
	}
}

func TestDeals(t *testing.T) {
	store, catalogStore := testStores(t)
	ctx := context.Background()

	deals, err := store.Deals(ctx)
	if err != nil {
		t.Fatalf("Deals: %v", err)
	}
	if len(deals) == 0 {
		t.Skip("no deal is active right now (the seeded ones have ended); run cmd/seed -reset to renew them")
	}
	if len(deals) > maxDeals {
		t.Errorf("got %d deals, want at most %d", len(deals), maxDeals)
	}

	now := time.Now()
	for i, deal := range deals {
		if deal.PriceCents >= deal.OriginalPriceCents {
			t.Errorf("%s: price %d is not below the old price %d", deal.Slug, deal.PriceCents, deal.OriginalPriceCents)
		}
		if deal.DiscountPercent < 1 || deal.DiscountPercent > 99 {
			t.Errorf("%s: discount of %d%%", deal.Slug, deal.DiscountPercent)
		}
		if !deal.EndsAt.After(now) {
			t.Errorf("%s: ended at %v but is still listed", deal.Slug, deal.EndsAt)
		}
		if i > 0 && deal.EndsAt.Before(deals[i-1].EndsAt) {
			t.Errorf("%s: deals are not ordered by end time", deal.Slug)
		}

		// The deal's offer must be buyable: the product page has to show an
		// in-stock offer from that seller at that price.
		product, err := catalogStore.Get(ctx, deal.Slug)
		if err != nil {
			t.Fatalf("Get(%s): %v", deal.Slug, err)
		}
		found := false
		for _, offer := range product.Offers {
			if offer.SellerName == deal.SellerName && offer.PriceCents == deal.PriceCents && offer.Stock > 0 {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no in-stock offer from %s at %d on the product page", deal.Slug, deal.SellerName, deal.PriceCents)
		}
	}
}

// titleOf returns the game's title from a product name: the part before its
// platform. "Quiet Harbor Xbox Key GLOBAL" on Xbox gives "Quiet Harbor".
func titleOf(p catalog.ProductSummary) string {
	title, _, _ := strings.Cut(p.Name, " "+p.Platform+" Key ")
	return title
}

func TestBestSellers(t *testing.T) {
	store, _ := testStores(t)
	items, err := store.BestSellers(context.Background())
	if err != nil {
		t.Fatalf("BestSellers: %v", err)
	}
	if len(items) == 0 || len(items) > productRowSize {
		t.Fatalf("got %d best sellers, want 1 to %d", len(items), productRowSize)
	}

	seenTitles := map[string]bool{}
	for i, item := range items {
		if item.FromPriceCents == nil {
			t.Errorf("%s is sold out but listed as a best seller", item.Slug)
		}
		if i > 0 && item.UnitsSold > items[i-1].UnitsSold {
			t.Errorf("%s sold %d, more than the one ranked above it (%d)", item.Slug, item.UnitsSold, items[i-1].UnitsSold)
		}
		// Each game once, even when two of its editions both sold well.
		title := titleOf(item.ProductSummary)
		if seenTitles[title] {
			t.Errorf("%q appears more than once", title)
		}
		seenTitles[title] = true
	}
}

// The edition shown for a game must be the one that sold the most. This reads
// the sales table directly and compares.
func TestBestSellersPicksTheBestSellingEdition(t *testing.T) {
	store, _ := testStores(t)
	ctx := context.Background()

	items, err := store.BestSellers(ctx)
	if err != nil {
		t.Fatalf("BestSellers: %v", err)
	}
	for _, item := range items {
		// The highest sales figure among this game's editions that have stock.
		var best int
		err := store.db.QueryRow(ctx, `
			SELECT MAX(sales.units_sold)
			  FROM products p
			  JOIN product_sales sales ON sales.product_id = p.id
			 WHERE p.title = (SELECT title FROM products WHERE slug = $1)
			   AND EXISTS (SELECT 1 FROM offers o JOIN keys k ON k.offer_id = o.id
			                WHERE o.product_id = p.id AND k.status = 'available')`,
			item.Slug).Scan(&best)
		if err != nil {
			t.Fatalf("reading sales for %s: %v", item.Slug, err)
		}
		if item.UnitsSold != best {
			t.Errorf("%s is shown with %d units sold, but another edition of the game sold %d", item.Slug, item.UnitsSold, best)
		}
	}
}

func TestNewReleases(t *testing.T) {
	store, catalogStore := testStores(t)
	ctx := context.Background()

	items, err := store.NewReleases(ctx)
	if err != nil {
		t.Fatalf("NewReleases: %v", err)
	}
	if len(items) == 0 || len(items) > productRowSize {
		t.Fatalf("got %d new releases, want 1 to %d", len(items), productRowSize)
	}

	// Each game once: no two cards may share a title.
	seenTitles := map[string]bool{}
	for i, item := range items {
		title := titleOf(item)
		if seenTitles[title] {
			t.Errorf("%q appears more than once", title)
		}
		seenTitles[title] = true

		if i > 0 && item.ReleaseDate > items[i-1].ReleaseDate {
			t.Errorf("%s (%s) comes after an older release (%s)", item.Slug, item.ReleaseDate, items[i-1].ReleaseDate)
		}
	}

	// The first card must be an edition of the newest game in the whole catalog.
	newest, err := catalogStore.List(ctx, catalog.ListParams{Sort: "newest", Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if items[0].ReleaseDate != newest.Items[0].ReleaseDate {
		t.Errorf("first new release is from %s, but the newest product is from %s", items[0].ReleaseDate, newest.Items[0].ReleaseDate)
	}
}

func TestPlatforms(t *testing.T) {
	store, catalogStore := testStores(t)
	ctx := context.Background()

	platforms, err := store.Platforms(ctx)
	if err != nil {
		t.Fatalf("Platforms: %v", err)
	}

	sum := 0
	for i, p := range platforms {
		sum += p.ProductCount
		if i > 0 && p.ProductCount > platforms[i-1].ProductCount {
			t.Errorf("%s is not in falling order of product count", p.Name)
		}
		// The tile's number must match what the store shows for that filter.
		listed, err := catalogStore.List(ctx, catalog.ListParams{Platforms: []string{p.Name}, Sort: "name", Page: 1})
		if err != nil {
			t.Fatalf("List(%s): %v", p.Name, err)
		}
		if listed.Total != p.ProductCount {
			t.Errorf("%s: tile says %d products, the store lists %d", p.Name, p.ProductCount, listed.Total)
		}
	}

	all, err := catalogStore.List(ctx, catalog.ListParams{Sort: "name", Page: 1})
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if sum != all.Total {
		t.Errorf("platform counts add up to %d, but there are %d products", sum, all.Total)
	}
}
