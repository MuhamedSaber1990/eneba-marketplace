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
	"os"
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
		// InsecureCookies is left at false: Secure cookies with the __Host-
		// prefix. DEMO_RESTOCK is set in the Vercel project for the public
		// demo, so the store cannot be bought empty.
		router = httpapi.NewRouter(pool, httpapi.Config{
			DemoRestock: os.Getenv("DEMO_RESTOCK") == "1",
		})
	})
	router.ServeHTTP(w, r)
}
