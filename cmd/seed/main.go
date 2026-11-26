// Command seed fills the database with invented demo data. It works in
// sections and fills each one only if it is still empty:
//
//   - the catalog: products, sellers, offers and keys
//   - the homepage: layout, banners, sales figures and deals
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
//	go run ./cmd/seed -reset                (delete all seed data first, then fill)
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
		_, err := tx.Exec(ctx, `
			TRUNCATE keys, offers, sellers, products,
			         home_layout, banners, deals, product_sales
			RESTART IDENTITY CASCADE`)
		if err != nil {
			return fmt.Errorf("emptying the seed tables: %w", err)
		}
		fmt.Println("reset: emptied the seed tables")
	}

	// Each section: skip it if its marker table already has rows.
	sections := []struct {
		name        string
		markerTable string
		fill        func(context.Context, pgx.Tx) error
	}{
		{"catalog", "products", seedCatalog},
		{"homepage", "banners", seedHome},
	}
	for _, section := range sections {
		var existing int
		// The table name comes from the list above, never from input.
		err := tx.QueryRow(ctx, "SELECT COUNT(*) FROM "+section.markerTable).Scan(&existing)
		if err != nil {
			return fmt.Errorf("counting %s: %w", section.markerTable, err)
		}
		if existing > 0 {
			fmt.Printf("%s: already there (%d rows in %s), left alone\n", section.name, existing, section.markerTable)
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
