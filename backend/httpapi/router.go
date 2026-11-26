// Package httpapi holds the HTTP routes of the API.
package httpapi

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"eneba-marketplace/backend/catalog"
	"eneba-marketplace/backend/home"
)

// API carries what the handlers need. Handlers are methods on it, so they can
// reach the database without global variables.
type API struct {
	db      *pgxpool.Pool
	catalog *catalog.Store
	home    *home.Store
}

// NewRouter returns the handler for every /api route. The same router is used
// by the local server (cmd/server) and by the Vercel function (api/index.go).
//
// db may be nil when DATABASE_URL is missing. The health check reports that,
// and every other route answers 503 instead of crashing.
func NewRouter(db *pgxpool.Pool) http.Handler {
	api := &API{db: db, catalog: catalog.NewStore(db), home: home.NewStore(db)}

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
