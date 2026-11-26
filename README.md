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

## Accounts

Sign up with any email and a password of at least 8 characters. Nothing is
sent to the address, so it does not have to be a real one.

A demo account also exists:

| Email | Password |
|---|---|
| `demo@keybay.test` | `keybay-demo-2026` |

These credentials are public on purpose. If the demo account misbehaves for
any reason, create your own account instead: it takes ten seconds.

How it is protected (the reasons are in DECISIONS.md, entries 49 to 56):

- Passwords are stored as bcrypt hashes.
- The session is a random token in an `HttpOnly`, `Secure`, `SameSite=Lax`
  cookie. The database holds only a hash of it. Sessions last 7 days and
  signing out ends them on the server.
- Requests that change something must come from the site itself and be JSON.
- Failed sign-ins are limited per email and per IP address.
- A failed sign-in does not say whether the email exists.

What is not built: email verification, password reset, changing a password,
deleting an account, two-factor login. Signing up with an email that is
already registered says so, which does reveal that the account exists.

## Buying something

Add an offer to the cart from a product page, go to checkout (you will be
asked to sign in), place the order and pay. Payment is simulated: the order
page has a "Pay (test)" button and a "Simulate a declined card" button. No
card is asked for and no money moves. After paying, the keys are on the
order page and stay in your order history.

- Placing an order holds its keys for 15 minutes. If it is not paid by then,
  it expires and the keys go back on sale.
- A key cannot be sold twice. The tests that check this start 20 buyers at
  the same instant against a real database (`backend/checkout/`).
- If someone else is holding the last key in an unpaid order, you are told
  it is out of stock, even though it may come back if they do not pay.

- On the live demo, stock is topped up automatically so the store cannot be
  bought empty. Paying is free there and the demo password is public, so
  without this anyone could buy every key. When a sale leaves an offer with
  fewer than 3 keys, 5 new ones are generated. The setting is `DEMO_RESTOCK`,
  and it is off by default and in the tests.

Not built: emailing the keys, refunds, cancelling an order, invoices, tax,
checkout without an account. Keys are stored unencrypted.

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
go test -p 1 ./...
```

Tests that need a database use `DATABASE_URL` from `.env` and expect the seed
data. Without a database they are skipped. `-p 1` runs one package at a time:
the packages share that database, and the checkout tests add a temporary
product that the catalog tests would otherwise count.

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
gofmt -l . && go vet ./... && go test -p 1 ./...
go run golang.org/x/vuln/cmd/govulncheck@latest ./...
cd frontend && npm run lint && npm test && npm run build
```

To see the site with production's security headers (they are defined in
`vercel.json`), build it and use preview mode instead of the dev server:

```
cd frontend && npm run build && npm run preview    # http://localhost:4173
```

Signing in on plain `http://localhost` needs `INSECURE_COOKIES=1` in `.env`
(it is in `.env.example`). Without it the session cookie requires HTTPS.

If port 8080 is taken, start the API on another port and tell the frontend:
`PORT=8090 go run ./cmd/server`, then `API_URL=http://localhost:8090 npm run dev`.

## Layout

```
api/index.go      entry point on Vercel (serverless function)
cmd/server/       entry point for local runs and for Render
cmd/migrate/      applies the .sql files in migrations/
cmd/seed/         fills the database with invented products
migrations/       the database schema, one file per step
backend/          the API: config, database pool, catalog, homepage, accounts, checkout, HTTP routes
frontend/         React app (Vite)
docs/             chat log and design reference
```
