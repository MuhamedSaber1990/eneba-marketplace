package home

import (
	"context"
	"fmt"
	"time"

	"eneba-marketplace/backend/catalog"
)

// Deals block. Source: the deals table, joined to the offer each deal is on.

// Deal is one discounted offer.
type Deal struct {
	Slug               string    `json:"slug"`
	Name               string    `json:"name"`
	Platform           string    `json:"platform"`
	Region             string    `json:"region"`
	CoverURL           string    `json:"cover_url"`
	SellerName         string    `json:"seller_name"`
	PriceCents         int       `json:"price_cents"`          // what a buyer pays now
	OriginalPriceCents int       `json:"original_price_cents"` // the price before the deal
	DiscountPercent    int       `json:"discount_percent"`
	EndsAt             time.Time `json:"ends_at"` // written as a timestamp like 2026-12-24T18:00:00Z
}

// maxDeals is how many cards the deals row shows.
const maxDeals = 6

// Deals returns the deals that can be bought right now, the ones ending
// soonest first. A deal is shown only when all of these hold:
//   - the current time is inside its window (started, not ended)
//   - the offer still has a key in stock
//   - the offer's price really is below the old price
func (s *Store) Deals(ctx context.Context) ([]Deal, error) {
	rows, err := s.db.Query(ctx, `
		SELECT p.slug, p.name, p.platform, p.region, s.name,
		       o.price_cents, d.original_price_cents, d.ends_at
		  FROM deals d
		  JOIN offers o   ON o.id = d.offer_id
		  JOIN products p ON p.id = o.product_id
		  JOIN sellers s  ON s.id = o.seller_id
		 WHERE d.starts_at <= now()
		   AND d.ends_at > now()
		   AND o.price_cents < d.original_price_cents
		   AND EXISTS (SELECT 1 FROM keys k
		                WHERE k.offer_id = o.id AND k.status = 'available')
		 ORDER BY d.ends_at, d.id
		 LIMIT $1`, maxDeals)
	if err != nil {
		return nil, fmt.Errorf("loading deals: %w", err)
	}
	defer rows.Close()

	deals := []Deal{}
	for rows.Next() {
		var d Deal
		err := rows.Scan(&d.Slug, &d.Name, &d.Platform, &d.Region, &d.SellerName,
			&d.PriceCents, &d.OriginalPriceCents, &d.EndsAt)
		if err != nil {
			return nil, fmt.Errorf("reading a deal row: %w", err)
		}
		d.EndsAt = d.EndsAt.UTC() // always written as ...Z, whatever the server's time zone
		d.CoverURL = catalog.CoverURL(d.Slug)
		d.DiscountPercent = DiscountPercent(d.OriginalPriceCents, d.PriceCents)
		deals = append(deals, d)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading deal rows: %w", err)
	}
	return deals, nil
}

// DiscountPercent says how far the price dropped, as a whole percent rounded
// to the nearest: from 4999 to 3499 is 30. It returns 0 when there is no drop.
func DiscountPercent(originalCents, priceCents int) int {
	if originalCents <= 0 || priceCents >= originalCents {
		return 0
	}
	saved := originalCents - priceCents
	// Adding half of the divisor before dividing rounds to nearest using
	// whole numbers only, so no float is involved in a money calculation.
	return (saved*100 + originalCents/2) / originalCents
}
