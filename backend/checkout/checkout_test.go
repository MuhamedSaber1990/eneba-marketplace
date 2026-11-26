package checkout

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
)

// ---- No database needed ----

func TestNormalizeLines(t *testing.T) {
	tests := []struct {
		name    string
		lines   []Line
		want    []Line
		wantErr bool
	}{
		{name: "one line", lines: []Line{{7, 2}}, want: []Line{{7, 2}}},
		{name: "sorted by offer id", lines: []Line{{9, 1}, {3, 1}, {5, 2}}, want: []Line{{3, 1}, {5, 2}, {9, 1}}},
		{name: "the same offer twice is merged", lines: []Line{{4, 2}, {4, 1}}, want: []Line{{4, 3}}},
		{name: "empty cart", lines: nil, wantErr: true},
		{name: "quantity zero", lines: []Line{{4, 0}}, wantErr: true},
		{name: "negative quantity", lines: []Line{{4, -1}}, wantErr: true},
		{name: "more than five of one offer", lines: []Line{{4, 6}}, wantErr: true},
		{name: "more than five after merging", lines: []Line{{4, 3}, {4, 3}}, wantErr: true},
		{name: "offer id zero", lines: []Line{{0, 1}}, wantErr: true},
		{name: "eleven different offers", lines: []Line{{1, 1}, {2, 1}, {3, 1}, {4, 1}, {5, 1}, {6, 1}, {7, 1}, {8, 1}, {9, 1}, {10, 1}, {11, 1}}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NormalizeLines(tt.lines)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error = %v, wantErr = %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLooksLikeUUID(t *testing.T) {
	for text, want := range map[string]bool{
		"0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f": true,
		"0B0F3A52-8F0F-4C0A-9D0E-2B7C1C3D4E5F": true,
		"":                                     false,
		"123":                                  false,
		"0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5":  false,
		"0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f; DROP TABLE orders": false,
	} {
		if got := looksLikeUUID(text); got != want {
			t.Errorf("looksLikeUUID(%q) = %v, want %v", text, got, want)
		}
	}
}

// ---- Fixtures for the database tests ----

// world is a private corner of the database for one test: its own product,
// offers with a chosen number of keys, and its own users. Nothing in it is
// shared with the seed data or with other tests, so tests cannot disturb
// each other, and everything is deleted afterwards.
type world struct {
	t      *testing.T
	pool   *pgxpool.Pool
	store  *Store
	suffix string
	// productID is the test product every offer of this world belongs to.
	productID int64
	userIDs   []int64
	cleaned   bool
}

// poolSize is how many database connections the tests may use at once, and
// so how many transactions can truly overlap in the concurrency tests.
const poolSize = 20

func newWorld(t *testing.T) *world {
	t.Helper()
	config.LoadDotEnv("../../.env")
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL is not set, skipping database tests")
	}
	ctx := context.Background()
	pool, err := db.ConnectWithMaxConns(ctx, poolSize)
	if err != nil {
		t.Fatalf("connecting: %v", err)
	}

	random := make([]byte, 6)
	rand.Read(random)
	w := &world{t: t, pool: pool, store: NewStore(pool), suffix: hex.EncodeToString(random)}

	err = pool.QueryRow(ctx, `
		INSERT INTO products (slug, title, name, description, platform, region, genre,
		                      release_date, cover_color_a, cover_color_b)
		VALUES ($1, 'Test Game', $2, 'A product that exists only during a test.',
		        'TestPlatform', 'TestRegion', 'Test', '2020-01-01', '#000000', '#111111')
		RETURNING id`,
		"test-"+w.suffix, "Test Game "+w.suffix,
	).Scan(&w.productID)
	if err != nil {
		t.Fatalf("creating the test product: %v", err)
	}

	t.Cleanup(w.cleanup)
	return w
}

// cleanup deletes everything the world created, in an order the foreign
// keys allow: keys point at order items, order items at orders and offers.
func (w *world) cleanup() {
	if w.cleaned { // a test may clean up early; the automatic call then does nothing
		return
	}
	w.cleaned = true
	ctx := context.Background()
	steps := []struct {
		sql string
		arg any
	}{
		{"DELETE FROM keys WHERE offer_id IN (SELECT id FROM offers WHERE product_id = $1)", w.productID},
		{"DELETE FROM orders WHERE user_id = ANY($1::bigint[])", w.userIDs},
		{"DELETE FROM product_sales WHERE product_id = $1", w.productID},
		{"DELETE FROM products WHERE id = $1", w.productID}, // its offers go with it
		{"DELETE FROM users WHERE id = ANY($1::bigint[])", w.userIDs},
	}
	for _, step := range steps {
		if _, err := w.pool.Exec(ctx, step.sql, step.arg); err != nil {
			w.t.Errorf("cleanup failed (%s): %v", step.sql, err)
		}
	}
	w.pool.Close()
}

// warmUp opens every connection of the pool before a race starts.
//
// A pool opens connections one at a time, as they are first needed, and
// opening one takes a couple of hundred milliseconds. Without this, the
// first goroutine of a race gets its connection well before the others and
// finishes its whole transaction before they have even begun. The test
// would then pass without two transactions ever meeting, and so would pass
// even with a broken reservation query. (That is exactly what an early
// version of these tests did.)
func (w *world) warmUp() {
	w.t.Helper()
	conns := make([]*pgxpool.Conn, 0, poolSize)
	for range poolSize {
		conn, err := w.pool.Acquire(context.Background())
		if err != nil {
			w.t.Fatalf("opening a connection: %v", err)
		}
		conns = append(conns, conn) // held, so the next Acquire must open a new one
	}
	for _, conn := range conns {
		if err := conn.Ping(context.Background()); err != nil {
			w.t.Fatalf("checking a connection: %v", err)
		}
		conn.Release()
	}
}

// offer creates an offer of the test product with the given price and
// number of available keys. Each call uses a different seller, because a
// product can have only one offer per seller.
func (w *world) offer(priceCents, keys int) int64 {
	w.t.Helper()
	ctx := context.Background()

	var offerID int64
	err := w.pool.QueryRow(ctx, `
		INSERT INTO offers (product_id, seller_id, price_cents)
		SELECT $1, s.id, $2
		  FROM sellers s
		 WHERE s.id NOT IN (SELECT seller_id FROM offers WHERE product_id = $1)
		 ORDER BY s.id LIMIT 1
		RETURNING id`, w.productID, priceCents,
	).Scan(&offerID)
	if err != nil {
		w.t.Fatalf("creating a test offer: %v", err)
	}

	codes := make([]string, keys)
	for i := range codes {
		codes[i] = fmt.Sprintf("TEST-%s-%d-%d", w.suffix, offerID, i)
	}
	_, err = w.pool.Exec(ctx,
		"INSERT INTO keys (offer_id, code) SELECT $1, unnest($2::text[])", offerID, codes)
	if err != nil {
		w.t.Fatalf("creating test keys: %v", err)
	}
	return offerID
}

// users creates n accounts and returns their ids. They cannot sign in: the
// stored hash is not a hash of anything.
func (w *world) users(n int) []int64 {
	w.t.Helper()
	ids := make([]int64, n)
	for i := range ids {
		email := fmt.Sprintf("buyer-%d-%s@keybay.example", len(w.userIDs), w.suffix)
		err := w.pool.QueryRow(context.Background(),
			"INSERT INTO users (email, password_hash) VALUES ($1, 'not-a-hash') RETURNING id", email,
		).Scan(&ids[i])
		if err != nil {
			w.t.Fatalf("creating a test user: %v", err)
		}
		w.userIDs = append(w.userIDs, ids[i])
	}
	return ids
}

// keyCounts returns how many keys of an offer are in each state.
func (w *world) keyCounts(offerID int64) (available, reserved, sold int) {
	w.t.Helper()
	err := w.pool.QueryRow(context.Background(), `
		SELECT COUNT(*) FILTER (WHERE status = 'available'),
		       COUNT(*) FILTER (WHERE status = 'reserved'),
		       COUNT(*) FILTER (WHERE status = 'sold')
		  FROM keys WHERE offer_id = $1`, offerID,
	).Scan(&available, &reserved, &sold)
	if err != nil {
		w.t.Fatalf("counting keys: %v", err)
	}
	return available, reserved, sold
}

func (w *world) wantKeys(offerID int64, available, reserved, sold int) {
	w.t.Helper()
	a, r, s := w.keyCounts(offerID)
	if a != available || r != reserved || s != sold {
		w.t.Errorf("keys of offer %d: %d available, %d reserved, %d sold; want %d, %d, %d", offerID, a, r, s, available, reserved, sold)
	}
}

// dbStatus is the status stored in the orders table, which can lag behind
// the status shown to the buyer until the cleanup has run.
func (w *world) dbStatus(orderID string) string {
	w.t.Helper()
	var status string
	if err := w.pool.QueryRow(context.Background(), "SELECT status FROM orders WHERE id = $1", orderID).Scan(&status); err != nil {
		w.t.Fatalf("reading the order's status: %v", err)
	}
	return status
}

// expireIn sets an order's deadline to a moment shortly from now, by the
// database's clock.
func (w *world) expireIn(orderID string, milliseconds int) {
	w.t.Helper()
	_, err := w.pool.Exec(context.Background(),
		"UPDATE orders SET expires_at = now() + make_interval(secs => $2::float8 / 1000) WHERE id = $1", orderID, milliseconds)
	if err != nil {
		w.t.Fatalf("setting the order's deadline: %v", err)
	}
}

// expire moves an order's deadline into the past, as if 15 minutes had gone by.
func (w *world) expire(orderID string) {
	w.t.Helper()
	_, err := w.pool.Exec(context.Background(), "UPDATE orders SET expires_at = now() - interval '1 second' WHERE id = $1", orderID)
	if err != nil {
		w.t.Fatalf("expiring the order: %v", err)
	}
}

func (w *world) unitsSold() int {
	w.t.Helper()
	var units int
	err := w.pool.QueryRow(context.Background(),
		"SELECT COALESCE((SELECT units_sold FROM product_sales WHERE product_id = $1), 0)", w.productID).Scan(&units)
	if err != nil {
		w.t.Fatalf("reading units sold: %v", err)
	}
	return units
}

var ctx = context.Background()

// ---- The ordinary paths ----

func TestCreateOrderReservesKeys(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 5)
	buyer := w.users(1)[0]

	order, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2*1999)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	if order.Status != StatusPending || order.TotalCents != 3998 {
		t.Errorf("status %q, total %d; want pending, 3998", order.Status, order.TotalCents)
	}
	if len(order.Items) != 1 || order.Items[0].Quantity != 2 || order.Items[0].UnitPriceCents != 1999 {
		t.Errorf("items = %+v", order.Items)
	}
	// An unpaid order must not show its keys, even though they are reserved.
	if len(order.Items[0].Keys) != 0 {
		t.Errorf("a pending order exposes keys: %v", order.Items[0].Keys)
	}
	w.wantKeys(offer, 3, 2, 0)
}

func TestCreateOrderFailuresReserveNothing(t *testing.T) {
	w := newWorld(t)
	plenty := w.offer(1000, 5)
	scarce := w.offer(2000, 1)
	buyer := w.users(1)[0]

	t.Run("the total differs from the one the buyer saw", func(t *testing.T) {
		_, err := w.store.CreateOrder(ctx, buyer, []Line{{plenty, 1}}, 999)
		if !errors.Is(err, ErrPriceChanged) {
			t.Fatalf("error = %v, want ErrPriceChanged", err)
		}
	})

	t.Run("an offer that does not exist", func(t *testing.T) {
		_, err := w.store.CreateOrder(ctx, buyer, []Line{{plenty, 1}, {999999999, 1}}, 1000)
		if !errors.Is(err, ErrOfferNotFound) {
			t.Fatalf("error = %v, want ErrOfferNotFound", err)
		}
	})

	t.Run("one line has too little stock, so no line is reserved", func(t *testing.T) {
		_, err := w.store.CreateOrder(ctx, buyer, []Line{{plenty, 3}, {scarce, 2}}, 3*1000+2*2000)
		var stock *StockError
		if !errors.As(err, &stock) {
			t.Fatalf("error = %v, want a StockError", err)
		}
		if stock.OfferID != scarce || stock.Requested != 2 || stock.Available != 1 {
			t.Errorf("StockError = %+v", stock)
		}
	})

	t.Run("an empty cart", func(t *testing.T) {
		_, err := w.store.CreateOrder(ctx, buyer, nil, 0)
		var validation ValidationError
		if !errors.As(err, &validation) {
			t.Fatalf("error = %v, want a ValidationError", err)
		}
	})

	// After all four failures: no key moved and no order exists.
	w.wantKeys(plenty, 5, 0, 0)
	w.wantKeys(scarce, 1, 0, 0)
	orders, err := w.store.ListOrders(ctx, buyer)
	if err != nil || len(orders) != 0 {
		t.Errorf("orders after failed attempts: %v, %v", orders, err)
	}
}

func TestPendingOrderLimit(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(500, 10)
	buyer := w.users(1)[0]

	for i := 1; i <= MaxPendingOrders; i++ {
		if _, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 500); err != nil {
			t.Fatalf("order %d: %v", i, err)
		}
	}
	_, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 500)
	if !errors.Is(err, ErrTooManyPending) {
		t.Errorf("order %d: error = %v, want ErrTooManyPending", MaxPendingOrders+1, err)
	}
	w.wantKeys(offer, 10-MaxPendingOrders, MaxPendingOrders, 0)
}

func TestPayApprovedDeliversKeys(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 3)
	buyer := w.users(1)[0]
	order, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 3998)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	paid, err := w.store.Pay(ctx, buyer, order.ID, true)
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	if paid.Status != StatusPaid || paid.PaidAt == nil {
		t.Errorf("status %q, paid_at %v", paid.Status, paid.PaidAt)
	}
	if len(paid.Items[0].Keys) != 2 || paid.Items[0].Keys[0] == paid.Items[0].Keys[1] {
		t.Errorf("keys = %v, want two different keys", paid.Items[0].Keys)
	}
	w.wantKeys(offer, 1, 0, 2)
	if got := w.unitsSold(); got != 2 {
		t.Errorf("units sold = %d, want 2", got)
	}

	// Paying again changes nothing: same keys, sale not counted twice.
	again, err := w.store.Pay(ctx, buyer, order.ID, true)
	if err != nil {
		t.Fatalf("second Pay: %v", err)
	}
	if fmt.Sprint(again.Items[0].Keys) != fmt.Sprint(paid.Items[0].Keys) {
		t.Errorf("the keys changed on a second payment")
	}
	// Not even a "decline" can undo a paid order.
	if after, err := w.store.Pay(ctx, buyer, order.ID, false); err != nil || after.Status != StatusPaid {
		t.Errorf("declining a paid order: status %q, error %v", after.Status, err)
	}
	w.wantKeys(offer, 1, 0, 2)
	if got := w.unitsSold(); got != 2 {
		t.Errorf("units sold after paying again = %d, want 2", got)
	}
}

func TestPayDeclinedReleasesKeys(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 3)
	buyer := w.users(1)[0]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 3998)

	failed, err := w.store.Pay(ctx, buyer, order.ID, false)
	if err != nil {
		t.Fatalf("Pay(decline): %v", err)
	}
	if failed.Status != StatusFailed || len(failed.Items[0].Keys) != 0 {
		t.Errorf("status %q, keys %v", failed.Status, failed.Items[0].Keys)
	}
	w.wantKeys(offer, 3, 0, 0)
	if got := w.unitsSold(); got != 0 {
		t.Errorf("units sold = %d, want 0", got)
	}

	// A failed order stays failed, even if a payment is approved afterwards.
	after, err := w.store.Pay(ctx, buyer, order.ID, true)
	if !errors.Is(err, ErrOrderNotPayable) || after.Status != StatusFailed {
		t.Errorf("paying a failed order: status %q, error %v", after.Status, err)
	}
	w.wantKeys(offer, 3, 0, 0)
}

func TestOrdersBelongToTheirOwner(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 3)
	users := w.users(2)
	owner, stranger := users[0], users[1]
	order, _ := w.store.CreateOrder(ctx, owner, []Line{{offer, 1}}, 1999)
	w.store.Pay(ctx, owner, order.ID, true)

	if _, err := w.store.GetOrder(ctx, stranger, order.ID); !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("a stranger reading the order: error = %v, want ErrOrderNotFound", err)
	}
	if _, err := w.store.Pay(ctx, stranger, order.ID, false); !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("a stranger paying the order: error = %v, want ErrOrderNotFound", err)
	}
	if list, _ := w.store.ListOrders(ctx, stranger); len(list) != 0 {
		t.Errorf("a stranger's order list has %d orders", len(list))
	}
	for _, id := range []string{"", "nonsense", "00000000-0000-0000-0000-000000000000"} {
		if _, err := w.store.GetOrder(ctx, owner, id); !errors.Is(err, ErrOrderNotFound) {
			t.Errorf("GetOrder(%q): error = %v, want ErrOrderNotFound", id, err)
		}
	}

	// The owner still sees it, with the key.
	mine, err := w.store.GetOrder(ctx, owner, order.ID)
	if err != nil || len(mine.Items[0].Keys) != 1 {
		t.Errorf("the owner's view: %+v, %v", mine, err)
	}
	if list, _ := w.store.ListOrders(ctx, owner); len(list) != 1 || list[0].ItemCount != 1 {
		t.Errorf("the owner's order list: %+v", list)
	}
}

func TestExpiredOrderReleasesKeys(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 1)
	users := w.users(2)
	first, second := users[0], users[1]

	order, _ := w.store.CreateOrder(ctx, first, []Line{{offer, 1}}, 1999)
	w.wantKeys(offer, 0, 1, 0)

	// While the first order holds the only key, a second buyer is refused.
	_, err := w.store.CreateOrder(ctx, second, []Line{{offer, 1}}, 1999)
	var stock *StockError
	if !errors.As(err, &stock) {
		t.Fatalf("second buyer while reserved: error = %v, want a StockError", err)
	}

	w.expire(order.ID)

	// The buyer sees "expired" at once, before any cleanup has run...
	shown, _ := w.store.GetOrder(ctx, first, order.ID)
	if shown.Status != StatusExpired {
		t.Errorf("shown status = %q, want expired", shown.Status)
	}
	if got := w.dbStatus(order.ID); got != StatusPending {
		t.Errorf("stored status before cleanup = %q, want pending", got)
	}

	// ...and the next order by anyone runs the cleanup, which frees the key
	// for that very order.
	won, err := w.store.CreateOrder(ctx, second, []Line{{offer, 1}}, 1999)
	if err != nil {
		t.Fatalf("second buyer after expiry: %v", err)
	}
	if got := w.dbStatus(order.ID); got != StatusExpired {
		t.Errorf("stored status after cleanup = %q, want expired", got)
	}
	w.wantKeys(offer, 0, 1, 0)

	// The expired order cannot be paid any more, and trying does not disturb
	// the key that now belongs to the second buyer.
	if after, err := w.store.Pay(ctx, first, order.ID, true); !errors.Is(err, ErrOrderNotPayable) || after.Status != StatusExpired {
		t.Errorf("paying an expired order: status %q, error %v", after.Status, err)
	}
	paid, err := w.store.Pay(ctx, second, won.ID, true)
	if err != nil || len(paid.Items[0].Keys) != 1 {
		t.Fatalf("the second buyer paying: %+v, %v", paid, err)
	}
	w.wantKeys(offer, 0, 0, 1)
}

func TestPayAfterTheDeadlineIsRefused(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1999, 2)
	buyer := w.users(1)[0]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 3998)
	w.expire(order.ID)

	after, err := w.store.Pay(ctx, buyer, order.ID, true)
	if !errors.Is(err, ErrOrderNotPayable) {
		t.Fatalf("error = %v, want ErrOrderNotPayable", err)
	}
	if after.Status != StatusExpired || w.dbStatus(order.ID) != StatusExpired {
		t.Errorf("status shown %q, stored %q; want expired", after.Status, w.dbStatus(order.ID))
	}
	w.wantKeys(offer, 2, 0, 0)
	if got := w.unitsSold(); got != 0 {
		t.Errorf("units sold = %d, want 0", got)
	}
}

func TestPriceCart(t *testing.T) {
	w := newWorld(t)
	cheap := w.offer(500, 2)
	empty := w.offer(900, 0)

	cart, err := w.store.PriceCart(ctx, []Line{{empty, 1}, {cheap, 2}, {999999999, 1}})
	if err != nil {
		t.Fatalf("PriceCart: %v", err)
	}
	if len(cart.Lines) != 2 || cart.TotalCents != 2*500+900 {
		t.Fatalf("lines %d, total %d", len(cart.Lines), cart.TotalCents)
	}
	if len(cart.MissingOfferIDs) != 1 || cart.MissingOfferIDs[0] != 999999999 {
		t.Errorf("missing = %v", cart.MissingOfferIDs)
	}
	if cart.CanCheckout {
		t.Error("a cart with a sold-out line and a missing offer can check out")
	}
	for _, line := range cart.Lines {
		wantAvailable := line.OfferID == cheap
		if line.Available != wantAvailable {
			t.Errorf("offer %d: available = %v, want %v", line.OfferID, line.Available, wantAvailable)
		}
	}
	// Pricing a cart reserves nothing.
	w.wantKeys(cheap, 2, 0, 0)
}
