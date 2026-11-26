# Decisions

Each entry: what was chosen, what else was considered, and why. Newest at the
bottom. Scope decisions (what is in depth, minimal and out) are in `PLAN.md`.

## 1. Backend in Go, standard library only for HTTP

- **Alternatives:** PHP with Laravel or Symfony; Go with a router library (chi,
  gin).
- **Why:** the brief allows PHP or Go. Go was new to me, so I wanted as little
  hidden framework behavior as possible. Since Go 1.22 the built-in `ServeMux`
  matches patterns like `GET /api/products/{slug}`, so a router library would
  add a dependency without adding anything I need.

## 2. REST with JSON

- **Alternatives:** GraphQL.
- **Why:** the homepage requirement ("each block gets its content from its own
  backend service") maps onto one endpoint per block. REST is also easy to test
  with curl and needs no schema tooling.

## 3. PostgreSQL, plain SQL, no ORM

- **Alternatives:** SQLite; MySQL; an ORM such as GORM.
- **Why:** PostgreSQL's `pg_trgm` extension gives typo-tolerant search inside
  the database, so there is no separate search service to host. Row locking
  (`FOR UPDATE SKIP LOCKED`) solves selling the same key twice. I am
  comfortable with SQL, and with plain queries every statement is visible in
  the code.

## 4. Search with trigrams in the database

- **Alternatives:** Meilisearch, Typesense or Algolia; fuzzy matching in the
  browser (Fuse.js); `ILIKE` only.
- **Why:** an external engine is one more free-tier service to keep alive and
  to keep in sync. Browser-side matching means shipping the whole catalog to
  every visitor. `ILIKE` alone handles partial names but not typos.

## 5. Sessions in an httpOnly cookie, stored in the database

- **Alternatives:** JWT in `localStorage`; JWT in a cookie.
- **Why:** an httpOnly cookie cannot be read by JavaScript, which removes the
  most common way tokens get stolen. A database row per session makes sign-out
  real (the row is deleted) and suits serverless functions, which keep no
  memory. The frontend and API share one domain, so there is no CORS setup.

## 6. Cart in `localStorage`, re-priced by the server

- **Alternatives:** a cart table in the database.
- **Why:** the brief's flow puts the cart before sign-in, so a visitor must be
  able to fill a cart without an account. The server recalculates every price
  and checks stock when the order is created, so nothing from the browser is
  trusted.

## 7. Homepage blocks as separate modules, not separate deployments

- **Alternatives:** one microservice per block; one endpoint returning the
  whole homepage.
- **Why:** each block has its own endpoint, its own query and its own data
  source, and the layout (which blocks, order, titles) comes from a table. The
  frontend renders what it receives and each block loads and fails on its own.
  Separate deployed services would multiply free-tier cold starts and add no
  behavior. One combined endpoint would mean a slow block delays all of them.

## 8. Vercel for frontend and API, Neon for the database

- **Alternatives:** Render for the API; everything on Render.
- **Why:** one domain for frontend and API, and Neon's free database does not
  expire, where Render's free database is removed after about a month. Vercel's
  Go runtime is in beta, so the API is written as a normal `net/http` server
  that also has a `cmd/server` entry point listening on `PORT`. It can move to
  Render with no code change. Neon's pooled connection string is used because
  serverless functions open many short connections.

## 9. Stripe hosted Checkout in test mode, plus a simulated mode

- **Alternatives:** simulated payment only; Stripe's embedded card form.
- **Why:** hosted Checkout is the least code and no card data touches the app.
  The webhook is what marks an order paid, which is how a real integration
  works. `PAYMENT_MODE=simulated` keeps tests and local runs working without
  Stripe keys, and is the fallback if time runs out.

## 10. Invented products with generated SVG covers

- **Alternatives:** a public games API; stock photos.
- **Why:** no licensing questions and no dependency on an external image host.
  The seed script draws each cover (gradient, title, platform badge).

## 11. Project lives in the WSL filesystem

- **Alternatives:** a folder on the Windows drive.
- **Why:** from WSL, `npm install` and the dev server's file watching are much
  slower on `/mnt/c`.

## 12. Design values measured from the live site

- **Alternatives:** matching by eye from screenshots.
- **Why:** computed styles give exact colors, sizes and weights. They are
  recorded in `docs/design-reference.md`. Finding worth noting: Eneba uses
  square corners everywhere and the open-source Metropolis font.

## 13. Go packages in `backend/`, not `internal/`

- **Alternatives:** the conventional `internal/` folder, as first planned.
- **Why:** Go only lets code import an `internal` package from inside the
  parent folder's tree. Vercel compiles `api/index.go` through a wrapper it
  generates itself, and I did not want the first deploy to depend on where
  that wrapper lives. This is a precaution, not a problem I reproduced. The
  module is named `eneba-marketplace` (no domain) because it is an
  application, not a library anyone imports.

## 14. The health check reports a missing database instead of crashing

- **Alternatives:** exit at startup when `DATABASE_URL` is missing or wrong.
- **Why:** `GET /api/health` answers 503 with `{"database": "not configured"}`
  or `"unreachable"`, and the homepage shows it. A wrong connection string on
  the host is then visible in the browser in seconds, where a crashed function
  only shows a generic error. The real error text goes to the server log, not
  to the visitor.

## 15. Metropolis through `@fontsource/metropolis`

- **Alternatives:** a font CDN; Arial as the fallback only.
- **Why:** the package puts the font files in the build, so the site serves
  them itself and makes no request to a third party. Only the four weights the
  design uses are imported (400, 600, 700, 800).

## 16. The shop is called "Keybay"

- **Alternatives:** using Eneba's name and logo.
- **Why:** the brief asks for Eneba's style, and the assignment rules I set
  say not to copy their logo or assets. A text wordmark with a yellow bar
  keeps the look without the trademark. The footer says the site is a demo and
  is not affiliated with Eneba.

## 17. Tiny `.env` loader instead of a library

- **Alternatives:** `godotenv`; exporting variables by hand in the shell.
- **Why:** it is about 20 lines (`backend/config/dotenv.go`) and I can explain
  each one. It only runs in `cmd/server`; on Vercel the variables come from
  the host.
