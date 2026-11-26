// Package handler is the entry point on Vercel.
//
// Vercel turns each .go file in /api into a serverless function and calls its
// exported Handler for every request. vercel.json sends all /api/* requests
// here, and this file passes them to the same router the local server uses.
package handler

import (
	"context"
	"log"
	"net/http"
	"sync"

	"eneba-marketplace/backend/db"
	"eneba-marketplace/backend/httpapi"
)

var (
	router http.Handler
	once   sync.Once
)

// Handler is called by Vercel. A function instance stays alive for a while
// and serves many requests, so the router and its database pool are built
// once (sync.Once) and reused.
func Handler(w http.ResponseWriter, r *http.Request) {
	once.Do(func() {
		pool, err := db.Connect(context.Background())
		if err != nil {
			log.Printf("database: %v", err)
		}
		// The zero Config is the secure one: Secure cookies with the __Host- prefix.
		router = httpapi.NewRouter(pool, httpapi.Config{})
	})
	router.ServeHTTP(w, r)
}
