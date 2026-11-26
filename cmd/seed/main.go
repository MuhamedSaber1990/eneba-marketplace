// Command seed fills the database with invented demo data. It works in
// sections and fills each one only if it is still empty:
//
//   - the catalog: products, sellers, offers and keys
//   - the homepage: layout, banners, sales figures and deals
//   - the demo account, whose password is published in the README
//
// So it is safe to run again after a new section is added: what is already
// there is left alone. Each section uses fixed random seeds, so every
// database gets the same data (only the deals' end dates depend on the day
// the command runs).
//
// Usage, from the repository root:
//
//	go run ./cmd/seed                       (uses .env, the dev database)
//	go run ./cmd/seed -env .env.production  (the production database)
//	go run ./cmd/seed -reset                (delete all seed data and all orders first, then fill)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

func main() {
	envFile := flag.String("env", ".env", "file to read DATABASE_URL_UNPOOLED from")
	reset := flag.Bool("reset", false, "delete the existing seed data before seeding")
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
	// Everything happens in one transaction: if any section fails, the
	// database is left exactly as it was.
	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if reset {
		// CASCADE also empties tables that point at these ones.
		// Orders are emptied too: they point at the offers and keys being
		// replaced, so they cannot outlive them.
		_, err := tx.Exec(ctx, `
			TRUNCATE order_items, orders, keys, offers, sellers, products,
			         home_layout, banners, deals, product_sales
			RESTART IDENTITY CASCADE`)
		if err != nil {
			return fmt.Errorf("emptying the seed tables: %w", err)
		}
		// Only the demo account is seed data. Accounts that people created
		// are theirs, and a reset never touches them.
		if _, err := tx.Exec(ctx, "DELETE FROM users WHERE is_demo"); err != nil {
			return fmt.Errorf("deleting the demo account: %w", err)
		}
		fmt.Println("reset: emptied the seed tables")
	}

	// Each section has a query that counts what it would create. If the
	// count is above zero the section is already there and is skipped.
	sections := []struct {
		name     string
		countSQL string
		fill     func(context.Context, pgx.Tx) error
	}{
		{"catalog", "SELECT COUNT(*) FROM products", seedCatalog},
		{"homepage", "SELECT COUNT(*) FROM banners", seedHome},
		{"demo account", "SELECT COUNT(*) FROM users WHERE is_demo", seedDemoUser},
	}
	for _, section := range sections {
		var existing int
		if err := tx.QueryRow(ctx, section.countSQL).Scan(&existing); err != nil {
			return fmt.Errorf("checking the %s: %w", section.name, err)
		}
		if existing > 0 {
			fmt.Printf("%s: already there, left alone\n", section.name)
			continue
		}
		if err := section.fill(ctx, tx); err != nil {
			return fmt.Errorf("seeding the %s: %w", section.name, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing: %w", err)
	}
	return nil
}
