// Command seed fills an empty database with the invented catalog: products,
// sellers, offers and keys.
//
// It uses a fixed random seed, so every database it fills gets exactly the
// same data. That keeps the dev and production databases comparable and lets
// tests rely on what is there.
//
// Usage, from the repository root:
//
//	go run ./cmd/seed                       (uses .env, the dev database)
//	go run ./cmd/seed -env .env.production  (the production database)
//	go run ./cmd/seed -reset                (delete the catalog first, then fill it)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"math/rand/v2"
	"strings"

	"github.com/jackc/pgx/v5"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

func main() {
	envFile := flag.String("env", ".env", "file to read DATABASE_URL_UNPOOLED from")
	reset := flag.Bool("reset", false, "delete the existing catalog before seeding")
	flag.Parse()

	config.LoadDotEnv(*envFile)
	ctx := context.Background()

	conn, label, err := db.ConnectDirect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	fmt.Printf("database: %s (from %s)\n", label, *envFile)

	if err := seed(ctx, conn, *reset); err != nil {
		log.Fatal(err)
	}
}

func seed(ctx context.Context, conn *pgx.Conn, reset bool) error {
	// Everything happens in one transaction: the catalog is either fully
	// there or not there at all.
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if reset {
		// CASCADE also empties tables that point at these ones.
		_, err := tx.Exec(ctx, "TRUNCATE keys, offers, sellers, products RESTART IDENTITY CASCADE")
		if err != nil {
			return fmt.Errorf("emptying the catalog: %w", err)
		}
	}

	var existing int
	if err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&existing); err != nil {
		return fmt.Errorf("counting products: %w", err)
	}
	if existing > 0 {
		return fmt.Errorf("the database already has %d products; use -reset to replace them", existing)
	}

	// The same two numbers always produce the same sequence of "random" values.
	rng := rand.New(rand.NewPCG(2026, 1126))

	sellerIDs, err := insertSellers(ctx, tx)
	if err != nil {
		return err
	}

	// Keys are collected here and inserted in one statement at the end.
	var keyOfferIDs []int64
	var keyCodes []string
	products, offers := 0, 0

	for gameIndex, g := range games {
		colors := coverColors[gameIndex%len(coverColors)]

		// Each game is sold in two or three editions, picked at random.
		editionCount := 2 + rng.IntN(2)
		for _, editionIndex := range rng.Perm(len(editions))[:editionCount] {
			e := editions[editionIndex]
			name := fmt.Sprintf("%s %s Key %s", g.title, e.platform, strings.ToUpper(e.region))

			var productID int64
			err := tx.QueryRow(ctx, `
				INSERT INTO products (slug, name, description, platform, region, genre,
				                      release_date, cover_color_a, cover_color_b)
				VALUES ($1, $2, $3, $4, $5, $6, $7::date, $8, $9)
				RETURNING id`,
				slugify(name), name, g.description, e.platform, e.region, g.genre,
				g.released, colors[0], colors[1],
			).Scan(&productID)
			if err != nil {
				return fmt.Errorf("inserting product %q: %w", name, err)
			}
			products++

			// A few products have no keys at all, so "Sold out" can be seen.
			soldOut := products%20 == 7

			// Two to five sellers offer each product.
			offerCount := 2 + rng.IntN(4)
			for _, sellerIndex := range rng.Perm(len(sellerIDs))[:offerCount] {
				price := sellerPrice(rng, g.basePrice)

				var offerID int64
				err := tx.QueryRow(ctx, `
					INSERT INTO offers (product_id, seller_id, price_cents)
					VALUES ($1, $2, $3)
					RETURNING id`,
					productID, sellerIDs[sellerIndex], price,
				).Scan(&offerID)
				if err != nil {
					return fmt.Errorf("inserting offer for %q: %w", name, err)
				}
				offers++

				// Stock: usually 1 to 15 keys, sometimes none.
				keyCount := 1 + rng.IntN(15)
				if soldOut || rng.IntN(100) < 12 {
					keyCount = 0
				}
				for range keyCount {
					keyOfferIDs = append(keyOfferIDs, offerID)
					keyCodes = append(keyCodes, randomKey(rng))
				}
			}
		}
	}

	// unnest turns the two arrays into rows, pairing them up by position, so
	// well over a thousand keys go in with a single round trip.
	_, err = tx.Exec(ctx, `
		INSERT INTO keys (offer_id, code)
		SELECT * FROM unnest($1::bigint[], $2::text[])`,
		keyOfferIDs, keyCodes,
	)
	if err != nil {
		return fmt.Errorf("inserting keys: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	fmt.Printf("seeded %d products, %d sellers, %d offers, %d keys\n",
		products, len(sellerIDs), offers, len(keyCodes))
	return nil
}

func insertSellers(ctx context.Context, tx pgx.Tx) ([]int64, error) {
	ids := make([]int64, 0, len(sellers))
	for _, s := range sellers {
		var id int64
		err := tx.QueryRow(ctx,
			"INSERT INTO sellers (name, rating, sales_count) VALUES ($1, $2, $3) RETURNING id",
			s.name, s.rating, s.sales,
		).Scan(&id)
		if err != nil {
			return nil, fmt.Errorf("inserting seller %q: %w", s.name, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// sellerPrice moves the base price up or down by as much as 15% and makes it
// end in 9, like shop prices do (for example 4199 cents, shown as 41.99).
func sellerPrice(rng *rand.Rand, basePrice int) int {
	percent := 85 + rng.IntN(31) // 85 to 115
	price := basePrice * percent / 100
	return price - price%10 + 9
}

// randomKey makes a code shaped like XXXXX-XXXXX-XXXXX. Letters and digits
// that look alike (0 and O, 1 and I) are left out.
func randomKey(rng *rand.Rand) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"
	var b strings.Builder
	for i := range 15 {
		if i > 0 && i%5 == 0 {
			b.WriteByte('-')
		}
		b.WriteByte(alphabet[rng.IntN(len(alphabet))])
	}
	return b.String()
}

// slugify turns a product name into the part of a URL that identifies it:
// "Ashen Crown Steam Key GLOBAL" becomes "ashen-crown-steam-key-global".
func slugify(name string) string {
	var b strings.Builder
	lastWasDash := true // avoids a dash at the very start
	for _, r := range strings.ToLower(name) {
		isLetterOrDigit := (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')
		switch {
		case isLetterOrDigit:
			b.WriteRune(r)
			lastWasDash = false
		case !lastWasDash:
			b.WriteByte('-')
			lastWasDash = true
		}
	}
	return strings.TrimSuffix(b.String(), "-")
}
