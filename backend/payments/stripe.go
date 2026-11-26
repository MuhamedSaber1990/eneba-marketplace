// Package payments connects the shop to Stripe. It is the only package that
// imports Stripe's library: the rest of the program sees the small Gateway
// interface in package checkout and a plain Event struct.
package payments

import (
	"context"
	"fmt"

	"github.com/stripe/stripe-go/v87"

	"eneba-marketplace/backend/checkout"
)

// Stripe implements checkout.Gateway with Stripe's hosted Checkout: Stripe
// shows the payment page, so no card data ever reaches this application.
type Stripe struct {
	client *stripe.Client
	appURL string // the site's public address, for sending the buyer back
}

// NewStripe builds the gateway. secretKey is a Stripe API key. A restricted
// key (rk_...) that may only write Checkout Sessions and Refunds is enough.
func NewStripe(secretKey, appURL string) *Stripe {
	return &Stripe{client: stripe.NewClient(secretKey), appURL: appURL}
}

// CreateSession asks Stripe for a payment page for one order.
func (s *Stripe) CreateSession(ctx context.Context, request checkout.SessionRequest) (checkout.Session, error) {
	// One line on Stripe's page per line of the order, with the price the
	// order recorded. Nothing is set up in Stripe's own product catalog.
	lineItems := make([]*stripe.CheckoutSessionCreateLineItemParams, 0, len(request.Items))
	for _, item := range request.Items {
		lineItems = append(lineItems, &stripe.CheckoutSessionCreateLineItemParams{
			Quantity: stripe.Int64(int64(item.Quantity)),
			PriceData: &stripe.CheckoutSessionCreateLineItemPriceDataParams{
				Currency:   stripe.String(checkout.Currency),
				UnitAmount: stripe.Int64(int64(item.UnitPriceCents)),
				ProductData: &stripe.CheckoutSessionCreateLineItemPriceDataProductDataParams{
					Name:        stripe.String(item.ProductName),
					Description: stripe.String("Sold by " + item.SellerName),
				},
			},
		})
	}

	orderURL := s.appURL + "/orders/" + request.OrderID
	params := &stripe.CheckoutSessionCreateParams{
		Mode:      stripe.String(stripe.CheckoutSessionModePayment),
		LineItems: lineItems,
		// The order id travels with the payment page and comes back in the
		// webhook. That is how a payment is matched to its order.
		ClientReferenceID: stripe.String(request.OrderID),
		ExpiresAt:         stripe.Int64(request.ExpiresAt.Unix()),
		// Stripe can offer to charge in the visitor's own currency, at its
		// exchange rate plus a fee ("adaptive pricing"). The shop's prices
		// are in euros and an order is only fulfilled when the payment is
		// for its exact euro total, so that offer is switched off. The buyer
		// sees and pays the same amount our site showed.
		AdaptivePricing: &stripe.CheckoutSessionCreateAdaptivePricingParams{Enabled: stripe.Bool(false)},
		// Where Stripe sends the buyer afterwards. Neither page fulfils the
		// order: only the webhook does. The query parameter just tells the
		// order page to wait for the webhook's result.
		SuccessURL: stripe.String(orderURL + "?checkout=done"),
		CancelURL:  stripe.String(orderURL + "?checkout=cancelled"),
	}
	// An idempotency key makes a repeated request return the first result.
	// If our request reaches Stripe but the answer is lost, trying again
	// gives the same payment page and not a second one.
	params.SetIdempotencyKey("checkout-" + request.OrderID)

	session, err := s.client.V1CheckoutSessions.Create(ctx, params)
	if err != nil {
		return checkout.Session{}, fmt.Errorf("stripe: creating a checkout session: %w", err)
	}
	return checkout.Session{ID: session.ID, URL: session.URL}, nil
}

// Refund sends a payment back in full.
func (s *Stripe) Refund(ctx context.Context, paymentID, orderID string) (string, error) {
	params := &stripe.RefundCreateParams{
		PaymentIntent: stripe.String(paymentID),
		Metadata:      map[string]string{"order_id": orderID, "reason": "order could not be fulfilled"},
	}
	params.SetIdempotencyKey("refund-" + orderID)

	refund, err := s.client.V1Refunds.Create(ctx, params)
	if err != nil {
		return "", fmt.Errorf("stripe: refunding %s: %w", paymentID, err)
	}
	return refund.ID, nil
}
