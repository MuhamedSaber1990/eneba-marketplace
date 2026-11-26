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

// BestSellers returns the products with the most units sold. Sold-out
// products are left out: a best seller nobody can buy is not useful on a
// shop's front page.
func (s *Store) BestSellers(ctx context.Context) ([]BestSeller, error) {
	rows, err := s.db.Query(ctx, catalog.PricedProducts+`
		SELECT `+catalog.SummaryColumns+`, sales.units_sold
		  FROM priced
		  JOIN product_sales sales ON sales.product_id = priced.id
		 WHERE priced.from_price_cents IS NOT NULL
		 ORDER BY sales.units_sold DESC, priced.slug
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
