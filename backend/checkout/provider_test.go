package checkout

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
)

// fakeGateway stands in for Stripe. It records what it was asked to do and
// answers with made-up ids, so the tests can check exactly how many payment
// pages and refunds the code asked for.
type fakeGateway struct {
	mu        sync.Mutex
	sessions  []SessionRequest
	refunds   []string // payment ids that were refunded
	failNext  bool     // make the next CreateSession fail
	sessionNo int
}

func (g *fakeGateway) CreateSession(_ context.Context, request SessionRequest) (Session, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.failNext {
		g.failNext = false
		return Session{}, errors.New("the provider is down")
	}
	g.sessions = append(g.sessions, request)
	g.sessionNo++
	id := fmt.Sprintf("cs_fake_%s_%d", request.OrderID[:8], g.sessionNo)
	return Session{ID: id, URL: "https://pay.example/" + id}, nil
}

func (g *fakeGateway) Refund(_ context.Context, paymentID, _ string) (string, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.refunds = append(g.refunds, paymentID)
	return "re_fake_" + paymentID, nil
}

// providerWorld is a test world whose store pays through the fake gateway.
func providerWorld(t *testing.T) (*world, *fakeGateway) {
	w := newWorld(t)
	gateway := &fakeGateway{}
	w.store.Gateway = gateway
	return w, gateway
}

// sessionOf returns the id of the payment page stored for an order.
func (w *world) sessionOf(orderID string) string {
	w.t.Helper()
	var session string
	if err := w.pool.QueryRow(ctx, "SELECT COALESCE(payment_ref, '') FROM orders WHERE id = $1", orderID).Scan(&session); err != nil {
		w.t.Fatalf("reading the order's payment page: %v", err)
	}
	return session
}

// paymentFor builds the provider's "paid" message for an order, as if the
// buyer had paid on the order's own payment page.
func (w *world) paymentFor(order Order) ProviderPayment {
	return ProviderPayment{
		OrderID:     order.ID,
		SessionID:   w.sessionOf(order.ID),
		PaymentID:   "pi_" + order.ID[:8],
		AmountCents: order.TotalCents,
		Currency:    Currency,
	}
}

func TestStartPaymentCreatesOnePagePerOrder(t *testing.T) {
	w, gateway := providerWorld(t)
	offer := w.offer(1999, 3)
	users := w.users(2)
	buyer, stranger := users[0], users[1]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 3998)

	url, err := w.store.StartPayment(ctx, buyer, order.ID)
	if err != nil {
		t.Fatalf("StartPayment: %v", err)
	}
	if len(gateway.sessions) != 1 {
		t.Fatalf("%d payment pages were created, want 1", len(gateway.sessions))
	}
	sent := gateway.sessions[0]
	if sent.OrderID != order.ID || len(sent.Items) != 1 || sent.Items[0].UnitPriceCents != 1999 || sent.Items[0].Quantity != 2 {
		t.Errorf("the provider was sent %+v", sent)
	}

	// Asking again, as after a double click or coming back from the page,
	// gives the same page and creates no second one.
	again, err := w.store.StartPayment(ctx, buyer, order.ID)
	if err != nil || again != url || len(gateway.sessions) != 1 {
		t.Errorf("second StartPayment: url %q (first was %q), %d pages, error %v", again, url, len(gateway.sessions), err)
	}

	// Going to pay stretches the reservation so it outlasts the payment page.
	after, _ := w.store.GetOrder(ctx, buyer, order.ID)
	if after.SecondsLeft < (paymentHoldMinutes-1)*60 || after.SecondsLeft > paymentHoldMinutes*60 {
		t.Errorf("seconds left after starting payment = %d, want about %d", after.SecondsLeft, paymentHoldMinutes*60)
	}
	w.wantKeys(offer, 1, 2, 0)

	// Someone else cannot start a payment for this order.
	if _, err := w.store.StartPayment(ctx, stranger, order.ID); !errors.Is(err, ErrOrderNotFound) {
		t.Errorf("a stranger starting the payment: error = %v, want ErrOrderNotFound", err)
	}
}

func TestStartPaymentFailures(t *testing.T) {
	w, gateway := providerWorld(t)
	offer := w.offer(1000, 5)
	buyer := w.users(1)[0]

	t.Run("the provider is down: nothing is saved and it can be tried again", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1000)
		gateway.failNext = true
		if _, err := w.store.StartPayment(ctx, buyer, order.ID); err == nil {
			t.Fatal("expected an error")
		}
		if got := w.sessionOf(order.ID); got != "" {
			t.Errorf("a payment page %q was saved although creating it failed", got)
		}
		if _, err := w.store.StartPayment(ctx, buyer, order.ID); err != nil {
			t.Errorf("trying again: %v", err)
		}
	})

	t.Run("an expired order cannot go to payment", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1000)
		w.expire(order.ID)
		pages := len(gateway.sessions)
		if _, err := w.store.StartPayment(ctx, buyer, order.ID); !errors.Is(err, ErrOrderNotPayable) {
			t.Errorf("error = %v, want ErrOrderNotPayable", err)
		}
		if len(gateway.sessions) != pages {
			t.Error("a payment page was created for an expired order")
		}
	})

	t.Run("without a provider configured", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1000)
		w.store.Gateway = nil
		defer func() { w.store.Gateway = gateway }()
		if _, err := w.store.StartPayment(ctx, buyer, order.ID); !errors.Is(err, ErrNoGateway) {
			t.Errorf("error = %v, want ErrNoGateway", err)
		}
	})
}

func TestProviderPaidFulfilsOnce(t *testing.T) {
	w, gateway := providerWorld(t)
	offer := w.offer(1999, 3)
	buyer := w.users(1)[0]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 3998)
	w.store.StartPayment(ctx, buyer, order.ID)
	payment := w.paymentFor(order)

	outcome, err := w.store.ProviderPaid(ctx, payment)
	if err != nil || outcome != OutcomeFulfilled {
		t.Fatalf("first message: outcome %q, error %v", outcome, err)
	}
	paid, _ := w.store.GetOrder(ctx, buyer, order.ID)
	if paid.Status != StatusPaid || len(paid.Items[0].Keys) != 2 || paid.Refunded {
		t.Errorf("after payment: status %q, %d keys, refunded %v", paid.Status, len(paid.Items[0].Keys), paid.Refunded)
	}

	// Stripe sends the same message again. Nothing changes.
	for range 3 {
		outcome, err := w.store.ProviderPaid(ctx, payment)
		if err != nil || outcome != OutcomeAlreadyPaid {
			t.Errorf("repeated message: outcome %q, error %v", outcome, err)
		}
	}
	w.wantKeys(offer, 1, 0, 2)
	if got := w.unitsSold(); got != 2 {
		t.Errorf("units sold = %d, want 2", got)
	}
	if len(gateway.refunds) != 0 {
		t.Errorf("refunds were issued: %v", gateway.refunds)
	}
}

// The same message delivered ten times at once must fulfil the order once.
func TestProviderPaidConcurrentDeliveries(t *testing.T) {
	w, gateway := providerWorld(t)
	offer := w.offer(1000, 5)
	buyer := w.users(1)[0]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 3}}, 3000)
	w.store.StartPayment(ctx, buyer, order.ID)
	payment := w.paymentFor(order)

	w.warmUp()
	gate := make(chan struct{})
	outcomes := make([]string, 10)
	var wg sync.WaitGroup
	for i := range outcomes {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-gate
			outcomes[i], _ = w.store.ProviderPaid(ctx, payment)
		}()
	}
	close(gate)
	wg.Wait()

	fulfilled := 0
	for _, outcome := range outcomes {
		switch outcome {
		case OutcomeFulfilled:
			fulfilled++
		case OutcomeAlreadyPaid:
		default:
			t.Errorf("unexpected outcome %q", outcome)
		}
	}
	if fulfilled != 1 {
		t.Errorf("%d deliveries fulfilled the order, want exactly 1", fulfilled)
	}
	w.wantKeys(offer, 2, 0, 3)
	if got := w.unitsSold(); got != 3 {
		t.Errorf("units sold = %d, want 3", got)
	}
	if len(gateway.refunds) != 0 {
		t.Errorf("refunds were issued: %v", gateway.refunds)
	}
}

// Money arrives for an order that has expired. The keys may already belong
// to someone else, so the order is not fulfilled and the payment goes back.
func TestProviderPaidAfterExpiryRefunds(t *testing.T) {
	w, gateway := providerWorld(t)
	offer := w.offer(1999, 1)
	buyer := w.users(1)[0]
	order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1999)
	w.store.StartPayment(ctx, buyer, order.ID)
	payment := w.paymentFor(order)
	w.expire(order.ID)

	outcome, err := w.store.ProviderPaid(ctx, payment)
	if err != nil || outcome != OutcomeRefunded {
		t.Fatalf("outcome %q, error %v; want refunded", outcome, err)
	}
	after, _ := w.store.GetOrder(ctx, buyer, order.ID)
	if after.Status != StatusExpired || !after.Refunded || len(after.Items[0].Keys) != 0 {
		t.Errorf("after the late payment: status %q, refunded %v, keys %v", after.Status, after.Refunded, after.Items[0].Keys)
	}
	w.wantKeys(offer, 1, 0, 0)
	if got := w.unitsSold(); got != 0 {
		t.Errorf("units sold = %d, want 0", got)
	}

	// The message comes again. The money must not be sent back twice.
	if outcome, err := w.store.ProviderPaid(ctx, payment); err != nil || outcome != OutcomeRefunded {
		t.Errorf("repeated message: outcome %q, error %v", outcome, err)
	}
	if len(gateway.refunds) != 1 || gateway.refunds[0] != payment.PaymentID {
		t.Errorf("refunds = %v, want exactly one for %s", gateway.refunds, payment.PaymentID)
	}
}

// A payment that does not match the order is never turned into keys.
func TestProviderPaidRefusesAMismatch(t *testing.T) {
	tests := []struct {
		name   string
		change func(p *ProviderPayment)
	}{
		{"a different payment page", func(p *ProviderPayment) { p.SessionID = "cs_someone_elses" }},
		{"a smaller amount", func(p *ProviderPayment) { p.AmountCents = 1 }},
		{"a different currency", func(p *ProviderPayment) { p.Currency = "usd" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, gateway := providerWorld(t)
			offer := w.offer(1999, 2)
			buyer := w.users(1)[0]
			order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1999)
			w.store.StartPayment(ctx, buyer, order.ID)
			payment := w.paymentFor(order)
			tt.change(&payment)

			outcome, err := w.store.ProviderPaid(ctx, payment)
			if err != nil || outcome != OutcomeRefunded {
				t.Fatalf("outcome %q, error %v; want refunded", outcome, err)
			}
			after, _ := w.store.GetOrder(ctx, buyer, order.ID)
			if after.Status == StatusPaid || len(after.Items[0].Keys) != 0 {
				t.Errorf("the order was fulfilled by a payment that does not match it: %+v", after)
			}
			if len(gateway.refunds) != 1 {
				t.Errorf("refunds = %v, want one", gateway.refunds)
			}
			w.wantKeys(offer, 1, 1, 0) // the order is still pending and holds its key
		})
	}
}

func TestProviderPaidForUnknownOrders(t *testing.T) {
	w, gateway := providerWorld(t)
	for _, id := range []string{"", "not-an-order", "00000000-0000-0000-0000-000000000000"} {
		outcome, err := w.store.ProviderPaid(ctx, ProviderPayment{OrderID: id, SessionID: "cs_x", PaymentID: "pi_x", AmountCents: 100, Currency: Currency})
		if err != nil || outcome != OutcomeIgnored {
			t.Errorf("order %q: outcome %q, error %v; want ignored", id, outcome, err)
		}
	}
	if len(gateway.refunds) != 0 {
		t.Errorf("refunds = %v", gateway.refunds)
	}
}

func TestProviderEndedReleasesKeys(t *testing.T) {
	w, _ := providerWorld(t)
	offer := w.offer(1000, 4)
	buyer := w.users(1)[0]

	t.Run("the payment page ran out of time", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 2}}, 2000)
		w.store.StartPayment(ctx, buyer, order.ID)
		if err := w.store.ProviderEnded(ctx, order.ID, w.sessionOf(order.ID), StatusExpired); err != nil {
			t.Fatalf("ProviderEnded: %v", err)
		}
		if got := w.dbStatus(order.ID); got != StatusExpired {
			t.Errorf("status = %q, want expired", got)
		}
		w.wantKeys(offer, 4, 0, 0)
	})

	t.Run("a message about another page changes nothing", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1000)
		w.store.StartPayment(ctx, buyer, order.ID)
		if err := w.store.ProviderEnded(ctx, order.ID, "cs_another_page", StatusExpired); err != nil {
			t.Fatalf("ProviderEnded: %v", err)
		}
		if got := w.dbStatus(order.ID); got != StatusPending {
			t.Errorf("status = %q, want pending", got)
		}
	})

	t.Run("a paid order stays paid", func(t *testing.T) {
		order, _ := w.store.CreateOrder(ctx, buyer, []Line{{offer, 1}}, 1000)
		w.store.StartPayment(ctx, buyer, order.ID)
		w.store.ProviderPaid(ctx, w.paymentFor(order))
		if err := w.store.ProviderEnded(ctx, order.ID, w.sessionOf(order.ID), StatusFailed); err != nil {
			t.Fatalf("ProviderEnded: %v", err)
		}
		if got := w.dbStatus(order.ID); got != StatusPaid {
			t.Errorf("status = %q, want paid", got)
		}
	})
}
