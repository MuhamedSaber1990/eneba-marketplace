package payments

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/stripe/stripe-go/v87"
)

// These tests need no Stripe account. Stripe's library can sign a payload
// with a secret of our choosing, exactly the way Stripe's servers would.

const testSecret = "whsec_test_secret_used_only_in_tests"

// signedEvent builds a webhook request as Stripe would send it: the JSON
// body and its Stripe-Signature header.
func signedEvent(t *testing.T, secret, eventType string, session map[string]any) (body []byte, header string) {
	t.Helper()
	payload, err := json.Marshal(map[string]any{
		"id":          "evt_test_1",
		"object":      "event",
		"api_version": "2020-01-01", // deliberately not the library's version
		"type":        eventType,
		"data":        map[string]any{"object": session},
	})
	if err != nil {
		t.Fatalf("building the payload: %v", err)
	}
	signed := stripe.GenerateTestSignedPayload(&stripe.UnsignedPayload{Payload: payload, Secret: secret})
	return signed.Payload, signed.Header
}

func paidSession() map[string]any {
	return map[string]any{
		"id":                  "cs_test_123",
		"object":              "checkout.session",
		"client_reference_id": "0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f",
		"payment_status":      "paid",
		"payment_intent":      "pi_test_456",
		"amount_total":        2607,
		"currency":            "eur",
	}
}

func TestParseWebhookReadsAPaidSession(t *testing.T) {
	body, header := signedEvent(t, testSecret, "checkout.session.completed", paidSession())

	event, err := ParseWebhook(body, header, testSecret)
	if err != nil {
		t.Fatalf("ParseWebhook: %v", err)
	}
	want := Event{
		ID:          "evt_test_1",
		Kind:        KindPaid,
		OrderID:     "0b0f3a52-8f0f-4c0a-9d0e-2b7c1c3d4e5f",
		SessionID:   "cs_test_123",
		PaymentID:   "pi_test_456",
		AmountCents: 2607,
		Currency:    "eur",
	}
	if event != want {
		t.Errorf("event = %+v\nwant    %+v", event, want)
	}
}

func TestParseWebhookKinds(t *testing.T) {
	tests := []struct {
		name          string
		eventType     string
		paymentStatus string
		want          string
	}{
		{"completed and paid", "checkout.session.completed", "paid", KindPaid},
		// The buyer finished the page, but the money has not arrived yet.
		// Fulfilling here would hand out keys for a payment that may fail.
		{"completed but still unpaid", "checkout.session.completed", "unpaid", KindOther},
		{"a delayed payment arrived", "checkout.session.async_payment_succeeded", "paid", KindPaid},
		{"a delayed payment failed", "checkout.session.async_payment_failed", "unpaid", KindFailed},
		{"the page ran out of time", "checkout.session.expired", "unpaid", KindExpired},
		{"an event the shop does not use", "customer.created", "", KindOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := paidSession()
			session["payment_status"] = tt.paymentStatus
			body, header := signedEvent(t, testSecret, tt.eventType, session)

			event, err := ParseWebhook(body, header, testSecret)
			if err != nil {
				t.Fatalf("ParseWebhook: %v", err)
			}
			if event.Kind != tt.want {
				t.Errorf("kind = %q, want %q", event.Kind, tt.want)
			}
		})
	}
}

// Everything below must be refused. If any of these were accepted, anyone
// who knows the webhook's address could mark orders as paid.
func TestParseWebhookRefusesWhatStripeDidNotSign(t *testing.T) {
	body, header := signedEvent(t, testSecret, "checkout.session.completed", paidSession())

	t.Run("no signature header", func(t *testing.T) {
		if _, err := ParseWebhook(body, "", testSecret); err == nil {
			t.Error("accepted a request with no signature")
		}
	})

	t.Run("signed with a different secret", func(t *testing.T) {
		_, forgedHeader := signedEvent(t, "whsec_a_guess", "checkout.session.completed", paidSession())
		if _, err := ParseWebhook(body, forgedHeader, testSecret); err == nil {
			t.Error("accepted a signature made with the wrong secret")
		}
	})

	t.Run("the body was changed after signing", func(t *testing.T) {
		// A real signature, but for a different amount than the body now says.
		tampered := []byte(strings.Replace(string(body), `"amount_total":2607`, `"amount_total":1`, 1))
		if string(tampered) == string(body) {
			t.Fatal("the test did not change the body")
		}
		if _, err := ParseWebhook(tampered, header, testSecret); err == nil {
			t.Error("accepted a body that does not match its signature")
		}
	})

	t.Run("a genuine request replayed ten minutes later", func(t *testing.T) {
		payload := []byte(`{"id":"evt_old","object":"event","type":"checkout.session.completed","data":{"object":{}}}`)
		old := stripe.GenerateTestSignedPayload(&stripe.UnsignedPayload{
			Payload:   payload,
			Secret:    testSecret,
			Timestamp: time.Now().Add(-10 * time.Minute),
		})
		if _, err := ParseWebhook(old.Payload, old.Header, testSecret); err == nil {
			t.Error("accepted a request signed ten minutes ago")
		}
	})

	t.Run("an empty secret on our side accepts nothing", func(t *testing.T) {
		if _, err := ParseWebhook(body, header, ""); err == nil {
			t.Error("accepted a request while no secret is configured")
		}
	})
}
