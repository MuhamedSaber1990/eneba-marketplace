// Command server runs the API as a normal long-lived HTTP server.
// It is used for local development and would be the entry point on Render.
// On Vercel the entry point is api/index.go instead.
package main

import (
	"context"
	"log"
	"net/http"
	"os"

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

	log.Printf("API listening on http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, httpapi.NewRouter(pool)))
}
