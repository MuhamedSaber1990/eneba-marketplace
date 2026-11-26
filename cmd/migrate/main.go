// Command migrate brings the database schema up to date.
//
// It runs every .sql file in the migrations folder that has not run before,
// in file name order (001_..., 002_...), and remembers what it ran in a table
// called schema_migrations. Running it twice is safe: the second run finds
// nothing to do.
//
// Usage, from the repository root:
//
//	go run ./cmd/migrate                       (uses .env, the dev database)
//	go run ./cmd/migrate -env .env.production  (the production database)
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

func main() {
	envFile := flag.String("env", ".env", "file to read DATABASE_URL_UNPOOLED from")
	dir := flag.String("dir", "migrations", "folder with the .sql files")
	flag.Parse()

	config.LoadDotEnv(*envFile)
	ctx := context.Background()

	conn, label, err := db.ConnectDirect(ctx)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close(ctx)
	fmt.Printf("database: %s (from %s)\n", label, *envFile)

	if err := run(ctx, conn, *dir); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, conn *pgx.Conn, dir string) error {
	// The bookkeeping table. IF NOT EXISTS makes the very first run work.
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version    TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT now()
		)`)
	if err != nil {
		return fmt.Errorf("creating schema_migrations: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return fmt.Errorf("listing %s: %w", dir, err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .sql files in %s (run this from the repository root)", dir)
	}
	sort.Strings(files) // 001 before 002 before 010

	applied := 0
	for _, file := range files {
		version := filepath.Base(file)

		var alreadyApplied bool
		err := conn.QueryRow(ctx,
			"SELECT EXISTS (SELECT 1 FROM schema_migrations WHERE version = $1)", version,
		).Scan(&alreadyApplied)
		if err != nil {
			return fmt.Errorf("checking %s: %w", version, err)
		}
		if alreadyApplied {
			continue
		}

		if err := apply(ctx, conn, file, version); err != nil {
			return err
		}
		fmt.Printf("applied %s\n", version)
		applied++
	}

	if applied == 0 {
		fmt.Println("nothing to do, the schema is up to date")
	}
	return nil
}

// apply runs one migration file and records it, inside a single transaction.
// If any statement in the file fails, the whole file is rolled back and it is
// not recorded, so a migration is never half applied.
func apply(ctx context.Context, conn *pgx.Conn, file, version string) error {
	sql, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("reading %s: %w", file, err)
	}

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction for %s: %w", version, err)
	}
	// Rollback after a successful Commit does nothing, so deferring it covers
	// every early return below.
	defer tx.Rollback(ctx)

	// Exec without arguments sends the text as is, which lets one file hold
	// many statements separated by semicolons.
	if _, err := tx.Exec(ctx, string(sql)); err != nil {
		return fmt.Errorf("running %s: %w", version, err)
	}
	if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", version); err != nil {
		return fmt.Errorf("recording %s: %w", version, err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing %s: %w", version, err)
	}
	return nil
}
