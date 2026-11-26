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

The homepage is built from blocks the API sends. The catalog is at `/store`,
and `/api/health` reports whether the API can reach the database.

## About the data

Everything in the shop is invented for this project and created by
`cmd/seed`: the games, the sellers, the prices and the keys.

Two things on the homepage look like facts and are not:

- **Best sellers.** The "units sold" behind the ranking are random numbers
  from the seed. Nothing was sold.
- **Deals.** The seed picks eight offers, records their price as the "was"
  price and lowers the real price by 15 to 40 percent. No price was ever
  higher. Their end dates are real and spread over about 60 days from the day
  the seed ran, so the block empties out over time. When a deal ends, its
  price does not go back up by itself.

## Search

The search box in the header works from every page. It tolerates typos and
partial words: `circut` finds "Circuit Breakers", `forge` finds "Mythforge
Legends", `kart xbox` finds the Xbox edition of "Thunder Circuit Kart".

What it does not do:

- Fix a typo in a word of three letters or fewer (`vwo` for "vow").
- Fix swapped letters in a short word (`krat` for "kart", `rouge` for
  "rogue").
- Understand meaning. It searches product names, so `racing` does not find a
  racing game.

## Tests

```
go test ./...
```

Tests that need a database use `DATABASE_URL` from `.env` and expect the seed
data. Without a database they are skipped.

```
cd frontend && npm test
```

Frontend unit tests (Vitest) for small pieces of logic such as price
formatting and the deals countdown.

## Checks

Every push and pull request runs `.github/workflows/ci.yml` on GitHub:
formatting, `go vet`, `govulncheck`, the migrations and seed on an empty
database, the Go tests, then the frontend lint, tests and build.

To run the same checks locally:

```
gofmt -l . && go vet ./... && go test ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
cd frontend && npm run lint && npm test && npm run build
```

To see the site with production's security headers (they are defined in
`vercel.json`), build it and use preview mode instead of the dev server:

```
cd frontend && npm run build && npm run preview    # http://localhost:4173
```

If port 8080 is taken, start the API on another port and tell the frontend:
`PORT=8090 go run ./cmd/server`, then `API_URL=http://localhost:8090 npm run dev`.

## Layout

```
api/index.go      entry point on Vercel (serverless function)
cmd/server/       entry point for local runs and for Render
cmd/migrate/      applies the .sql files in migrations/
cmd/seed/         fills the database with invented products
migrations/       the database schema, one file per step
backend/          the API: config, database pool, catalog and homepage queries, HTTP routes
frontend/         React app (Vite)
docs/             chat log and design reference
```
