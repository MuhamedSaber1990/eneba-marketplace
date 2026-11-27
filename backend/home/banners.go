package home

import (
	"context"
	"fmt"
)

// Banner block. Source: the banners table.

// Banner is one promotion panel: text and a button on a picture, or on a
// two-color gradient when it has no picture.
type Banner struct {
	Title    string `json:"title"`
	Subtitle string `json:"subtitle"`
	CTALabel string `json:"cta_label"` // CTA = call to action, the button text
	Link     string `json:"link"`
	ColorA   string `json:"color_a"`
	ColorB   string `json:"color_b"`
	// ImageURL is a pointer so it can be nil, which becomes null in JSON: the
	// banner has no picture and is drawn as the gradient alone.
	ImageURL *string `json:"image_url"`
}

// maxBanners is how many panels the banner layout has room for.
const maxBanners = 3

// Banners returns the active banners in display order. The first one is the
// large panel.
func (s *Store) Banners(ctx context.Context) ([]Banner, error) {
	rows, err := s.db.Query(ctx, `
		SELECT title, subtitle, cta_label, link, color_a, color_b, image_url
		  FROM banners
		 WHERE active
		 ORDER BY position, id
		 LIMIT $1`, maxBanners)
	if err != nil {
		return nil, fmt.Errorf("loading banners: %w", err)
	}
	defer rows.Close()

	banners := []Banner{}
	for rows.Next() {
		var b Banner
		if err := rows.Scan(&b.Title, &b.Subtitle, &b.CTALabel, &b.Link, &b.ColorA, &b.ColorB, &b.ImageURL); err != nil {
			return nil, fmt.Errorf("reading a banner row: %w", err)
		}
		banners = append(banners, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading banner rows: %w", err)
	}
	return banners, nil
}
