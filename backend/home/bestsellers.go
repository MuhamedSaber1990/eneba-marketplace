package home

import (
	"context"
	"fmt"

	"eneba-marketplace/backend/catalog"
)

// Best sellers block. Source: the product_sales table (units sold per product).

// BestSeller is a product together with how many units of it were sold.
// The ProductSummary is embedded, so its fields appear directly in the JSON.
type BestSeller struct {
	catalog.ProductSummary
	UnitsSold int `json:"units_sold"`
}

// productRowSize is how many cards a product row on the homepage shows.
const productRowSize = 6

// BestSellers returns the games with the most units sold, each game once.
//
// Sales are counted per product, and a product is one edition (game,
// platform, region). Ranking the rows directly could put two editions of one
// game in the row. So the query first keeps one edition per title, the one
// that sold the most (DISTINCT ON keeps the first row of each title in the
// ORDER BY), and then ranks those.
//
// Sold-out editions are left out before that choice is made: a best seller
// nobody can buy is not useful on a shop's front page.
func (s *Store) BestSellers(ctx context.Context) ([]BestSeller, error) {
	rows, err := s.db.Query(ctx, catalog.PricedProducts+`,
		best_edition_per_title AS (
		    SELECT DISTINCT ON (priced.title) priced.*, sales.units_sold
		      FROM priced
		      JOIN product_sales sales ON sales.product_id = priced.id
		     WHERE priced.from_price_cents IS NOT NULL
		     ORDER BY priced.title, sales.units_sold DESC, priced.slug
		)
		SELECT `+catalog.SummaryColumns+`, units_sold
		  FROM best_edition_per_title
		 ORDER BY units_sold DESC, slug
		 LIMIT $1`, productRowSize)
	if err != nil {
		return nil, fmt.Errorf("loading best sellers: %w", err)
	}
	defer rows.Close()

	items := []BestSeller{}
	for rows.Next() {
		var item BestSeller
		item.ProductSummary, err = catalog.ScanSummary(rows, &item.UnitsSold)
		if err != nil {
			return nil, fmt.Errorf("reading a best seller row: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading best seller rows: %w", err)
	}
	return items, nil
}
