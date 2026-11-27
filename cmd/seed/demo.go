package main

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"

	"eneba-marketplace/backend/auth"
)

// The demo account, for anyone who wants to try the signed-in parts of the
// site without registering. These credentials are public on purpose: they
// are also printed in the README. Nothing of value is behind them.
const (
	demoEmail    = "demo@keybay.test"
	demoPassword = "keybay-demo-2026"
)

// seedDemoUser creates the demo account. The password goes through the same
// hashing as a real sign-up, so the stored row looks like any other user's.
func seedDemoUser(ctx context.Context, tx pgx.Tx) error {
	hash, err := auth.HashPassword(demoPassword)
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx,
		// The demo account never goes through the sign-up form, so it has
		// no terms_accepted_at. Its name matches what migration 008 gives
		// a demo account that existed before it.
		"INSERT INTO users (email, password_hash, name, is_demo) VALUES ($1, $2, 'Demo Player', true)",
		demoEmail, hash)
	if err != nil {
		return fmt.Errorf("inserting the demo account: %w", err)
	}
	fmt.Printf("demo account: seeded %s\n", demoEmail)
	return nil
}
