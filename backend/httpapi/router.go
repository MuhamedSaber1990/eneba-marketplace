// Package httpapi holds the HTTP routes of the API.
package httpapi

import (
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

// API carries what the handlers need. Handlers are methods on it, so they can
// reach the database without global variables.
type API struct {
	db *pgxpool.Pool
}

// NewRouter returns the handler for every /api route. The same router is used
// by the local server (cmd/server) and by the Vercel function (api/index.go).
//
// db may be nil when DATABASE_URL is missing. The health check reports that
// instead of the whole API failing to start.
func NewRouter(db *pgxpool.Pool) http.Handler {
	api := &API{db: db}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", api.handleHealth)

	// Anything else under /api is an unknown route. Answer in JSON, not with
	// Go's plain-text 404, so the frontend can parse every response.
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, http.StatusNotFound, "route not found")
	})
	return mux
}
