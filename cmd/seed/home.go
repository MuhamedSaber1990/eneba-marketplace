package main

import (
	"context"
	"fmt"
	"math/rand/v2"
	"time"

	"github.com/jackc/pgx/v5"
)

// How many deals the seed creates, and how far apart their end dates are.
// The last deal ends about 60 days after seeding, so the deals block does
// not empty out all at once.
const (
	dealCount         = 8
	firstDealEndsDays = 5
	dealSpacingDays   = 8
)

// seedHome inserts the homepage content: the layout, the banners, an invented
// sales history for the best sellers block, and the deals.
//
// The sales figures and the deals are made up. Nothing was sold and no price
// was ever higher: they exist so the blocks have something to show.
func seedHome(ctx context.Context, tx pgx.Tx) error {
	// A different fixed seed from the catalog's, so changing one section's
	// data does not shift the other's random numbers.
	rng := rand.New(rand.NewPCG(2026, 3))

	for position, block := range homeLayout {
		_, err := tx.Exec(ctx,
			"INSERT INTO home_layout (position, block_type, title, see_all_url) VALUES ($1, $2, $3, $4)",
			position+1, block.blockType, block.title, block.seeAllURL)
		if err != nil {
			return fmt.Errorf("inserting layout block %q: %w", block.blockType, err)
		}
	}

	for position, b := range banners {
		_, err := tx.Exec(ctx, `
			INSERT INTO banners (title, subtitle, cta_label, link, color_a, color_b, position)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`,
			b.title, b.subtitle, b.ctaLabel, b.link, b.colorA, b.colorB, position+1)
		if err != nil {
			return fmt.Errorf("inserting banner %q: %w", b.title, err)
		}
	}

	salesRows, err := seedSales(ctx, tx, rng)
	if err != nil {
		return err
	}
	deals, err := seedDeals(ctx, tx, rng)
	if err != nil {
		return err
	}

	fmt.Printf("homepage: seeded %d layout blocks, %d banners, %d sales rows, %d deals\n",
		len(homeLayout), len(banners), salesRows, deals)
	return nil
}

// seedSales gives every product an invented number of units sold. Most
// products sell a little and about one in five sells a lot, so the best
// sellers block has clear leaders.
func seedSales(ctx context.Context, tx pgx.Tx, rng *rand.Rand) (int, error) {
	productIDs, err := queryIDs(ctx, tx, "SELECT id FROM products ORDER BY id")
	if err != nil {
		return 0, fmt.Errorf("listing products: %w", err)
	}

	units := make([]int32, len(productIDs))
	for i := range productIDs {
		units[i] = int32(20 + rng.IntN(300))
		if rng.IntN(5) == 0 {
			units[i] += int32(1000 + rng.IntN(4000))
		}
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO product_sales (product_id, units_sold)
		SELECT * FROM unnest($1::bigint[], $2::int[])`,
		productIDs, units)
	if err != nil {
		return 0, fmt.Errorf("inserting sales figures: %w", err)
	}
	return len(productIDs), nil
}

// seedDeals puts a deal on the cheapest in-stock offer of several different
// games. For each one it remembers the current price as the "was" price and
// then lowers the offer's real price, so the deal price is what the store,
// the product page and the checkout all see.
func seedDeals(ctx context.Context, tx pgx.Tx, rng *rand.Rand) (int, error) {
	// One candidate per game title: its cheapest offer that has stock.
	rows, err := tx.Query(ctx, `
		SELECT DISTINCT ON (p.title) o.id, o.price_cents
		  FROM offers o
		  JOIN products p ON p.id = o.product_id
		 WHERE EXISTS (SELECT 1 FROM keys k
		                WHERE k.offer_id = o.id AND k.status = 'available')
		 ORDER BY p.title, o.price_cents, o.id`)
	if err != nil {
		return 0, fmt.Errorf("listing deal candidates: %w", err)
	}
	type candidate struct {
		offerID    int64
		priceCents int
	}
	candidates, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (candidate, error) {
		var c candidate
		err := row.Scan(&c.offerID, &c.priceCents)
		return c, err
	})
	if err != nil {
		return 0, fmt.Errorf("reading deal candidates: %w", err)
	}

	count := min(dealCount, len(candidates))
	now := time.Now().UTC()

	for i, pick := range rng.Perm(len(candidates))[:count] {
		c := candidates[pick]

		discount := 15 + rng.IntN(26) // 15 to 40 percent
		newPrice := c.priceCents * (100 - discount) / 100
		newPrice = newPrice - newPrice%10 + 9 // end in 9, like the other prices
		if newPrice >= c.priceCents {
			continue // a very cheap offer where rounding ate the discount
		}

		// Deal 1 ends in 5 days, deal 2 in 13, and so on, plus a few hours
		// so the countdowns are not all identical.
		endsAt := now.AddDate(0, 0, firstDealEndsDays+i*dealSpacingDays).Add(time.Duration(rng.IntN(20)) * time.Hour)

		_, err := tx.Exec(ctx,
			"INSERT INTO deals (offer_id, original_price_cents, starts_at, ends_at) VALUES ($1, $2, $3, $4)",
			c.offerID, c.priceCents, now, endsAt)
		if err != nil {
			return 0, fmt.Errorf("inserting a deal: %w", err)
		}
		_, err = tx.Exec(ctx, "UPDATE offers SET price_cents = $1 WHERE id = $2", newPrice, c.offerID)
		if err != nil {
			return 0, fmt.Errorf("lowering an offer's price: %w", err)
		}
	}

	var inserted int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM deals").Scan(&inserted); err != nil {
		return 0, fmt.Errorf("counting deals: %w", err)
	}
	return inserted, nil
}

// queryIDs runs a query that returns one bigint column and collects the values.
func queryIDs(ctx context.Context, tx pgx.Tx, sql string) ([]int64, error) {
	rows, err := tx.Query(ctx, sql)
	if err != nil {
		return nil, err
	}
	return pgx.CollectRows(rows, pgx.RowTo[int64])
}
