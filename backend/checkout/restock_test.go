package checkout

import (
	"regexp"
	"testing"
)

func TestNewKeyCode(t *testing.T) {
	shape := regexp.MustCompile(`^[A-HJ-NP-Z2-9]{5}-[A-HJ-NP-Z2-9]{5}-[A-HJ-NP-Z2-9]{5}$`)
	seen := map[string]bool{}
	for range 200 {
		code, err := newKeyCode()
		if err != nil {
			t.Fatalf("newKeyCode: %v", err)
		}
		if !shape.MatchString(code) {
			t.Fatalf("code %q does not look like XXXXX-XXXXX-XXXXX", code)
		}
		if seen[code] {
			t.Fatalf("code %q was generated twice", code)
		}
		seen[code] = true
	}
}

// restockingWorld is a test world whose store has the demo restock switched
// on. Every other test uses the default, where it is off, because they
// depend on exact stock numbers.
func restockingWorld(t *testing.T) *world {
	w := newWorld(t)
	w.store.RestockSold = true
	return w
}

func TestRestockTopsUpAnOfferThatRunsLow(t *testing.T) {
	w := restockingWorld(t)
	offer := w.offer(1000, 4)
	buyer := w.users(1)[0]

	order, err := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2000)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	// Reserving alone never restocks: nothing has been sold yet.
	w.wantKeys(offer, 2, 2, 0)

	paid, err := w.store.Pay(ctx, buyer, order.ID, true)
	if err != nil {
		t.Fatalf("Pay: %v", err)
	}
	// 4 keys, 2 sold, 2 left: below the threshold of 3, so 5 are added.
	w.wantKeys(offer, 2+restockBatch, 0, 2)

	// The buyer got the two keys that existed, not freshly made ones.
	for _, code := range paid.Items[0].Keys {
		if len(code) < 5 || code[:5] != "TEST-" {
			t.Errorf("the buyer received %q, which is not one of the offer's original keys", code)
		}
	}

	// Paying again must not restock a second time.
	if _, err := w.store.Pay(ctx, buyer, order.ID, true); err != nil {
		t.Fatalf("second Pay: %v", err)
	}
	w.wantKeys(offer, 2+restockBatch, 0, 2)
}

func TestRestockLeavesAWellStockedOfferAlone(t *testing.T) {
	w := restockingWorld(t)
	offer := w.offer(1000, 10)
	buyer := w.users(1)[0]

	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2000)
	if _, err := w.store.Pay(ctx, buyer, order.ID, true); err != nil {
		t.Fatalf("Pay: %v", err)
	}
	// 8 left, well above the threshold: stock simply went down by 2.
	w.wantKeys(offer, 8, 0, 2)
}

func TestRestockDoesNotHappenWithoutASale(t *testing.T) {
	w := restockingWorld(t)
	low := w.offer(1000, 1)
	empty := w.offer(2000, 0) // an offer that was sold out from the start
	buyer := w.users(1)[0]

	// A declined payment and an expired order sell nothing.
	declined, _ := w.store.CreateOrder(ctx, buyer, []Line{{low, 1}}, 1000)
	if _, err := w.store.Pay(ctx, buyer, declined.ID, false); err != nil {
		t.Fatalf("Pay(decline): %v", err)
	}
	expired, _ := w.store.CreateOrder(ctx, buyer, []Line{{low, 1}}, 1000)
	w.expire(expired.ID)
	w.store.Pay(ctx, buyer, expired.ID, true)

	w.wantKeys(low, 1, 0, 0)
	// Nothing can be sold from an empty offer, so it is never topped up and
	// stays sold out.
	w.wantKeys(empty, 0, 0, 0)
}

func TestRestockIsOffByDefault(t *testing.T) {
	w := newWorld(t)
	offer := w.offer(1000, 2)
	buyer := w.users(1)[0]

	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2000)
	if _, err := w.store.Pay(ctx, buyer, order.ID, true); err != nil {
		t.Fatalf("Pay: %v", err)
	}
	// Bought empty, and it stays empty.
	w.wantKeys(offer, 0, 0, 2)
}
