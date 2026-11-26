// Package db opens the connection pool to PostgreSQL.
package db

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect builds a connection pool from the DATABASE_URL environment variable.
//
// A pool keeps a few connections open and hands them out per query, so each
// request does not pay the cost of connecting. pgxpool connects lazily: this
// function does not touch the network, the first query does.
func Connect(ctx context.Context) (*pgxpool.Pool, error) {
	// On Vercel every function instance has its own pool, and many instances
	// can run at once. A small pool per instance keeps the total low.
	return ConnectWithMaxConns(ctx, 4)
}

// ConnectWithMaxConns is Connect with a chosen pool size. The concurrency
// tests use a larger pool, so that many transactions really run at once.
func ConnectWithMaxConns(ctx context.Context, maxConns int32) (*pgxpool.Pool, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}

	config, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, err
	}
	config.MaxConns = maxConns

	// Neon's pooled connection string goes through PgBouncer, which shares
	// server connections between clients. pgx normally prepares a statement on
	// one connection and reuses it, which can break when the next query lands
	// on a different connection. Exec mode sends each query with its
	// parameters in one step and prepares nothing.
	config.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	return pgxpool.NewWithConfig(ctx, config)
}
