package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

// writeJSON sends any Go value as a JSON response with the given status code.
func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		log.Printf("writeJSON: %v", err)
	}
}

// writeError sends {"error": "..."} so the frontend can always read the same
// shape when a request fails.
func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

// serverError is for failures that are not the caller's fault. The details go
// to the server log, and the caller gets a message that reveals nothing about
// the database.
func serverError(w http.ResponseWriter, r *http.Request, err error) {
	// The visitor went away before the answer was ready (closed the tab,
	// moved to another page). The request was cancelled on purpose, nobody
	// is left to answer, and it is not a failure worth a line in the log.
	if errors.Is(err, context.Canceled) {
		return
	}
	log.Printf("%s %s: %v", r.Method, r.URL.Path, err)
	writeError(w, http.StatusInternalServerError, "something went wrong on our side")
}
