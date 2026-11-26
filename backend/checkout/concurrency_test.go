package checkout

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// These tests are the evidence that a key cannot be sold twice. Each one
// starts many goroutines that all wait at a closed gate, opens the gate so
// they run at the same moment against the real database, and then checks
// what is in the tables.
//
// They check the tables, not only the return values. A bug could make every
// call report success or failure plausibly while the rows say otherwise.

// result is what one buyer got from CreateOrder.
type result struct {
	order Order
	err   error
}

// race makes every buyer place the same order at the same instant and
// returns what each one got. It also reports the largest number of calls
// that were in flight together, to show the transactions really overlapped.
func race(w *world, buyers []int64, lines []Line, total int) ([]result, int) {
	w.warmUp() // every connection open, so nobody gets a head start
	results := make([]result, len(buyers))
	gate := make(chan struct{})
	var inFlight, peak atomic.Int32
	var wg sync.WaitGroup

	for i, buyer := range buyers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate // every goroutine blocks here until the gate is closed

			now := inFlight.Add(1)
			for {
				old := peak.Load()
				if now <= old || peak.CompareAndSwap(old, now) {
					break
				}
			}
			order, err := w.store.CreateOrder(ctx, buyer, lines, total)
			inFlight.Add(-1)

			results[i] = result{order, err}
		}()
	}
	close(gate) // release them all at once
	wg.Wait()
	return results, int(peak.Load())
}

// tally counts successes and "not enough stock" answers, and fails the test
// on any other error.
func tally(t *testing.T, results []result) (succeeded []Order, outOfStock int) {
	t.Helper()
	for _, r := range results {
		var stock *StockError
		switch {
		case r.err == nil:
			succeeded = append(succeeded, r.order)
		case errors.As(r.err, &stock):
			outOfStock++
		default:
			t.Errorf("unexpected error: %v", r.err)
		}
	}
	return succeeded, outOfStock
}

// itemsShortOfKeys returns how many live order items hold a different number
// of keys than their quantity. A key row has a single order_item_id, so the database
// cannot literally store two owners. What a broken reservation does instead
// is hand the same key to two orders one after the other, so that an order
// believes it holds keys it no longer does. That shows up as an order item
// holding fewer keys than its quantity.
func (w *world) itemsShortOfKeys(offerID int64) int {
	w.t.Helper()
	var short int
	err := w.pool.QueryRow(ctx, `
		SELECT COUNT(*)
		  FROM order_items oi
		  JOIN orders o ON o.id = oi.order_id
		 WHERE oi.offer_id = $1
		   AND o.status IN ('pending', 'paid')
		   AND oi.quantity <> (SELECT COUNT(*) FROM keys k WHERE k.order_item_id = oi.id)`, offerID,
	).Scan(&short)
	if err != nil {
		w.t.Fatalf("checking order items against their keys: %v", err)
	}
	return short
}

// Test 1. Twenty buyers want the last key at the same instant.
func TestConcurrentBuyersOfTheLastKey(t *testing.T) {
	const buyers = 20
	w := newWorld(t)
	offer := w.offer(1999, 1)
	users := w.users(buyers)

	results, peak := race(w, users, []Line{{offer, 1}}, 1999)
	succeeded, outOfStock := tally(t, results)
	t.Logf("%d buyers, up to %d calls in flight at once on a pool of %d connections: %d got the key, %d were told it is out of stock", buyers, peak, poolSize, len(succeeded), outOfStock)

	if len(succeeded) != 1 {
		t.Errorf("%d orders were created for 1 key, want exactly 1", len(succeeded))
	}
	if outOfStock != buyers-1 {
		t.Errorf("%d buyers were told out of stock, want %d", outOfStock, buyers-1)
	}
	// What the tables say, whatever the calls returned.
	w.wantKeys(offer, 0, 1, 0)
	if short := w.itemsShortOfKeys(offer); short != 0 {
		t.Errorf("%d order items hold fewer keys than they were sold", short)
	}

	// Paying the winning order delivers exactly that one key.
	if len(succeeded) == 1 {
		winner := succeeded[0]
		var owner int64
		w.pool.QueryRow(ctx, "SELECT user_id FROM orders WHERE id = $1", winner.ID).Scan(&owner)
		paid, err := w.store.Pay(ctx, owner, winner.ID, true)
		if err != nil || len(paid.Items[0].Keys) != 1 {
			t.Errorf("paying the winning order: %+v, %v", paid.Items, err)
		}
		w.wantKeys(offer, 0, 0, 1)
	}
}

// Test 2. Five keys, twenty buyers who each want two. Only two orders can be
// filled (four keys). The fifth key must stay available: nobody may end up
// with half of what they asked for.
func TestConcurrentBuyersWithMoreDemandThanStock(t *testing.T) {
	const buyers, stock, each = 20, 5, 2
	w := newWorld(t)
	offer := w.offer(1000, stock)
	users := w.users(buyers)

	results, peak := race(w, users, []Line{{offer, each}}, each*1000)
	succeeded, outOfStock := tally(t, results)
	t.Logf("%d buyers wanting %d each from %d keys, up to %d calls in flight at once on a pool of %d connections: %d orders filled, %d refused", buyers, each, stock, peak, poolSize, len(succeeded), outOfStock)

	if want := stock / each; len(succeeded) != want {
		t.Errorf("%d orders were created, want exactly %d", len(succeeded), want)
	}
	if len(succeeded)+outOfStock != buyers {
		t.Errorf("%d answers accounted for out of %d", len(succeeded)+outOfStock, buyers)
	}

	// Every order that succeeded holds exactly its quantity, no key is held
	// by a refused buyer, and the totals add up.
	if short := w.itemsShortOfKeys(offer); short != 0 {
		t.Errorf("%d order items hold a different number of keys than their quantity", short)
	}
	available, reserved, sold := w.keyCounts(offer)
	if reserved != len(succeeded)*each || available+reserved+sold != stock || sold != 0 {
		t.Errorf("keys: %d available, %d reserved, %d sold; with %d orders of %d each out of %d keys", available, reserved, sold, len(succeeded), each, stock)
	}

	// No order of a refused buyer exists at all.
	var orders int
	w.pool.QueryRow(ctx, "SELECT COUNT(*) FROM orders WHERE user_id = ANY($1::bigint[])", users).Scan(&orders)
	if orders != len(succeeded) {
		t.Errorf("%d orders exist, but %d buyers succeeded", orders, len(succeeded))
	}
}

// Test 3. One order, paid ten times at the same instant: a double click, two
// tabs, a webhook delivered twice.
func TestConcurrentPaymentsOfOneOrder(t *testing.T) {
	const payments, quantity = 10, 3
	w := newWorld(t)
	offer := w.offer(1000, 5)
	buyer := w.users(1)[0]
	order, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, quantity}}, quantity*1000)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	w.warmUp()
	gate := make(chan struct{})
	errs := make([]error, payments)
	keys := make([][]string, payments)
	var wg sync.WaitGroup
	for i := range payments {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			paid, err := w.store.Pay(ctx, buyer, order.ID, true)
			errs[i] = err
			if err == nil {
				keys[i] = paid.Items[0].Keys
			}
		}()
	}
	close(gate)
	wg.Wait()

	// Every call reports the paid order, with the same keys each time.
	for i := range payments {
		if errs[i] != nil {
			t.Errorf("payment %d: %v", i, errs[i])
			continue
		}
		if len(keys[i]) != quantity || keys[i][0] != keys[0][0] {
			t.Errorf("payment %d returned keys %v, the first returned %v", i, keys[i], keys[0])
		}
	}
	// The keys were sold once and the sale was counted once.
	w.wantKeys(offer, 5-quantity, 0, quantity)
	if got := w.unitsSold(); got != quantity {
		t.Errorf("units sold = %d after %d simultaneous payments, want %d", got, payments, quantity)
	}
}

// Test 4. A payment and the cleanup collide at an order's deadline. Each
// round sets the deadline 300 ms ahead, then starts one payment and several
// other buyers' orders (each of which runs the cleanup first), all timed to
// land around that deadline. The payment's delay differs per round, so some
// rounds it arrives just before the deadline and some just after.
//
// Either outcome is acceptable, a mixture is not:
//
//	paid     and its keys are sold and still attached to it, or
//	expired  and its keys are available again
//
// What must never happen is a paid order whose keys were released under it,
// or an expired order that still holds keys.
func TestPaymentRacingAgainstExpiry(t *testing.T) {
	const rounds, others, deadlineMs = 12, 6, 300
	paidCount, expiredCount := 0, 0

	for round := range rounds {
		w := newWorld(t)
		offer := w.offer(1000, 2)
		// A second offer for the other buyers, so their orders succeed and
		// the cleanup at the start of each one really runs.
		bystander := w.offer(700, others)
		buyer := w.users(1)[0]
		otherBuyers := w.users(others)

		order, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2000)
		if err != nil {
			t.Fatalf("round %d: CreateOrder: %v", round, err)
		}
		w.expireIn(order.ID, deadlineMs)

		// The payment starts between 100 ms and 430 ms after now, so it
		// straddles the deadline across the rounds.
		payDelay := time.Duration(100+round*30) * time.Millisecond

		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(payDelay)
			w.store.Pay(ctx, buyer, order.ID, true)
		}()
		for i, other := range otherBuyers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				// The other buyers are spread from 200 ms to 450 ms.
				time.Sleep(time.Duration(200+i*50) * time.Millisecond)
				w.store.CreateOrder(ctx, other, []Line{{bystander, 1}}, 700)
			}()
		}
		wg.Wait()
		time.Sleep(deadlineMs * time.Millisecond) // make sure the deadline is behind us

		// One more cleanup, so the stored state is final whichever way it went.
		if _, err := w.pool.Exec(ctx, releaseExpiredSQL); err != nil {
			t.Fatalf("round %d: cleanup: %v", round, err)
		}

		final, err := w.store.GetOrder(ctx, buyer, order.ID)
		if err != nil {
			t.Fatalf("round %d: GetOrder: %v", round, err)
		}
		available, reserved, sold := w.keyCounts(offer)
		switch final.Status {
		case StatusPaid:
			paidCount++
			if available != 0 || reserved != 0 || sold != 2 || len(final.Items[0].Keys) != 2 {
				t.Errorf("round %d: paid, but keys are %d available, %d reserved, %d sold, %d shown", round, available, reserved, sold, len(final.Items[0].Keys))
			}
		case StatusExpired:
			expiredCount++
			if available != 2 || reserved != 0 || sold != 0 {
				t.Errorf("round %d: expired, but keys are %d available, %d reserved, %d sold", round, available, reserved, sold)
			}
			if len(final.Items[0].Keys) != 0 {
				t.Errorf("round %d: an expired order shows keys", round)
			}
		default:
			t.Errorf("round %d: final status %q", round, final.Status)
		}
		if stored := w.dbStatus(order.ID); stored != final.Status {
			t.Errorf("round %d: stored status %q differs from the shown %q", round, stored, final.Status)
		}
		if short := w.itemsShortOfKeys(offer); short != 0 {
			t.Errorf("round %d: %d order items hold a different number of keys than their quantity", round, short)
		}
		w.cleanup()
	}

	t.Logf("%d rounds: %d ended paid, %d ended expired, none mixed", rounds, paidCount, expiredCount)
	if paidCount == 0 || expiredCount == 0 {
		// Not a failure of the code, but the test did not see both sides of
		// the race, so it proved less than it is meant to.
		t.Logf("note: only one outcome occurred in this run; the timing did not straddle the deadline")
	}
}
