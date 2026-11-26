package payments

import (
	"encoding/json"
	"fmt"

	"github.com/stripe/stripe-go/v87"
)

// What a webhook event means for an order.
const (
	KindPaid    = "paid"    // the payment succeeded
	KindFailed  = "failed"  // a delayed payment failed
	KindExpired = "expired" // the payment page closed without a payment
	KindOther   = "other"   // nothing this shop acts on
)

// Event is a Stripe webhook event reduced to what the shop needs.
type Event struct {
	ID          string
	Kind        string
	OrderID     string // the order id given to Stripe when the page was created
	SessionID   string
	PaymentID   string
	AmountCents int
	Currency    string
}

// sessionPayload is the part of a Checkout Session the shop reads from an
// event.
type sessionPayload struct {
	ID                string `json:"id"`
	ClientReferenceID string `json:"client_reference_id"`
	PaymentStatus     string `json:"payment_status"`
	PaymentIntent     string `json:"payment_intent"`
	AmountTotal       int    `json:"amount_total"`
	Currency          string `json:"currency"`
}

// ParseWebhook checks that a webhook request really comes from Stripe and
// reads the event from it.
//
// The check is the whole security of the webhook endpoint. Stripe signs each
// request with a secret only Stripe and this server know, over the exact
// bytes of the body plus the time it was sent. ConstructEvent recomputes
// that signature and refuses the request if it differs, or if it is more
// than five minutes old (so a captured request cannot be replayed later).
// Without the secret, nobody can produce a request this function accepts.
//
// body must be the raw bytes as received. Parsing and re-encoding the JSON
// first would change the bytes and break the signature.
func ParseWebhook(body []byte, signatureHeader, signingSecret string) (Event, error) {
	// Stripe's library refuses events written for a different API version
	// than its own by default. This code reads six plain fields that are the
	// same across versions, so that check is switched off: otherwise a
	// library update or a dashboard setting could silently stop every
	// payment from being fulfilled.
	raw, err := stripe.ConstructEvent(body, signatureHeader, signingSecret, stripe.WithIgnoreAPIVersionMismatch())
	if err != nil {
		return Event{}, fmt.Errorf("verifying the webhook signature: %w", err)
	}

	event := Event{ID: raw.ID, Kind: KindOther}
	switch raw.Type {
	case "checkout.session.completed",
		"checkout.session.async_payment_succeeded",
		"checkout.session.async_payment_failed",
		"checkout.session.expired":
	default:
		return event, nil
	}

	var session sessionPayload
	if raw.Data == nil {
		return Event{}, fmt.Errorf("event %s has no data", raw.ID)
	}
	if err := json.Unmarshal(raw.Data.Raw, &session); err != nil {
		return Event{}, fmt.Errorf("reading the session of event %s: %w", raw.ID, err)
	}
	event.OrderID = session.ClientReferenceID
	event.SessionID = session.ID
	event.PaymentID = session.PaymentIntent
	event.AmountCents = session.AmountTotal
	event.Currency = session.Currency

	switch raw.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		// "completed" only means the buyer finished the page. With payment
		// methods that confirm later, the money has not arrived yet and the
		// status is still "unpaid". Fulfil only on "paid".
		if session.PaymentStatus == "paid" {
			event.Kind = KindPaid
		}
	case "checkout.session.async_payment_failed":
		event.Kind = KindFailed
	case "checkout.session.expired":
		event.Kind = KindExpired
	}
	return event, nil
}
