package httpapi

import (
	"errors"
	"net/http"

	"eneba-marketplace/backend/checkout"
)

// cartRequest is a cart as the browser sends it. ExpectedTotalCents is only
// used when placing an order: it is the total the buyer was looking at.
type cartRequest struct {
	Items              []checkout.Line `json:"items"`
	ExpectedTotalCents int             `json:"expected_total_cents"`
}

// handlePriceCart prices a cart and reports stock. It is public, because the
// cart exists before anyone signs in, and it changes nothing. It is a POST
// only because a cart does not fit comfortably in a URL.
// POST /api/cart/price  {"items": [{"offer_id": 4, "quantity": 1}]}
func (api *API) handlePriceCart(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	var request cartRequest
	if !readJSON(w, r, &request) {
		return
	}

	cart, err := api.checkout.PriceCart(r.Context(), request.Items)
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, cart)
}

// handleCreateOrder turns the cart into a pending order with reserved keys.
// POST /api/orders  {"items": [...], "expected_total_cents": 4899}
func (api *API) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	var request cartRequest
	if !readJSON(w, r, &request) {
		return
	}

	order, err := api.checkout.CreateOrder(r.Context(), currentUser(r).ID, request.Items, request.ExpectedTotalCents)
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"order": api.withMode(order)})
}

// handlePayOrder is the simulated payment: the buyer chooses whether the
// "card" is approved or declined. The Stripe phase replaces this endpoint
// with Stripe's webhook, which calls the same checkout.Pay.
// POST /api/orders/{id}/pay  {"outcome": "approve"}  or  {"outcome": "decline"}
func (api *API) handlePayOrder(w http.ResponseWriter, r *http.Request) {
	var request struct {
		Outcome string `json:"outcome"`
	}
	if !readJSON(w, r, &request) {
		return
	}
	if request.Outcome != "approve" && request.Outcome != "decline" {
		writeError(w, http.StatusBadRequest, `outcome must be "approve" or "decline"`)
		return
	}

	order, err := api.checkout.Pay(r.Context(), currentUser(r).ID, r.PathValue("id"), request.Outcome == "approve")
	if errors.Is(err, checkout.ErrOrderNotPayable) {
		// The order ended earlier (declined or run out of time). Send its
		// final state along, so the page can show it.
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "This order can no longer be paid.",
			"code":  "not_payable",
			"order": api.withMode(order),
		})
		return
	}
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": api.withMode(order)})
}

// handleGetOrder returns one of the signed-in user's orders. Keys are part
// of it only once it is paid.
// GET /api/orders/{id}
func (api *API) handleGetOrder(w http.ResponseWriter, r *http.Request) {
	order, err := api.checkout.GetOrder(r.Context(), currentUser(r).ID, r.PathValue("id"))
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"order": api.withMode(order)})
}

// handleListOrders returns the signed-in user's order history.
// GET /api/orders
func (api *API) handleListOrders(w http.ResponseWriter, r *http.Request) {
	orders, err := api.checkout.ListOrders(r.Context(), currentUser(r).ID)
	if writeCheckoutError(w, r, err) {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": orders})
}

// withMode adds the active payment mode to an order, so the order page knows
// which payment step to show.
func (api *API) withMode(order checkout.Order) checkout.Order {
	order.PaymentMode = api.config.PaymentMode
	return order
}

// writeCheckoutError turns an error from package checkout into a response.
// It reports true when it has written one.
//
// Failures the frontend reacts to carry a "code" next to the message, so the
// page can decide what to do without reading English text.
func writeCheckoutError(w http.ResponseWriter, r *http.Request, err error) bool {
	var validation checkout.ValidationError
	var stock *checkout.StockError
	switch {
	case err == nil:
		return false
	case errors.As(err, &validation):
		writeError(w, http.StatusBadRequest, validation.Message)
	case errors.Is(err, checkout.ErrOrderNotFound):
		writeError(w, http.StatusNotFound, "order not found")
	case errors.Is(err, checkout.ErrOfferNotFound):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "An item in your cart is no longer for sale.",
			"code":  "offer_gone",
		})
	case errors.Is(err, checkout.ErrPriceChanged):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "A price changed while you were checking out. Review the new total.",
			"code":  "price_changed",
		})
	case errors.As(err, &stock):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error":     "Not enough keys left for " + stock.ProductName + ".",
			"code":      "out_of_stock",
			"offer_id":  stock.OfferID,
			"requested": stock.Requested,
		})
	case errors.Is(err, checkout.ErrOrderNotPayable):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "This order can no longer be paid.",
			"code":  "not_payable",
		})
	case errors.Is(err, checkout.ErrNoGateway):
		writeError(w, http.StatusServiceUnavailable, "Card payments are not available right now.")
	case errors.Is(err, checkout.ErrTooManyPending):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "You have unpaid orders waiting. Pay one, or wait for it to run out, before starting another.",
			"code":  "too_many_pending",
		})
	default:
		serverError(w, r, err)
	}
	return true
}
