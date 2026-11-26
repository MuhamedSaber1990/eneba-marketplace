package checkout

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
)

// Paying through a payment provider (Stripe), as opposed to the simulated
// buttons. The provider hosts the page where the buyer pays, and tells us the
// result by calling our webhook. This file holds what happens on our side:
// starting a payment, and reacting to what the provider reports.
//
// Nothing here knows about Stripe itself. It talks to a Gateway, and
// backend/payments implements that for Stripe. The tests use a fake one.

// Gateway is the part of a payment provider this package needs.
type Gateway interface {
	// CreateSession opens a payment page for an order and returns its id and
	// address.
	CreateSession(ctx context.Context, request SessionRequest) (Session, error)
	// Refund sends a payment back and returns the refund's id.
	Refund(ctx context.Context, paymentID, orderID string) (string, error)
}

// SessionRequest is what the provider needs to build the payment page.
type SessionRequest struct {
	OrderID   string
	Items     []OrderItem
	ExpiresAt time.Time // when the payment page stops accepting payment
}

// Session is a payment page at the provider.
type Session struct {
	ID  string
	URL string
}

// A payment page at Stripe stays open for at least 30 minutes; a shorter
// time is refused. So when a buyer goes to pay, the order's reservation is
// stretched to outlast the page:
//
//	the page closes       sessionMinutes after the buyer goes to pay
//	the reservation ends  paymentHoldMinutes after the buyer goes to pay
//
// The minutes in between are for the provider's message to reach us. A buyer
// who never goes to pay still loses the reservation after the usual
// ReservationMinutes.
const (
	sessionMinutes     = 31 // one more than Stripe's minimum, in case our clocks differ
	paymentHoldMinutes = 36
)

// ErrNoGateway means a payment through the provider was asked for, but no
// provider is configured.
var ErrNoGateway = errors.New("no payment provider is configured")

// StartPayment returns the address of the provider's payment page for a
// pending order, creating the page the first time it is asked for.
func (s *Store) StartPayment(ctx context.Context, userID int64, orderID string) (string, error) {
	if s.Gateway == nil {
		return "", ErrNoGateway
	}
	// The items are needed for the payment page. Loading the order here also
	// answers "does this order exist and belong to this user".
	order, err := s.GetOrder(ctx, userID, orderID)
	if err != nil {
		return "", err
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if _, err := tx.Exec(ctx, releaseExpiredSQL); err != nil {
		return "", fmt.Errorf("releasing expired reservations: %w", err)
	}

	// Lock the order row. If the buyer double-clicks, the second request
	// waits here and then finds the page the first one created.
	var status string
	var overdue bool
	var existingURL *string // nil while the order has no payment page
	err = tx.QueryRow(ctx, `
		SELECT status, expires_at <= now(), payment_url
		  FROM orders
		 WHERE id = $1 AND user_id = $2
		   FOR UPDATE`, orderID, userID,
	).Scan(&status, &overdue, &existingURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrOrderNotFound
	}
	if err != nil {
		return "", fmt.Errorf("locking the order: %w", err)
	}

	if status != StatusPending || overdue {
		if status == StatusPending {
			if err := endOrder(ctx, tx, orderID, StatusExpired); err != nil {
				return "", err
			}
			if err := tx.Commit(ctx); err != nil {
				return "", fmt.Errorf("committing: %w", err)
			}
		}
		return "", ErrOrderNotPayable
	}

	// One page per order: send the buyer back to the one that exists.
	if existingURL != nil {
		return *existingURL, nil
	}

	session, err := s.Gateway.CreateSession(ctx, SessionRequest{
		OrderID:   orderID,
		Items:     order.Items,
		ExpiresAt: time.Now().Add(sessionMinutes * time.Minute),
	})
	if err != nil {
		return "", fmt.Errorf("creating the payment page: %w", err)
	}

	// Remember the page, and stretch the reservation to outlast it.
	_, err = tx.Exec(ctx, `
		UPDATE orders
		   SET payment_ref = $2, payment_url = $3,
		       expires_at = now() + make_interval(mins => $4::int)
		 WHERE id = $1`, orderID, session.ID, session.URL, paymentHoldMinutes)
	if err != nil {
		return "", fmt.Errorf("saving the payment page: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("committing: %w", err)
	}
	return session.URL, nil
}

// ProviderPayment is what the provider reports about a completed payment.
type ProviderPayment struct {
	OrderID     string // the order the payment page was created for
	SessionID   string // the payment page
	PaymentID   string // the payment itself, needed for a refund
	AmountCents int
	Currency    string
}

// What ProviderPaid did with a payment.
const (
	OutcomeFulfilled   = "fulfilled"    // the order is now paid and its keys sold
	OutcomeAlreadyPaid = "already paid" // this payment was handled before
	OutcomeRefunded    = "refunded"     // the order could not be fulfilled, the money went back
	OutcomeIgnored     = "ignored"      // no such order
)

// Currency is the one currency the shop sells in.
const Currency = "eur"

// ProviderPaid handles the provider's message that a payment succeeded. It
// is the provider's way into markPaid, the same fulfilment step the
// simulated payment uses.
//
// The provider may send the same message more than once, or late. So this
// function is safe to repeat, and it never fulfils an order it should not:
//
//   - the payment must be for this order's own payment page and for the
//     order's exact amount and currency
//   - the order must still be pending and within its deadline
//
// If money arrived but the order cannot be fulfilled, the payment is
// refunded. Taking money and delivering nothing is the one outcome that must
// not happen.
func (s *Store) ProviderPaid(ctx context.Context, payment ProviderPayment) (string, error) {
	if !looksLikeUUID(payment.OrderID) {
		return OutcomeIgnored, nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return "", fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// No cleanup first, unlike the other entry points: if this very order
	// has just passed its deadline, the decision about it is made below,
	// under the lock, together with the refund.
	var status string
	var overdue bool
	var totalCents int
	var sessionID, refundRef *string
	err = tx.QueryRow(ctx, `
		SELECT status, expires_at <= now(), total_cents, payment_ref, refund_ref
		  FROM orders
		 WHERE id = $1
		   FOR UPDATE`, payment.OrderID,
	).Scan(&status, &overdue, &totalCents, &sessionID, &refundRef)
	if errors.Is(err, pgx.ErrNoRows) {
		return OutcomeIgnored, nil
	}
	if err != nil {
		return "", fmt.Errorf("locking the order: %w", err)
	}

	matches := sessionID != nil && *sessionID == payment.SessionID &&
		payment.AmountCents == totalCents && payment.Currency == Currency

	switch {
	case status == StatusPaid && matches:
		return OutcomeAlreadyPaid, nil // a repeated message

	case status == StatusPending && !overdue && matches:
		if err := markPaid(ctx, tx, payment.OrderID); err != nil {
			return "", err
		}
		if s.RestockSold {
			if err := restockLowOffers(ctx, tx, payment.OrderID); err != nil {
				return "", err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return "", fmt.Errorf("committing the payment: %w", err)
		}
		return OutcomeFulfilled, nil
	}

	// Money arrived that cannot be turned into keys: the order expired or
	// failed, or the payment does not match the order. Close the order if it
	// is still open, then send the money back.
	if status == StatusPending && matches {
		if err := endOrder(ctx, tx, payment.OrderID, StatusExpired); err != nil {
			return "", err
		}
	}
	if refundRef != nil {
		return OutcomeRefunded, tx.Commit(ctx) // refunded when the message first came
	}
	if s.Gateway == nil || payment.PaymentID == "" {
		return "", fmt.Errorf("order %s: a payment arrived that cannot be fulfilled and cannot be refunded automatically", payment.OrderID)
	}

	// The order row is still locked, so a repeated message waits and then
	// finds refund_ref set. The provider is also given the order id as a key
	// that makes a repeated refund request return the first refund.
	refundID, err := s.Gateway.Refund(ctx, payment.PaymentID, payment.OrderID)
	if err != nil {
		return "", fmt.Errorf("refunding the payment for order %s: %w", payment.OrderID, err)
	}
	if _, err := tx.Exec(ctx, "UPDATE orders SET refund_ref = $2 WHERE id = $1", payment.OrderID, refundID); err != nil {
		return "", fmt.Errorf("recording the refund: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return "", fmt.Errorf("committing the refund: %w", err)
	}
	log.Printf("order %s: payment refunded (%s), order status was %s", payment.OrderID, refundID, status)
	return OutcomeRefunded, nil
}

// ProviderEnded handles the provider's message that a payment page closed
// without a payment: it ran out of time (status expired) or a delayed
// payment failed (status failed). The order ends and its keys go back on
// sale at once, without waiting for the deadline.
func (s *Store) ProviderEnded(ctx context.Context, orderID, sessionID, status string) error {
	if !looksLikeUUID(orderID) || (status != StatusExpired && status != StatusFailed) {
		return nil
	}

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var current string
	var currentSession *string
	err = tx.QueryRow(ctx,
		"SELECT status, payment_ref FROM orders WHERE id = $1 FOR UPDATE", orderID,
	).Scan(&current, &currentSession)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("locking the order: %w", err)
	}

	// Only a pending order, and only if the message is about its own page.
	if current != StatusPending || currentSession == nil || *currentSession != sessionID {
		return nil
	}
	if err := endOrder(ctx, tx, orderID, status); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
