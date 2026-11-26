package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5"
)

// ConnectDirect opens one direct (not pooled) connection, read from
// DATABASE_URL_UNPOOLED. The migrate and seed commands use it.
//
// Neon's pooled address shares server connections between clients, which can
// break schema changes and long transactions. Tools that change the schema
// should talk to Postgres directly, and the API keeps using the pool.
//
// The second return value is a label for the database (the first part of the
// host name), safe to print so the person running the command can see which
// database is about to change.
func ConnectDirect(ctx context.Context) (*pgx.Conn, string, error) {
	url := os.Getenv("DATABASE_URL_UNPOOLED")
	if url == "" {
		return nil, "", errors.New("DATABASE_URL_UNPOOLED is not set")
	}

	config, err := pgx.ParseConfig(url)
	if err != nil {
		return nil, "", fmt.Errorf("parsing DATABASE_URL_UNPOOLED: %w", err)
	}

	label, _, _ := strings.Cut(config.Host, ".")
	if strings.Contains(label, "-pooler") {
		return nil, "", errors.New("DATABASE_URL_UNPOOLED points at the pooled host (its name contains -pooler)")
	}

	conn, err := pgx.ConnectConfig(ctx, config)
	if err != nil {
		return nil, "", fmt.Errorf("connecting to %s: %w", label, err)
	}
	return conn, label, nil
}
