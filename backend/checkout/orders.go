package checkout

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"time"

	"github.com/jackc/pgx/v5"

	"eneba-marketplace/backend/catalog"
)

// ReservationMinutes is how long an unpaid order holds its keys.
const ReservationMinutes = 15

// MaxPendingOrders is how many unpaid orders one user may have at a time.
// Without a limit a single account could reserve the whole stock, let the
// reservations run out, and do it again.
const MaxPendingOrders = 3

// The states of an order.
const (
	StatusPending = "pending"
	StatusPaid    = "paid"
	StatusFailed  = "failed"
	StatusExpired = "expired"
)

// Errors the handlers translate into responses.
var (
	// ErrOrderNotFound: no such order, or it belongs to someone else. The
	// two cases are deliberately one error, so the answer never confirms
	// that another person's order exists.
	ErrOrderNotFound = errors.New("order not found")
	// ErrPriceChanged: the total differs from the one the buyer was shown.
	ErrPriceChanged = errors.New("a price changed since the cart was shown")
	// ErrOfferNotFound: the cart names an offer that does not exist.
	ErrOfferNotFound = errors.New("an offer in the cart no longer exists")
	// ErrTooManyPending: the user already has MaxPendingOrders unpaid orders.
	ErrTooManyPending = errors.New("too many unpaid orders")
	// ErrOrderNotPayable: the order is failed or expired and cannot be paid.
	ErrOrderNotPayable = errors.New("this order can no longer be paid")
)

// StockError says which line could not be filled and how many keys are left.
type StockError struct {
	OfferID     int64
	ProductName string
	Requested   int
	Available   int
}

func (e *StockError) Error() string {
	return fmt.Sprintf("not enough stock for %q: wanted %d, %d left", e.ProductName, e.Requested, e.Available)
}

// OrderItem is one line of an order. Keys is filled only for a paid order.
type OrderItem struct {
	ProductSlug    string   `json:"product_slug"`
	ProductName    string   `json:"product_name"`
	SellerName     string   `json:"seller_name"`
	CoverURL       string   `json:"cover_url"`
	UnitPriceCents int      `json:"unit_price_cents"`
	Quantity       int      `json:"quantity"`
	Keys           []string `json:"keys"`
}

// Order is an order as its owner sees it.
type Order struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalCents int       `json:"total_cents"`
	CreatedAt  time.Time `json:"created_at"`
	ExpiresAt  time.Time `json:"expires_at"`
	// SecondsLeft is how long a pending order still holds its keys, measured
	// by the database's clock. The order page counts down from this number.
	// It does not compare ExpiresAt with the visitor's own clock, which may
	// be wrong by minutes or hours. Zero for an order that is not pending.
	SecondsLeft int        `json:"seconds_left"`
	PaidAt      *time.Time `json:"paid_at"` // null until paid
	// Refunded is true when a payment arrived too late to be fulfilled and
	// was sent back.
	Refunded bool `json:"refunded"`
	// PaymentMode tells the order page which payment step to show:
	// "simulated" or "stripe". The HTTP handler fills it in.
	PaymentMode string      `json:"payment_mode"`
	Items       []OrderItem `json:"items"`
}

// reserveKeysSQL takes up to $2 available keys of offer $1 for order item $3
// and returns the ids of the keys it took. This statement is what makes
// selling a key twice impossible.
//
// The inner SELECT picks available keys and locks those rows (FOR UPDATE)
// until the transaction ends. If another order is taking keys of the same
// offer at this moment, its rows are locked, and SKIP LOCKED makes this
// query pass over them and pick others. Two transactions can therefore never
// pick the same row: whoever locks it first has it, and the other one does
// not even see it.
//
// It is a variable only so that a test can swap in a version without the
// lock and show that the concurrency tests catch the difference.
var reserveKeysSQL = `
	UPDATE keys
	   SET status = 'reserved', order_item_id = $3
	 WHERE id IN (
	       SELECT id FROM keys
	        WHERE offer_id = $1 AND status = 'available'
	        ORDER BY id
	        LIMIT $2
	          FOR UPDATE SKIP LOCKED)
	RETURNING id`

// releaseExpiredSQL is the cleanup. There is no scheduled job, so it runs at
// the start of every order creation and every payment.
//
// It marks pending orders past their deadline as expired and, in the same
// statement, makes the keys of exactly those orders available again.
//
// It goes through the order row first on purpose. A payment locks its order
// row while it works. SKIP LOCKED makes the cleanup pass over an order that
// is being paid at this moment, so a key is never released while it is being
// sold. That order's fate is decided by the payment, which checks the
// deadline itself.
const releaseExpiredSQL = `
	WITH expired AS (
	    UPDATE orders
	       SET status = 'expired'
	     WHERE id IN (
	           SELECT id FROM orders
	            WHERE status = 'pending' AND expires_at <= now()
	            ORDER BY id
	              FOR UPDATE SKIP LOCKED)
	    RETURNING id
	)
	UPDATE keys
	   SET status = 'available', order_item_id = NULL
	 WHERE status = 'reserved'
	   AND order_item_id IN (SELECT oi.id FROM order_items oi
	                           JOIN expired e ON e.id = oi.order_id)`

// releaseOrderKeysSQL makes the reserved keys of one order available again.
const releaseOrderKeysSQL = `
	UPDATE keys
	   SET status = 'available', order_item_id = NULL
	 WHERE status = 'reserved'
	   AND order_item_id IN (SELECT id FROM order_items WHERE order_id = $1)`

// CreateOrder turns a cart into a pending order and reserves its keys.
// expectedTotalCents is the total the buyer was shown.
//
// Everything happens in one transaction. Any early return below rolls it
// all back: no order, no reserved key.
func (s *Store) CreateOrder(ctx context.Context, userID int64, lines []Line, expectedTotalCents int) (Order, error) {
	lines, err := NormalizeLines(lines)
	if err != nil {
		return Order{}, err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx) // does nothing after a successful Commit

	// 1. Give back what has run out of time, so that stock is counted
	//    correctly below.
	if _, err := tx.Exec(ctx, releaseExpiredSQL); err != nil {
		return Order{}, fmt.Errorf("releasing expired reservations: %w", err)
	}

	// 2. The limit on unpaid orders.
	var pending int
	err = tx.QueryRow(ctx,
		"SELECT COUNT(*) FROM orders WHERE user_id = $1 AND status = 'pending'", userID,
	).Scan(&pending)
	if err != nil {
		return Order{}, fmt.Errorf("counting pending orders: %w", err)
	}
	if pending >= MaxPendingOrders {
		return Order{}, ErrTooManyPending
	}

	// 3. Prices come from the database, never from the request.
	cart, err := priceLines(ctx, tx, lines)
	if err != nil {
		return Order{}, err
	}
	if len(cart.MissingOfferIDs) > 0 {
		return Order{}, ErrOfferNotFound
	}

	// 4. Nobody pays a price they did not see.
	if cart.TotalCents != expectedTotalCents {
		return Order{}, ErrPriceChanged
	}

	// 5. The order. The deadline is computed by the database's clock, the
	//    same clock every later check compares it with.
	var orderID string
	err = tx.QueryRow(ctx, `
		INSERT INTO orders (user_id, total_cents, expires_at)
		VALUES ($1, $2, now() + make_interval(mins => $3::int))
		RETURNING id::text`,
		userID, cart.TotalCents, ReservationMinutes,
	).Scan(&orderID)
	if err != nil {
		return Order{}, fmt.Errorf("creating the order: %w", err)
	}

	// 6. Each line: its item row, then its keys.
	for _, line := range cart.Lines {
		var itemID int64
		err := tx.QueryRow(ctx, `
			INSERT INTO order_items (order_id, offer_id, product_slug, product_name,
			                         seller_name, unit_price_cents, quantity)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id`,
			orderID, line.OfferID, line.ProductSlug, line.ProductName,
			line.SellerName, line.UnitPriceCents, line.Quantity,
		).Scan(&itemID)
		if err != nil {
			return Order{}, fmt.Errorf("creating an order item: %w", err)
		}

		reserved, err := countRows(tx.Query(ctx, reserveKeysSQL, line.OfferID, line.Quantity, itemID))
		if err != nil {
			return Order{}, fmt.Errorf("reserving keys: %w", err)
		}
		if reserved < line.Quantity {
			// Not enough keys, possibly because another buyer took the last
			// one a moment ago. Returning here rolls back this line's
			// partial reservation and every earlier line with it.
			return Order{}, &StockError{
				OfferID:     line.OfferID,
				ProductName: line.ProductName,
				Requested:   line.Quantity,
				Available:   reserved,
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("committing the order: %w", err)
	}
	return s.GetOrder(ctx, userID, orderID)
}

// countRows reads a query's rows to the end and returns how many there were.
func countRows(rows pgx.Rows, err error) (int, error) {
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	count := 0
	for rows.Next() {
		count++
	}
	return count, rows.Err()
}

// Pay settles a pending order: approved marks it paid and its keys sold,
// declined marks it failed and releases its keys.
//
// It is safe to call more than once. A second call for a paid order changes
// nothing and returns the order, so a double click, a second tab or a
// repeated webhook cannot sell keys twice or count a sale twice.
//
// It returns ErrOrderNotPayable when the order is failed or expired. The
// returned Order is still valid in that case and shows the final state.
func (s *Store) Pay(ctx context.Context, userID int64, orderID string, approved bool) (Order, error) {
	if !looksLikeUUID(orderID) {
		return Order{}, ErrOrderNotFound
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return Order{}, fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, releaseExpiredSQL); err != nil {
		return Order{}, fmt.Errorf("releasing expired reservations: %w", err)
	}

	// Lock the order row. A second payment for the same order waits here
	// until this transaction ends, then reads the state this one left.
	var status string
	var overdue bool
	err = tx.QueryRow(ctx, `
		SELECT status, expires_at <= now()
		  FROM orders
		 WHERE id = $1 AND user_id = $2
		   FOR UPDATE`, orderID, userID,
	).Scan(&status, &overdue)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("locking the order: %w", err)
	}

	payable := true
	switch {
	case status == StatusPaid:
		// Already done. Change nothing.
		return s.GetOrder(ctx, userID, orderID)

	case status != StatusPending:
		payable = false // failed or expired earlier

	case overdue:
		// The deadline passed between the cleanup above and taking the lock.
		if err := endOrder(ctx, tx, orderID, StatusExpired); err != nil {
			return Order{}, err
		}
		payable = false

	case !approved:
		if err := endOrder(ctx, tx, orderID, StatusFailed); err != nil {
			return Order{}, err
		}

	default:
		if err := markPaid(ctx, tx, orderID); err != nil {
			return Order{}, err
		}
		if s.RestockSold {
			if err := restockLowOffers(ctx, tx, orderID); err != nil {
				return Order{}, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return Order{}, fmt.Errorf("committing the payment: %w", err)
	}

	order, err := s.GetOrder(ctx, userID, orderID)
	if err != nil {
		return Order{}, err
	}
	if !payable {
		return order, ErrOrderNotPayable
	}
	return order, nil
}

// endOrder moves a pending order to failed or expired and releases its keys.
func endOrder(ctx context.Context, tx pgx.Tx, orderID, status string) error {
	if _, err := tx.Exec(ctx, "UPDATE orders SET status = $2 WHERE id = $1", orderID, status); err != nil {
		return fmt.Errorf("marking the order %s: %w", status, err)
	}
	if _, err := tx.Exec(ctx, releaseOrderKeysSQL, orderID); err != nil {
		return fmt.Errorf("releasing the order's keys: %w", err)
	}
	return nil
}

// markPaid is the one place an order is fulfilled: keys become sold, the
// order becomes paid, and the sale is counted for the best sellers block.
func markPaid(ctx context.Context, tx pgx.Tx, orderID string) error {
	sold, err := tx.Exec(ctx, `
		UPDATE keys SET status = 'sold'
		 WHERE status = 'reserved'
		   AND order_item_id IN (SELECT id FROM order_items WHERE order_id = $1)`, orderID)
	if err != nil {
		return fmt.Errorf("marking keys sold: %w", err)
	}

	// A last check before taking the order as paid: the number of keys just
	// sold must be the number ordered. If it is not, something is wrong with
	// the data, and failing here (which rolls everything back) is better
	// than handing over fewer keys than were paid for.
	var ordered int
	err = tx.QueryRow(ctx, "SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE order_id = $1", orderID).Scan(&ordered)
	if err != nil {
		return fmt.Errorf("summing the order's quantities: %w", err)
	}
	if int(sold.RowsAffected()) != ordered {
		return fmt.Errorf("order %s has %d keys reserved but %d ordered", orderID, sold.RowsAffected(), ordered)
	}

	if _, err := tx.Exec(ctx, "UPDATE orders SET status = 'paid', paid_at = now() WHERE id = $1", orderID); err != nil {
		return fmt.Errorf("marking the order paid: %w", err)
	}

	// Add this order's quantities to the sales counts. ON CONFLICT handles
	// both a product's first sale (insert) and later ones (add to the row).
	_, err = tx.Exec(ctx, `
		INSERT INTO product_sales (product_id, units_sold)
		SELECT o.product_id, SUM(oi.quantity)
		  FROM order_items oi
		  JOIN offers o ON o.id = oi.offer_id
		 WHERE oi.order_id = $1
		 GROUP BY o.product_id
		ON CONFLICT (product_id)
		DO UPDATE SET units_sold = product_sales.units_sold + EXCLUDED.units_sold`, orderID)
	if err != nil {
		return fmt.Errorf("counting the sale: %w", err)
	}
	return nil
}

// displayStatusSQL is the status shown to the buyer. A pending order past
// its deadline has not been touched by the cleanup yet, but for the buyer it
// is already expired.
const displayStatusSQL = `CASE WHEN status = 'pending' AND expires_at <= now() THEN 'expired' ELSE status END`

// GetOrder returns one order of this user, with its items. The keys are
// included only when the order is paid.
func (s *Store) GetOrder(ctx context.Context, userID int64, orderID string) (Order, error) {
	if !looksLikeUUID(orderID) {
		return Order{}, ErrOrderNotFound
	}

	var order Order
	err := s.db.QueryRow(ctx, `
		SELECT id::text, `+displayStatusSQL+`, total_cents, created_at, expires_at, paid_at,
		       CASE WHEN status = 'pending'
		            THEN GREATEST(0, FLOOR(EXTRACT(EPOCH FROM expires_at - now())))::int
		            ELSE 0 END,
		       refund_ref IS NOT NULL
		  FROM orders
		 WHERE id = $1 AND user_id = $2`, orderID, userID,
	).Scan(&order.ID, &order.Status, &order.TotalCents, &order.CreatedAt, &order.ExpiresAt, &order.PaidAt, &order.SecondsLeft, &order.Refunded)
	if errors.Is(err, pgx.ErrNoRows) {
		return Order{}, ErrOrderNotFound
	}
	if err != nil {
		return Order{}, fmt.Errorf("loading the order: %w", err)
	}

	// The keys are selected only for a sold key of a paid order. Both
	// conditions are in the query itself, so no code path can return a key
	// of an unpaid order by forgetting an "if".
	rows, err := s.db.Query(ctx, `
		SELECT oi.product_slug, oi.product_name, oi.seller_name, oi.unit_price_cents, oi.quantity,
		       ARRAY(SELECT k.code FROM keys k
		              WHERE k.order_item_id = oi.id AND k.status = 'sold' AND $2::boolean
		              ORDER BY k.id),
		       -- The game's title, to find its cover. An order keeps its own
		       -- copy of the product's name but not of the title, so it is
		       -- looked up. If the product is gone, the title is empty and
		       -- the cover falls back to the drawn one.
		       COALESCE((SELECT p.title FROM products p WHERE p.slug = oi.product_slug), '')
		  FROM order_items oi
		 WHERE oi.order_id = $1
		 ORDER BY oi.id`, orderID, order.Status == StatusPaid)
	if err != nil {
		return Order{}, fmt.Errorf("loading the order's items: %w", err)
	}
	defer rows.Close()

	order.Items = []OrderItem{}
	for rows.Next() {
		var item OrderItem
		var title string
		err := rows.Scan(&item.ProductSlug, &item.ProductName, &item.SellerName,
			&item.UnitPriceCents, &item.Quantity, &item.Keys, &title)
		if err != nil {
			return Order{}, fmt.Errorf("reading an order item: %w", err)
		}
		item.CoverURL, _ = catalog.CoverFor(title, item.ProductSlug)
		order.Items = append(order.Items, item)
	}
	if err := rows.Err(); err != nil {
		return Order{}, fmt.Errorf("reading order items: %w", err)
	}

	order.CreatedAt, order.ExpiresAt = order.CreatedAt.UTC(), order.ExpiresAt.UTC()
	return order, nil
}

// OrderSummary is one row of the order history.
type OrderSummary struct {
	ID         string    `json:"id"`
	Status     string    `json:"status"`
	TotalCents int       `json:"total_cents"`
	CreatedAt  time.Time `json:"created_at"`
	KeyCount   int       `json:"key_count"`  // keys in the order, over all lines
	LineCount  int       `json:"line_count"` // different products in the order
	FirstItem  string    `json:"first_item"` // a product name, to recognise the order by
}

// ListOrders returns this user's orders, newest first. No keys: those are on
// the order's own page.
func (s *Store) ListOrders(ctx context.Context, userID int64) ([]OrderSummary, error) {
	rows, err := s.db.Query(ctx, `
		SELECT o.id::text, `+displayStatusSQL+`, o.total_cents, o.created_at,
		       (SELECT COALESCE(SUM(quantity), 0) FROM order_items WHERE order_id = o.id),
		       (SELECT COUNT(*) FROM order_items WHERE order_id = o.id),
		       COALESCE((SELECT product_name FROM order_items WHERE order_id = o.id ORDER BY id LIMIT 1), '')
		  FROM orders o
		 WHERE o.user_id = $1
		 ORDER BY o.created_at DESC, o.id
		 LIMIT 100`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing orders: %w", err)
	}
	defer rows.Close()

	orders := []OrderSummary{}
	for rows.Next() {
		var o OrderSummary
		if err := rows.Scan(&o.ID, &o.Status, &o.TotalCents, &o.CreatedAt, &o.KeyCount, &o.LineCount, &o.FirstItem); err != nil {
			return nil, fmt.Errorf("reading an order row: %w", err)
		}
		o.CreatedAt = o.CreatedAt.UTC()
		orders = append(orders, o)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("reading order rows: %w", err)
	}
	return orders, nil
}

var uuidPattern = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// looksLikeUUID keeps text that is not an order id away from the database.
// PostgreSQL would reject it with an error, and that error would surface as
// a 500 where a 404 is the right answer.
func looksLikeUUID(text string) bool {
	return uuidPattern.MatchString(text)
}
