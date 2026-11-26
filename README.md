# Keybay: a marketplace for digital game keys

A take-home assignment for the Eneba internship: a marketplace web app in
Eneba's style. React and TypeScript frontend, Go API, PostgreSQL.

**Live:** https://eneba-marketplace-nine.vercel.app

This README is a working draft and is completed at the end of the project
(live URL, test credentials, architecture overview, known limitations).

- Scope and architecture: [PLAN.md](PLAN.md)
- Why each choice was made: [DECISIONS.md](DECISIONS.md)
- Design values measured from eneba.com: [docs/design-reference.md](docs/design-reference.md)
- Full AI chat history: [docs/chat-log.md](docs/chat-log.md)

## Run it locally

Needs Go 1.22 or newer and Node 20 or newer.

```
cp .env.example .env        # then put your two Neon connection strings in .env
go run ./cmd/migrate        # create the tables
go run ./cmd/seed           # fill them with the invented catalog
go run ./cmd/server         # API on http://localhost:8080
```

In a second terminal:

```
cd frontend
npm install
npm run dev                 # site on http://localhost:5173
```

The homepage shows a status panel. "Database: ok" means the frontend, the API
and the database are connected. The catalog is at `/store`.

## Tests

```
go test ./...
```

Tests that need a database use `DATABASE_URL` from `.env` and expect the seed
data. Without a database they are skipped.

## Layout

```
api/index.go      entry point on Vercel (serverless function)
cmd/server/       entry point for local runs and for Render
cmd/migrate/      applies the .sql files in migrations/
cmd/seed/         fills the database with invented products
migrations/       the database schema, one file per step
backend/          the API: config, database pool, catalog queries, HTTP routes
frontend/         React app (Vite)
docs/             chat log and design reference
```
