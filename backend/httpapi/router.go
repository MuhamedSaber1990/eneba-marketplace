// Package httpapi holds the HTTP routes of the API.
package httpapi

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"eneba-marketplace/backend/auth"
	"eneba-marketplace/backend/catalog"
	"eneba-marketplace/backend/checkout"
	"eneba-marketplace/backend/home"
)

// Config holds the settings that differ between where the API runs.
type Config struct {
	// InsecureCookies drops the Secure flag and the __Host- prefix from the
	// session cookie. It exists for plain http://localhost and must stay
	// false anywhere real.
	InsecureCookies bool

	// DemoRestock makes the store add fresh keys to an offer that runs low
	// after a sale, so the public demo cannot be bought empty. See
	// backend/checkout/restock.go. A real shop would leave it off.
	DemoRestock bool

	// PaymentMode is PaymentSimulated or PaymentStripe. It decides which
	// payment routes exist and which payment step the order page shows.
	PaymentMode string

	// Gateway is the payment provider, used in Stripe mode. It is nil in
	// simulated mode, and also in Stripe mode when the keys are missing.
	Gateway checkout.Gateway

	// StripeWebhookSecret is what the webhook's signatures are checked with.
	StripeWebhookSecret string
}

// API carries what the handlers need. Handlers are methods on it, so they can
// reach the database without global variables.
type API struct {
	db       *pgxpool.Pool
	config   Config
	catalog  *catalog.Store
	home     *home.Store
	auth     *auth.Store
	checkout *checkout.Store
}

// NewRouter returns the handler for every /api route. The same router is used
// by the local server (cmd/server) and by the Vercel function (api/index.go).
//
// db may be nil when DATABASE_URL is missing. The health check reports that,
// and every other route answers 503 instead of crashing.
func NewRouter(db *pgxpool.Pool, config Config) http.Handler {
	if config.PaymentMode == "" {
		config.PaymentMode = PaymentSimulated
	}
	checkoutStore := checkout.NewStore(db)
	checkoutStore.RestockSold = config.DemoRestock
	checkoutStore.Gateway = config.Gateway

	api := &API{
		db:       db,
		config:   config,
		catalog:  catalog.NewStore(db),
		home:     home.NewStore(db),
		auth:     auth.NewStore(db),
		checkout: checkoutStore,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", api.handleHealth)

	mux.HandleFunc("GET /api/products", api.needsDB(api.handleListProducts))
	mux.HandleFunc("GET /api/products/{slug}", api.needsDB(api.handleGetProduct))
	mux.HandleFunc("GET /api/products/{slug}/cover.svg", api.needsDB(api.handleProductCover))
	mux.HandleFunc("GET /api/filters", api.needsDB(api.handleFilters))
	mux.HandleFunc("GET /api/search/suggest", api.needsDB(api.handleSuggest))

	// The homepage: one endpoint for the layout, one per block.
	mux.HandleFunc("GET /api/home", api.needsDB(api.handleHomeLayout))
	mux.HandleFunc("GET /api/home/banners", api.needsDB(api.handleHomeBanners))
	mux.HandleFunc("GET /api/home/deals", api.needsDB(api.handleHomeDeals))
	mux.HandleFunc("GET /api/home/best-sellers", api.needsDB(api.handleHomeBestSellers))
	mux.HandleFunc("GET /api/home/new-releases", api.needsDB(api.handleHomeNewReleases))
	mux.HandleFunc("GET /api/home/platforms", api.needsDB(api.handleHomePlatforms))

	// Accounts. The three POST routes change something, so each one is
	// wrapped in sameOrigin. Middleware reads from the outside in: check
	// the database is there, then check the origin, then run the handler.
	mux.HandleFunc("POST /api/auth/signup", api.needsDB(sameOrigin(api.handleSignUp)))
	mux.HandleFunc("POST /api/auth/signin", api.needsDB(sameOrigin(api.handleSignIn)))
	mux.HandleFunc("POST /api/auth/signout", api.needsDB(sameOrigin(api.handleSignOut)))
	mux.HandleFunc("GET /api/auth/me", api.needsDB(api.handleMe))
	mux.HandleFunc("POST /api/account", api.needsDB(sameOrigin(api.requireUser(api.handleUpdateAccount))))

	// Checkout. Pricing a cart is public. Everything about orders needs a
	// signed-in user, and the POSTs must come from the site itself.
	mux.HandleFunc("POST /api/cart/price", api.needsDB(sameOrigin(api.handlePriceCart)))
	mux.HandleFunc("POST /api/orders", api.needsDB(sameOrigin(api.requireUser(api.handleCreateOrder))))
	mux.HandleFunc("GET /api/orders", api.needsDB(api.requireUser(api.handleListOrders)))
	mux.HandleFunc("GET /api/orders/{id}", api.needsDB(api.requireUser(api.handleGetOrder)))

	// Paying. Which of these exist depends on the payment mode.
	//   simulated: the buyer picks the outcome with a button.
	//   stripe:    the buyer is sent to Stripe, and Stripe reports the result
	//              to the webhook. The webhook is called by Stripe's servers,
	//              so it has no session and no Origin check; its signature
	//              is what proves where it came from.
	mux.HandleFunc("POST /api/orders/{id}/pay", api.simulatedOnly(api.needsDB(sameOrigin(api.requireUser(api.handlePayOrder)))))
	mux.HandleFunc("POST /api/orders/{id}/checkout", api.stripeOnly(api.needsDB(sameOrigin(api.requireUser(api.handleStartCheckout)))))
	mux.HandleFunc("POST /api/stripe/webhook", api.stripeOnly(api.needsDB(api.handleStripeWebhook)))
	// Anything else under /api is an unknown route. Answer in JSON, not with
	// Go's plain-text 404, so the frontend can parse every response.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "route not found")
	})
	return mux
}

// needsDB wraps a handler that uses the database. It is a small piece of
// middleware: a function that takes a handler and returns a new one that
// checks something first.
func (api *API) needsDB(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if api.db == nil {
			writeError(w, http.StatusServiceUnavailable, "the database is not configured")
			return
		}
		next(w, r)
	}
}
