// Package home serves the homepage: the list of blocks on the page and the
// content of each block. Every block has its own file, its own query and its
// own data source, and knows nothing about the others.
package home

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Store runs the homepage queries.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// blockEndpoints lists the block types the backend can serve and where each
// one's content lives. A row in home_layout with any other type is ignored.
var blockEndpoints = map[string]string{
	"banner":       "/api/home/banners",
	"deals":        "/api/home/deals",
	"best_sellers": "/api/home/best-sellers",
	"new_releases": "/api/home/new-releases",
	"platforms":    "/api/home/platforms",
}

// LayoutRow is one row of the home_layout table.
type LayoutRow struct {
	BlockType string
	Title     string
	SeeAllURL string
	Enabled   bool
}

// Block is one entry of the homepage as the frontend receives it: what kind
// of block to draw, its heading, and the endpoint to fetch its content from.
type Block struct {
	Type      string `json:"type"`
	Title     string `json:"title"`
	SeeAllURL string `json:"see_all_url"`
	Endpoint  string `json:"endpoint"`
}

// BuildLayout turns layout rows (already in page order) into the blocks to
// show. Disabled rows and unknown block types are left out. It touches no
// database, so its rules can be tested directly.
func BuildLayout(rows []LayoutRow) []Block {
	blocks := []Block{}
	for _, row := range rows {
		endpoint, known := blockEndpoints[row.BlockType]
		if !row.Enabled || !known {
			continue
		}
		blocks = append(blocks, Block{
			Type:      row.BlockType,
			Title:     row.Title,
			SeeAllURL: row.SeeAllURL,
			Endpoint:  endpoint,
		})
	}
	return blocks
}

// Layout reads home_layout and returns the blocks of the homepage, top to
// bottom.
func (s *Store) Layout(ctx context.Context) ([]Block, error) {
	rows, err := s.db.Query(ctx,
		"SELECT block_type, title, see_all_url, enabled FROM home_layout ORDER BY position")
	if err != nil {
		return nil, fmt.Errorf("loading the home layout: %w", err)
	}
	defer rows.Close()

	layoutRows := []LayoutRow{}
	for rows.Next() {
		var row LayoutRow
		if err := rows.Scan(&row.BlockType, &row.Title, &row.SeeAllURL, &row.Enabled); err != nil {
			return nil, fmt.Errorf("reading a layout row: %w", err)
		}
		layoutRows = append(layoutRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading layout rows: %w", err)
	}
	return BuildLayout(layoutRows), nil
}
