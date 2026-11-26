package httpapi

import (
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"eneba-marketplace/backend/checkout"
	"eneba-marketplace/backend/payments"
)

// The two ways an order can be paid. Exactly one is active.
const (
	PaymentSimulated = "simulated" // two buttons on the order page; for local runs, tests, and as a fallback
	PaymentStripe    = "stripe"    // Stripe's hosted payment page, in test mode
)

// ConfigFromEnv reads the API's settings from environment variables. Both
// entry points (cmd/server and the Vercel function) call it, so they cannot
// drift apart.
//
//	INSECURE_COOKIES=1      plain http://localhost only
//	DEMO_RESTOCK=1          the public demo tops up stock after sales
//	PAYMENT_MODE=stripe     use Stripe; anything else means simulated
//	STRIPE_SECRET_KEY       a Stripe API key (test mode)
//	STRIPE_WEBHOOK_SECRET   the signing secret of the webhook endpoint
//	APP_URL                 the site's public address, e.g. https://example.com
func ConfigFromEnv() Config {
	config := Config{
		InsecureCookies: os.Getenv("INSECURE_COOKIES") == "1",
		DemoRestock:     os.Getenv("DEMO_RESTOCK") == "1",
		PaymentMode:     PaymentSimulated,
	}

	if os.Getenv("PAYMENT_MODE") == PaymentStripe {
		secretKey := os.Getenv("STRIPE_SECRET_KEY")
		webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")
		appURL := strings.TrimRight(os.Getenv("APP_URL"), "/")

		config.PaymentMode = PaymentStripe
		if secretKey == "" || webhookSecret == "" || appURL == "" {
			// Stay in Stripe mode with no gateway: paying is refused with a
			// clear error. Falling back to simulated payment here would
			// quietly let anyone "pay" for free on a site meant to use Stripe.
			log.Print("PAYMENT_MODE is stripe but STRIPE_SECRET_KEY, STRIPE_WEBHOOK_SECRET or APP_URL is missing: payments are disabled")
			return config
		}
		config.Gateway = payments.NewStripe(secretKey, appURL)
		config.StripeWebhookSecret = webhookSecret
	}
	return config
}

// simulatedOnly and stripeOnly make a route exist in one payment mode only.
// In the other mode it answers 404, as if it were not there. Hiding a button
// in the frontend would not be enough: on a site that uses Stripe, the
// simulated endpoint must not be callable at all.
func (api *API) simulatedOnly(next http.HandlerFunc) http.HandlerFunc {
	return api.onlyInMode(PaymentSimulated, next)
}

func (api *API) stripeOnly(next http.HandlerFunc) http.HandlerFunc {
	return api.onlyInMode(PaymentStripe, next)
}

func (api *API) onlyInMode(mode string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if api.config.PaymentMode != mode {
			writeError(w, http.StatusNotFound, "route not found")
			return
		}
		next(w, r)
	}
}

// handleStartCheckout sends the buyer to the payment provider's page.
// POST /api/orders/{id}/checkout  ->  {"url": "https://checkout.stripe.com/..."}
func (api *API) handleStartCheckout(w http.ResponseWriter, r *http.Request) {
	url, err := api.checkout.StartPayment(r.Context(), currentUser(r).ID, r.PathValue("id"))
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"url": url})
}

// maxWebhookBytes caps the size of a webhook body. Stripe's events for a
// checkout session are a few kilobytes.
const maxWebhookBytes = 512 * 1024

// handleStripeWebhook receives Stripe's messages about payments. Stripe's
// servers call it, not a browser, so it has no cookie and no Origin check.
// What protects it is the signature on every request (payments.ParseWebhook).
//
// Its answer tells Stripe what to do next: 2xx means "received", anything
// else makes Stripe send the event again later. So a failure on our side
// answers 500 on purpose, and an event we do not act on answers 200.
// POST /api/stripe/webhook
func (api *API) handleStripeWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxWebhookBytes))
	if err != nil {
		writeError(w, http.StatusBadRequest, "the request body could not be read")
		return
	}

	event, err := payments.ParseWebhook(body, r.Header.Get("Stripe-Signature"), api.config.StripeWebhookSecret)
	if err != nil {
		// Not from Stripe, or damaged on the way. Nothing is done with it.
		log.Printf("stripe webhook refused: %v", err)
		writeError(w, http.StatusBadRequest, "invalid signature")
		return
	}

	outcome := "nothing to do"
	switch event.Kind {
	case payments.KindPaid:
		outcome, err = api.checkout.ProviderPaid(r.Context(), checkout.ProviderPayment{
			OrderID:     event.OrderID,
			SessionID:   event.SessionID,
			PaymentID:   event.PaymentID,
			AmountCents: event.AmountCents,
			Currency:    event.Currency,
		})
	case payments.KindFailed:
		outcome = "order failed"
		err = api.checkout.ProviderEnded(r.Context(), event.OrderID, event.SessionID, checkout.StatusFailed)
	case payments.KindExpired:
		outcome = "order expired"
		err = api.checkout.ProviderEnded(r.Context(), event.OrderID, event.SessionID, checkout.StatusExpired)
	}
	if err != nil {
		serverError(w, r, err) // 500: Stripe will send the event again
		return
	}

	log.Printf("stripe webhook %s (%s) for order %s: %s", event.ID, event.Kind, event.OrderID, outcome)
	writeJSON(w, http.StatusOK, map[string]string{"result": outcome})
}
