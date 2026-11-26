package home

import (
	"context"
	"fmt"

	"eneba-marketplace/backend/catalog"
)

// New releases block. Source: the products table, by release date.

// NewReleases returns the most recently released games, each game once.
//
// A product row is one edition (game, platform, region), so the newest rows
// would be the same game several times. DISTINCT ON (title) keeps one row per
// title: the first one in the ORDER BY, which is the cheapest edition that
// has stock (NULLS LAST pushes sold-out editions behind the others).
func (s *Store) NewReleases(ctx context.Context) ([]catalog.ProductSummary, error) {
	rows, err := s.db.Query(ctx, catalog.PricedProducts+`,
		one_per_title AS (
		    SELECT DISTINCT ON (title) *
		      FROM priced
		     ORDER BY title, from_price_cents ASC NULLS LAST, slug
		)
		SELECT `+catalog.SummaryColumns+`
		  FROM one_per_title
		 ORDER BY release_date DESC, slug
		 LIMIT $1`, productRowSize)
	if err != nil {
		return nil, fmt.Errorf("loading new releases: %w", err)
	}
	defer rows.Close()

	items := []catalog.ProductSummary{}
	for rows.Next() {
		item, err := catalog.ScanSummary(rows)
		if err != nil {
			return nil, fmt.Errorf("reading a new release row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading new release rows: %w", err)
	}
	return items, nil
}
