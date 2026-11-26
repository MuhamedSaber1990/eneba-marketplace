// Package catalog reads products, offers and filter options from the database.
// It knows SQL and nothing about HTTP; the handlers in httpapi call it.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when no product has the requested slug. Handlers
// check for it with errors.Is and answer 404.
var ErrNotFound = errors.New("product not found")

// PageSize is how many products one page of the store shows.
const PageSize = 20

// Store runs the catalog queries.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// ProductSummary is a product as shown on a card in a list.
type ProductSummary struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Platform    string `json:"platform"`
	Region      string `json:"region"`
	Genre       string `json:"genre"`
	ReleaseDate string `json:"release_date"` // YYYY-MM-DD
	// FromPriceCents is the cheapest offer that has keys in stock. It is a
	// pointer so it can be nil, which becomes null in JSON: the product is
	// sold out.
	FromPriceCents *int   `json:"from_price_cents"`
	OfferCount     int    `json:"offer_count"`
	CoverURL       string `json:"cover_url"`
}

// ListParams are the filters, sort order and page of a store listing.
type ListParams struct {
	Platforms     []string
	Regions       []string
	Query         string // search text; see search.go for how it is matched
	MinPriceCents *int   // nil means no lower limit
	MaxPriceCents *int   // nil means no upper limit
	Sort          string // a key of sortClauses, or SortRelevance
	Page          int    // starts at 1
}

// ListResult is one page of products plus what the pager needs.
type ListResult struct {
	Items      []ProductSummary `json:"items"`
	Total      int              `json:"total"`
	Page       int              `json:"page"`
	PageSize   int              `json:"page_size"`
	TotalPages int              `json:"total_pages"`
}

// DefaultSort is used when the request names no sort order and has no search.
const DefaultSort = "newest"

// SortRelevance orders search results by how well they match. It is the
// default when the request has a search text. Without one it means nothing,
// and the listing falls back to DefaultSort.
const SortRelevance = "relevance"

// sortClauses maps the sort names the API accepts to ORDER BY clauses. The
// clause is the only part of a query chosen by a request, and it is always
// one of these fixed strings, never text typed by a visitor. Each clause ends
// with slug so the order is stable between pages.
var sortClauses = map[string]string{
	"newest":     "release_date DESC, slug",
	"price_asc":  "from_price_cents ASC NULLS LAST, slug",
	"price_desc": "from_price_cents DESC NULLS LAST, slug",
	"name":       "name ASC, slug",
}

// ValidSort reports whether name is a sort order the API accepts.
func ValidSort(name string) bool {
	_, ok := sortClauses[name]
	return ok || name == SortRelevance
}

// PricedProducts is the start of every query that shows products with a
// price. It defines "priced": every product plus two computed columns.
//   - from_price_cents: the lowest price among its offers that still have an
//     available key, or NULL when none do
//   - offer_count: how many sellers offer it
//
// It is exported (capital P) because the homepage blocks in package home
// build on the same definition, so "the price of a product" means one thing
// everywhere.
const PricedProducts = `
WITH priced AS (
    SELECT p.id, p.slug, p.title, p.name, p.platform, p.region, p.genre, p.release_date,
           (SELECT MIN(o.price_cents)
              FROM offers o
             WHERE o.product_id = p.id
               AND EXISTS (SELECT 1 FROM keys k
                            WHERE k.offer_id = o.id AND k.status = 'available')
           ) AS from_price_cents,
           (SELECT COUNT(*) FROM offers o WHERE o.product_id = p.id) AS offer_count
      FROM products p
)`

// SummaryColumns are the columns of "priced" that make up a ProductSummary,
// in the order ScanSummary reads them.
const SummaryColumns = `slug, name, platform, region, genre,
       to_char(release_date, 'YYYY-MM-DD'), from_price_cents, offer_count`

// ScanSummary reads one row selected with SummaryColumns. Extra columns
// selected after them can be read by passing where to put them.
//
// The parameter is "anything with a Scan method", which both a single row and
// a row inside a loop have. That is an interface: it names the behavior
// needed and not a concrete type.
func ScanSummary(row interface{ Scan(dest ...any) error }, extra ...any) (ProductSummary, error) {
	var p ProductSummary
	dest := []any{&p.Slug, &p.Name, &p.Platform, &p.Region, &p.Genre, &p.ReleaseDate, &p.FromPriceCents, &p.OfferCount}
	if err := row.Scan(append(dest, extra...)...); err != nil {
		return ProductSummary{}, err
	}
	p.CoverURL = CoverURL(p.Slug)
	return p, nil
}

// listFilter always has the same six parameters. A filter that is not in use
// is passed as an empty array or NULL, and its line then lets every row
// through. The SQL text never changes, so there is nothing to assemble and
// nothing to inject into.
//
//	$1 platforms   $2 regions   $3 search words   $4 min price   $5 max price
//	$6 similarity threshold for the search
const listFilter = `
 WHERE (cardinality($1::text[]) = 0 OR platform = ANY($1::text[]))
   AND (cardinality($2::text[]) = 0 OR region = ANY($2::text[]))
   AND ($4::int IS NULL OR from_price_cents >= $4::int)
   AND ($5::int IS NULL OR from_price_cents <= $5::int)` + searchFilter

// List returns one page of products matching the filters.
func (s *Store) List(ctx context.Context, params ListParams) (ListResult, error) {
	if !ValidSort(params.Sort) {
		return ListResult{}, fmt.Errorf("unknown sort %q", params.Sort)
	}
	empty := ListResult{Items: []ProductSummary{}, Page: params.Page, PageSize: PageSize}

	words := SearchWords(params.Query)
	if strings.TrimSpace(params.Query) != "" && len(words) == 0 {
		// Something was typed, but nothing searchable ("%", "!!!"). That is a
		// search with no results, not a request for the whole catalog.
		return empty, nil
	}

	filterArgs := []any{
		emptyIfNil(params.Platforms),
		emptyIfNil(params.Regions),
		words,
		params.MinPriceCents,
		params.MaxPriceCents,
		similarityThreshold,
	}

	// First query: how many products match in total, for the pager.
	var total int
	err := s.db.QueryRow(ctx, PricedProducts+" SELECT COUNT(*) FROM priced"+listFilter, filterArgs...).Scan(&total)
	if err != nil {
		return ListResult{}, fmt.Errorf("counting products: %w", err)
	}

	// Second query: the rows of the requested page.
	items, err := s.queryPage(ctx, params.Sort, filterArgs, words, PageSize, (params.Page-1)*PageSize)
	if err != nil {
		return ListResult{}, err
	}

	result := empty
	result.Items = items
	result.Total = total
	result.TotalPages = (total + PageSize - 1) / PageSize // rounds up
	return result, nil
}

// Suggest returns the best few matches for a search text, for the dropdown
// under the search box. Text shorter than MinSuggestLength gets no
// suggestions.
func (s *Store) Suggest(ctx context.Context, text string) ([]ProductSummary, error) {
	words := SearchWords(text)
	if len([]rune(strings.Join(words, ""))) < MinSuggestLength {
		return []ProductSummary{}, nil
	}
	// No platform, region or price filter: only the search words.
	filterArgs := []any{[]string{}, []string{}, words, nil, nil, similarityThreshold}
	return s.queryPage(ctx, SortRelevance, filterArgs, words, SuggestLimit, 0)
}

// queryPage runs the listing query for one page. List and Suggest share it.
func (s *Store) queryPage(ctx context.Context, sort string, filterArgs []any, words []string, limit, offset int) ([]ProductSummary, error) {
	args := append(filterArgs, limit, offset) // $7 and $8

	orderBy := sortClauses[sort]
	switch {
	case sort == SortRelevance && len(words) > 0:
		orderBy = relevanceOrder
		args = append(args, strings.Join(words, " ")) // $9, used only by relevanceOrder
	case sort == SortRelevance:
		orderBy = sortClauses[DefaultSort] // nothing to be relevant to
	}

	query := PricedProducts + `
		SELECT ` + SummaryColumns + `
		  FROM priced` + listFilter + `
		 ORDER BY ` + orderBy + `
		 LIMIT $7 OFFSET $8`
	rows, err := s.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("listing products: %w", err)
	}
	defer rows.Close()

	items := []ProductSummary{} // not nil, so an empty page is [] in JSON
	for rows.Next() {
		p, err := ScanSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a product row: %w", err)
		}
		items = append(items, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading product rows: %w", err)
	}
	return items, nil
}

// Offer is one seller's price for a product.
type Offer struct {
	ID               int64   `json:"id"`
	SellerName       string  `json:"seller_name"`
	SellerRating     float64 `json:"seller_rating"`
	SellerSalesCount int     `json:"seller_sales_count"`
	PriceCents       int     `json:"price_cents"`
	Stock            int     `json:"stock"` // available keys
}

// ProductDetail is everything the product page shows.
type ProductDetail struct {
	ProductSummary
	Description string  `json:"description"`
	Offers      []Offer `json:"offers"`
}

// Get returns one product with its offers. Offers that have stock come first,
// cheapest first, and sold-out offers go to the end.
func (s *Store) Get(ctx context.Context, slug string) (ProductDetail, error) {
	var detail ProductDetail
	var productID int64
	err := s.db.QueryRow(ctx, `
		SELECT id, slug, name, platform, region, genre,
		       to_char(release_date, 'YYYY-MM-DD'), description
		  FROM products
		 WHERE slug = $1`, slug,
	).Scan(&productID, &detail.Slug, &detail.Name, &detail.Platform, &detail.Region,
		&detail.Genre, &detail.ReleaseDate, &detail.Description)
	if errors.Is(err, pgx.ErrNoRows) {
		return ProductDetail{}, ErrNotFound
	}
	if err != nil {
		return ProductDetail{}, fmt.Errorf("loading product %q: %w", slug, err)
	}
	detail.CoverURL = CoverURL(detail.Slug)

	rows, err := s.db.Query(ctx, `
		SELECT id, seller_name, seller_rating, seller_sales_count, price_cents, stock
		  FROM (SELECT o.id, s.name AS seller_name, s.rating::float8 AS seller_rating,
		               s.sales_count AS seller_sales_count, o.price_cents,
		               (SELECT COUNT(*) FROM keys k
		                 WHERE k.offer_id = o.id AND k.status = 'available') AS stock
		          FROM offers o
		          JOIN sellers s ON s.id = o.seller_id
		         WHERE o.product_id = $1) AS product_offers
		 ORDER BY (stock = 0), price_cents, id`, productID)
	if err != nil {
		return ProductDetail{}, fmt.Errorf("loading offers of %q: %w", slug, err)
	}
	defer rows.Close()

	detail.Offers = []Offer{}
	for rows.Next() {
		var o Offer
		err := rows.Scan(&o.ID, &o.SellerName, &o.SellerRating, &o.SellerSalesCount, &o.PriceCents, &o.Stock)
		if err != nil {
			return ProductDetail{}, fmt.Errorf("reading an offer row: %w", err)
		}
		detail.Offers = append(detail.Offers, o)

		// The first in-stock offer is the cheapest one, because of the ORDER BY.
		if o.Stock > 0 && detail.FromPriceCents == nil {
			price := o.PriceCents
			detail.FromPriceCents = &price
		}
	}
	if err := rows.Err(); err != nil {
		return ProductDetail{}, fmt.Errorf("reading offer rows: %w", err)
	}
	detail.OfferCount = len(detail.Offers)
	return detail, nil
}

// Filters are the choices the store's filter panel offers.
type Filters struct {
	Platforms     []string `json:"platforms"`
	Regions       []string `json:"regions"`
	MinPriceCents int      `json:"min_price_cents"`
	MaxPriceCents int      `json:"max_price_cents"`
}

// Filters reads the filter choices from the data, so a new platform shows up
// in the panel as soon as a product uses it.
func (s *Store) Filters(ctx context.Context) (Filters, error) {
	var f Filters
	err := s.db.QueryRow(ctx, `
		SELECT ARRAY(SELECT DISTINCT platform FROM products ORDER BY platform),
		       ARRAY(SELECT DISTINCT region FROM products ORDER BY region),
		       COALESCE((SELECT MIN(price_cents) FROM offers), 0),
		       COALESCE((SELECT MAX(price_cents) FROM offers), 0)`,
	).Scan(&f.Platforms, &f.Regions, &f.MinPriceCents, &f.MaxPriceCents)
	if err != nil {
		return Filters{}, fmt.Errorf("loading filters: %w", err)
	}
	return f, nil
}

// Cover is what is needed to draw a product's cover image.
type Cover struct {
	Name     string // the full product name, used as the image's description
	Title    string // the game's title, drawn on the cover
	Platform string
	ColorA   string
	ColorB   string
}

func (s *Store) Cover(ctx context.Context, slug string) (Cover, error) {
	var c Cover
	err := s.db.QueryRow(ctx,
		"SELECT name, title, platform, cover_color_a, cover_color_b FROM products WHERE slug = $1", slug,
	).Scan(&c.Name, &c.Title, &c.Platform, &c.ColorA, &c.ColorB)
	if errors.Is(err, pgx.ErrNoRows) {
		return Cover{}, ErrNotFound
	}
	if err != nil {
		return Cover{}, fmt.Errorf("loading cover of %q: %w", slug, err)
	}
	return c, nil
}

// CoverURL is the address of a product's generated cover image.
func CoverURL(slug string) string {
	return "/api/products/" + slug + "/cover.svg"
}

// emptyIfNil makes sure a missing list reaches Postgres as an empty array.
// A nil slice would be sent as NULL, and cardinality(NULL) = 0 is not true,
// so the filter line would wrongly reject every row.
func emptyIfNil(values []string) []string {
	if values == nil {
		return []string{}
	}
	return values
}
