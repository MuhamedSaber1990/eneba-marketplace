package checkout

import (
	"context"
	"crypto/rand"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// The demo restock.
//
// On the public demo site paying costs nothing: payment is simulated or uses
// test cards, and the demo account's password is published. Anyone could buy
// every key and leave the store sold out. Since the keys are invented
// anyway, the store makes more: when a sale leaves an offer with fewer than
// restockBelow available keys, restockBatch new ones are added.
//
// A buyer still sees stock go down after a purchase, and reserving the last
// key still works as it should. What cannot happen is an offer being bought
// empty. A real shop would never do this, which is why it sits behind the
// Store.RestockSold switch and is off everywhere but the demo site.
const (
	restockBelow = 3
	restockBatch = 5
)

// restockLowOffers tops up the offers of a just-paid order that have run
// low. It runs inside the payment's transaction, so the new keys appear
// together with the sale or not at all.
func restockLowOffers(ctx context.Context, tx pgx.Tx, orderID string) error {
	rows, err := tx.Query(ctx, `
		SELECT DISTINCT oi.offer_id
		  FROM order_items oi
		 WHERE oi.order_id = $1
		   AND (SELECT COUNT(*) FROM keys k
		         WHERE k.offer_id = oi.offer_id AND k.status = 'available') < $2`,
		orderID, restockBelow)
	if err != nil {
		return fmt.Errorf("finding offers to restock: %w", err)
	}
	lowOffers, err := pgx.CollectRows(rows, pgx.RowTo[int64])
	if err != nil {
		return fmt.Errorf("reading offers to restock: %w", err)
	}

	for _, offerID := range lowOffers {
		codes := make([]string, restockBatch)
		for i := range codes {
			if codes[i], err = newKeyCode(); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx,
			"INSERT INTO keys (offer_id, code) SELECT $1, unnest($2::text[])", offerID, codes)
		if err != nil {
			return fmt.Errorf("restocking offer %d: %w", offerID, err)
		}
	}
	return nil
}

// keyAlphabet leaves out letters and digits that look alike (0 and O, 1 and
// I). It has 32 characters, so one random byte maps onto it evenly: a byte
// has 256 values and 256 is a multiple of 32.
const keyAlphabet = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

// newKeyCode makes a code shaped like XXXXX-XXXXX-XXXXX from the operating
// system's secure random source. With 32 possibilities for each of its 15
// characters, two codes being equal is not a practical concern, and the
// UNIQUE rule on keys.code would refuse a duplicate anyway.
func newKeyCode() (string, error) {
	random := make([]byte, 15)
	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generating a key: %w", err)
	}
	var b strings.Builder
	for i, value := range random {
		if i > 0 && i%5 == 0 {
			b.WriteByte('-')
		}
		b.WriteByte(keyAlphabet[int(value)%len(keyAlphabet)])
	}
	return b.String(), nil
}
