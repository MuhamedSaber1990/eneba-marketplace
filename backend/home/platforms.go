package home

import (
	"context"
	"fmt"

	"eneba-marketplace/backend/catalog"
)

// Platforms block. Source: the products table, counted per platform.

// Platform is one tile: a platform, how many products it has, and a few of
// their covers to show what is behind the tile.
type Platform struct {
	Name         string   `json:"name"`
	ProductCount int      `json:"product_count"`
	Covers       []string `json:"covers"` // addresses of up to tileCovers illustrations
}

// tileCovers is how many covers one tile shows.
const tileCovers = 3

// Platforms returns every platform that has products, the largest first.
func (s *Store) Platforms(ctx context.Context) ([]Platform, error) {
	rows, err := s.db.Query(ctx, `
		SELECT p.platform, COUNT(*),
		       -- A few game titles sold on this platform. Ordering by a hash
		       -- of title and platform gives each platform its own fixed
		       -- pick. Ordering by date would put the same newest games on
		       -- every tile.
		       ARRAY(SELECT t.title
		               FROM (SELECT DISTINCT p2.title
		                       FROM products p2
		                      WHERE p2.platform = p.platform) t
		              ORDER BY md5(t.title || p.platform)
		              LIMIT $1)
		  FROM products p
		 GROUP BY p.platform
		 ORDER BY COUNT(*) DESC, p.platform`, tileCovers)
	if err != nil {
		return nil, fmt.Errorf("loading platforms: %w", err)
	}
	defer rows.Close()

	platforms := []Platform{}
	for rows.Next() {
		var p Platform
		var titles []string
		if err := rows.Scan(&p.Name, &p.ProductCount, &titles); err != nil {
			return nil, fmt.Errorf("reading a platform row: %w", err)
		}
		// Only illustrations go on a tile. A title without one is left out:
		// its drawn cover carries a title of its own, too small to read here.
		p.Covers = []string{}
		for _, title := range titles {
			if url, isArt := catalog.CoverFor(title, ""); isArt {
				p.Covers = append(p.Covers, url)
			}
		}
		platforms = append(platforms, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading platform rows: %w", err)
	}
	return platforms, nil
}
