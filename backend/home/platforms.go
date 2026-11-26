package home

import (
	"context"
	"fmt"
)

// Platforms block. Source: the products table, counted per platform.

// Platform is one tile: a platform and how many products it has.
type Platform struct {
	Name         string `json:"name"`
	ProductCount int    `json:"product_count"`
}

// Platforms returns every platform that has products, the largest first.
func (s *Store) Platforms(ctx context.Context) ([]Platform, error) {
	rows, err := s.db.Query(ctx, `
		SELECT platform, COUNT(*)
		  FROM products
		 GROUP BY platform
		 ORDER BY COUNT(*) DESC, platform`)
	if err != nil {
		return nil, fmt.Errorf("loading platforms: %w", err)
	}
	defer rows.Close()

	platforms := []Platform{}
	for rows.Next() {
		var p Platform
		if err := rows.Scan(&p.Name, &p.ProductCount); err != nil {
			return nil, fmt.Errorf("reading a platform row: %w", err)
		}
		platforms = append(platforms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading platform rows: %w", err)
	}
	return platforms, nil
}
