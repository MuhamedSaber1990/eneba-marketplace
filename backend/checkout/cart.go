// Package checkout turns a cart into an order: it prices carts, reserves
// keys, settles payments and delivers the keys. The rules it enforces are
// the ones that must never break:
//
//   - a key is sold at most once
//   - a buyer pays the price they were shown
//   - an order's keys are visible only to its owner, and only once paid
package checkout

import (
	"context"
	"fmt"
	"slices"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"eneba-marketplace/backend/catalog"
)

// Limits on what one order may contain.
const (
	MaxLines           = 10 // different offers in one order
	MaxQuantityPerLine = 5  // keys of one offer in one order
)

// ValidationError is a problem with the request itself. Its message is
// written for the visitor.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

// Store runs the checkout queries.
type Store struct {
	db *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{db: db}
}

// Line is one entry of a cart as the browser sends it: which offer and how
// many. It carries no price on purpose. Prices always come from the database.
type Line struct {
	OfferID  int64 `json:"offer_id"`
	Quantity int   `json:"quantity"`
}

// NormalizeLines checks a cart and puts it in a standard form: the same
// offer listed twice is merged, and lines are sorted by offer id.
//
// Sorting is not cosmetic. Every order then takes its keys in the same order
// of offers, so two orders that want the same two offers can never each hold
// one and wait for the other.
func NormalizeLines(lines []Line) ([]Line, error) {
	if len(lines) == 0 {
		return nil, ValidationError{"The cart is empty."}
	}

	quantities := map[int64]int{}
	for _, line := range lines {
		if line.OfferID <= 0 {
			return nil, ValidationError{"The cart contains an unknown offer."}
		}
		if line.Quantity < 1 {
			return nil, ValidationError{"Every item needs a quantity of at least 1."}
		}
		quantities[line.OfferID] += line.Quantity
	}
	if len(quantities) > MaxLines {
		return nil, ValidationError{fmt.Sprintf("An order can hold at most %d different items.", MaxLines)}
	}

	merged := make([]Line, 0, len(quantities))
	for offerID, quantity := range quantities {
		if quantity > MaxQuantityPerLine {
			return nil, ValidationError{fmt.Sprintf("You can buy at most %d of one item per order.", MaxQuantityPerLine)}
		}
		merged = append(merged, Line{OfferID: offerID, Quantity: quantity})
	}
	slices.SortFunc(merged, func(a, b Line) int { return int(a.OfferID - b.OfferID) })
	return merged, nil
}

// PricedLine is a cart line with everything the cart page shows, read from
// the database.
type PricedLine struct {
	OfferID        int64  `json:"offer_id"`
	Quantity       int    `json:"quantity"`
	ProductSlug    string `json:"product_slug"`
	ProductName    string `json:"product_name"`
	Region         string `json:"region"`
	SellerName     string `json:"seller_name"`
	CoverURL       string `json:"cover_url"`
	UnitPriceCents int    `json:"unit_price_cents"`
	LineTotalCents int    `json:"line_total_cents"`
	Stock          int    `json:"stock"`     // keys available right now
	Available      bool   `json:"available"` // stock covers the quantity
}

// PricedCart is a whole cart, priced.
type PricedCart struct {
	Lines []PricedLine `json:"lines"`
	// MissingOfferIDs are offers from the request that no longer exist. The
	// cart page removes them.
	MissingOfferIDs []int64 `json:"missing_offer_ids"`
	TotalCents      int     `json:"total_cents"`
	// CanCheckout is true when there is at least one line and every line
	// has enough stock.
	CanCheckout bool `json:"can_checkout"`
}

// querier is anything that can run a query. Both the connection pool and a
// transaction can, so the same pricing code serves the cart page (no
// transaction) and order creation (inside one).
type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

// PriceCart looks up the current price and stock of every line. It reserves
// nothing: two visitors can both see "1 in stock". Who gets the key is
// decided when an order is created.
func (s *Store) PriceCart(ctx context.Context, lines []Line) (PricedCart, error) {
	lines, err := NormalizeLines(lines)
	if err != nil {
		return PricedCart{}, err
	}
	return priceLines(ctx, s.db, lines)
}

func priceLines(ctx context.Context, q querier, lines []Line) (PricedCart, error) {
	offerIDs := make([]int64, len(lines))
	quantityOf := map[int64]int{}
	for i, line := range lines {
		offerIDs[i] = line.OfferID
		quantityOf[line.OfferID] = line.Quantity
	}

	rows, err := q.Query(ctx, `
		SELECT o.id, o.price_cents, p.slug, p.name, p.region, s.name,
		       (SELECT COUNT(*) FROM keys k
		         WHERE k.offer_id = o.id AND k.status = 'available') AS stock
		  FROM offers o
		  JOIN products p ON p.id = o.product_id
		  JOIN sellers s  ON s.id = o.seller_id
		 WHERE o.id = ANY($1::bigint[])
		 ORDER BY o.id`, offerIDs)
	if err != nil {
		return PricedCart{}, fmt.Errorf("pricing the cart: %w", err)
	}
	defer rows.Close()

	cart := PricedCart{Lines: []PricedLine{}, MissingOfferIDs: []int64{}, CanCheckout: true}
	found := map[int64]bool{}
	for rows.Next() {
		var line PricedLine
		err := rows.Scan(&line.OfferID, &line.UnitPriceCents, &line.ProductSlug, &line.ProductName,
			&line.Region, &line.SellerName, &line.Stock)
		if err != nil {
			return PricedCart{}, fmt.Errorf("reading a cart line: %w", err)
		}
		line.Quantity = quantityOf[line.OfferID]
		line.LineTotalCents = line.UnitPriceCents * line.Quantity
		line.Available = line.Stock >= line.Quantity
		line.CoverURL = catalog.CoverURL(line.ProductSlug)

		cart.Lines = append(cart.Lines, line)
		cart.TotalCents += line.LineTotalCents
		cart.CanCheckout = cart.CanCheckout && line.Available
		found[line.OfferID] = true
	}
	if err := rows.Err(); err != nil {
		return PricedCart{}, fmt.Errorf("reading cart lines: %w", err)
	}

	for _, id := range offerIDs {
		if !found[id] {
			cart.MissingOfferIDs = append(cart.MissingOfferIDs, id)
		}
	}
	cart.CanCheckout = cart.CanCheckout && len(cart.Lines) > 0 && len(cart.MissingOfferIDs) == 0
	return cart, nil
}
