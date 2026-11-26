package httpapi

import (
	"context"
	"log"
	"net/http"
	"time"
)

type healthResponse struct {
	Status   string `json:"status"`   // "ok" or "degraded"
	Database string `json:"database"` // "ok", "not configured" or "unreachable"
}

// handleHealth reports whether the API is up and can reach the database.
// GET /api/health
func (api *API) handleHealth(w http.ResponseWriter, r *http.Request) {
	if api.db == nil {
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "degraded", Database: "not configured"})
		return
	}

	// Give the database 5 seconds. Neon's free tier suspends an idle database
	// and needs a moment to wake up on the first query.
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	var one int
	if err := api.db.QueryRow(ctx, "SELECT 1").Scan(&one); err != nil {
		log.Printf("health: database check failed: %v", err) // details stay in the server log
		writeJSON(w, http.StatusServiceUnavailable, healthResponse{Status: "degraded", Database: "unreachable"})
		return
	}

	writeJSON(w, http.StatusOK, healthResponse{Status: "ok", Database: "ok"})
}
