// Command server runs the API as a normal long-lived HTTP server.
// It is used for local development and would be the entry point on Render.
// On Vercel the entry point is api/index.go instead.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"eneba-marketplace/backend/config"
	"eneba-marketplace/backend/db"
	"eneba-marketplace/backend/httpapi"
)

func main() {
	config.LoadDotEnv(".env")

	pool, err := db.Connect(context.Background())
	if err != nil {
		// Keep running: /api/health will report the database as not configured.
		log.Printf("database: %v", err)
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// http.ListenAndServe has no time limits at all, so one slow or stuck
	// client could hold a connection open forever. An http.Server lets us
	// set them.
	server := &http.Server{
		Addr: ":" + port,
		Handler: httpapi.NewRouter(pool, httpapi.Config{
			// Set INSECURE_COOKIES=1 in .env for plain http://localhost.
			InsecureCookies: os.Getenv("INSECURE_COOKIES") == "1",
			DemoRestock:     os.Getenv("DEMO_RESTOCK") == "1",
		}),

		ReadHeaderTimeout: 5 * time.Second,  // time to send the request line and headers
		ReadTimeout:       10 * time.Second, // time to send the whole request
		WriteTimeout:      30 * time.Second, // time for us to send the response
		IdleTimeout:       60 * time.Second, // how long an unused keep-alive connection stays open
	}

	log.Printf("API listening on http://localhost:%s", port)
	log.Fatal(server.ListenAndServe())
}
