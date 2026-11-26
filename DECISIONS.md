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

## 18. Pushes to `main` deploy to production

- **Alternatives:** deploying by hand with the Vercel CLI only.
- **Why:** `vercel link` connected the GitHub repository, so Vercel builds and
  publishes every push to `main`. That keeps the public URL equal to the
  repository, which is what a reviewer expects. The first deploy (2026-11-26)
  confirmed the two things phase 1 was meant to test: Vercel builds the Go
  function from `api/index.go` with the packages in `backend/`, and the
  `/api/(.*)` rewrite passes the original path to the router, so
  `/api/health` and the JSON 404 both work.

## 19. Our own migration runner

- **Alternatives:** golang-migrate, Atlas, Flyway.
- **Why:** `cmd/migrate` is about 100 lines of standard library plus pgx. It
  applies the `.sql` files in name order, each in a transaction, and records
  them in `schema_migrations`. I can explain every line, which I could not
  say of a tool's internals. What it lacks: rollbacks ("down" files) and a
  lock against two people migrating at once. With one developer neither has
  come up; on a team I would switch to golang-migrate.

## 20. Two connection strings: pooled for the API, direct for schema changes

- **Alternatives:** one pooled string for everything.
- **Why:** Neon's pooler (PgBouncer) hands a server connection to another
  client between transactions, which can break migrations and anything that
  relies on session state. So the API reads `DATABASE_URL` (pooled, right for
  many short serverless connections) and `cmd/migrate` and `cmd/seed` read
  `DATABASE_URL_UNPOOLED` (direct). Both commands refuse to run against a
  host whose name contains `-pooler`, and print which database they are
  about to change.

## 21. Production schema changes follow a fixed order

- **Alternatives:** migrating automatically on deploy.
- **Why:** each migration is applied to the `dev` branch first (a copy of
  production), tested, then applied to production by hand with
  `-env .env.production`, and only then is the code that needs it pushed. A
  deploy that runs before its tables exist would break the live site, and an
  automatic migration on a serverless function has no single moment to run.

## 22. One fixed SQL statement for the store filters

- **Alternatives:** a query builder that adds a `WHERE` line per active
  filter; a query-builder library.
- **Why:** every filter is always in the query as
  `($1 is empty OR column matches $1)`. An unused filter is passed as an empty
  array, an empty string or NULL and lets every row through. The SQL text
  never changes, so there are no placeholder numbers to track and nothing a
  visitor types can become part of the statement. Only `ORDER BY` varies, and
  it is picked from a fixed map of four clauses. The listing runs two queries
  (count, then the page) so a page past the end still reports the total.

## 23. A product's price is its cheapest offer that has stock

- **Alternatives:** the cheapest offer regardless of stock; an average.
- **Why:** it is the price a buyer can pay right now, so it is what the card
  shows ("From"), what the price filter compares and what price sorting uses.
  A product with no stock has no price: it shows "Sold out", is left out when
  a price range is set, and sorts last.

## 24. Covers are drawn by an endpoint from two stored colors

- **Alternatives:** a `cover_svg` column with the whole image (the first
  plan); image files in the repository.
- **Why:** `GET /api/products/{slug}/cover.svg` builds the SVG from the
  product's name, platform and two colors. Embedding the SVG in the listing
  JSON would have made every page of results several times larger. As an
  image URL the browser and the CDN cache it for a day. Text from the
  database is escaped before it goes into the SVG.

## 25. Store filters live in the URL

- **Alternatives:** React state; a state library.
- **Why:** `/store?platform=Steam&sort=price_asc&page=2` can be shared,
  bookmarked and reloaded, and the back button undoes a filter. The page reads
  everything from `useSearchParams` and the same query string is passed to
  the API, which uses the same parameter names.

## 26. A 40-line `useAsync` hook instead of a data-fetching library

- **Alternatives:** TanStack Query, SWR; repeating `useEffect` in each page.
- **Why:** four components load data the same way: loading, ready or error,
  and cancel the request if the input changes. The hook covers that and
  nothing else (no cache, no retries). It stores each result with the key it
  was loaded for, so a slow old response cannot replace a newer one.

## 27. Database tests check properties against the seeded dev database

- **Alternatives:** mocking the database; a throwaway Postgres in Docker;
  asserting exact product names.
- **Why:** the logic worth testing here is the SQL, and a mock would not run
  it. There is no Docker on this machine, and Neon's `dev` branch is free.
  Tests assert rules that hold for any data ("every result is a Steam
  product", "prices never go down", "no product on two pages") by comparing
  the query's answer with the same filter applied in Go. They skip themselves
  when `DATABASE_URL` is not set.

## 28. CI on GitHub Actions, as a signal and not a gate

- **Alternatives:** no CI; CI that also performs the deploy.
- **What it does:** `.github/workflows/ci.yml` runs on every push and pull
  request. The Go job checks formatting (`gofmt`), runs `go vet`, runs
  `govulncheck` (known vulnerabilities in dependencies that our code can
  reach), then migrates and seeds an empty database and runs the tests. The
  frontend job runs `npm ci`, the linter and the build, which type-checks.
- **The database in CI:** a throwaway `postgres:18` container that exists only
  for the job (18 is what Neon runs). It needs no secret, it does not depend
  on Neon being awake, and it proves on every push that the migrations and the
  seed work from nothing. The alternative, a GitHub secret holding the Neon
  `dev` string, would put a database password into GitHub. What the container
  does not cover is Neon's pooler; that path is covered by local test runs,
  which use the pooled string.
- **What it does not do:** it does not stop a deploy. Vercel builds every push
  to `main` by itself and does not wait for GitHub Actions, so a failing run
  shows a red mark on the commit while the site still goes out. For one
  developer who reads the result, that is enough.
- **What a real gate would take:** turn off Vercel's automatic Git deploys,
  store a Vercel token and the project ids as GitHub secrets, and add a deploy
  job that runs `vercel deploy --prod` only after both check jobs pass. About
  15 more lines, and a long-lived token to look after. Worth it on a team.
- **Small choices:** the workflow has read-only access to the repository, a
  newer push cancels the older run, and the Go version is read from `go.mod`.
  Actions are pinned to a major version (`@v7`). Pinning to a commit hash is
  stricter and is what I would do for a repository with secrets.

## 29. Security headers, with a strict Content-Security-Policy

- **Alternatives:** no headers (the state before); a looser policy with
  `'unsafe-inline'`.
- **Where:** a `headers` block in `vercel.json`, applied to every path.
- **The policy, line by line:**
  - `default-src 'self'`: anything not named below may only come from our own
    domain.
  - `script-src 'self'`: only our bundled script runs. Injected inline scripts
    do not, which is the main defense against cross-site scripting.
  - `style-src 'self'`: only our stylesheet. It blocks `<style>` tags and
    `style="..."` attributes written into the HTML. **Correction (phase 3):**
    I first wrote that this also blocks React's `style={...}` prop and moved
    two styles out of `PlaceholderPage` for that reason. That was wrong. React
    applies the prop through the browser's style object, not by writing an
    attribute, and the policy allows that. The banner block relies on it
    (it passes two colors as CSS variables) and was tested under the policy
    with no violations. The `PlaceholderPage` change was unnecessary, though
    harmless, and stays.
  - `img-src 'self' data:`: covers come from `/api/products/.../cover.svg`.
  - `font-src 'self'`: Metropolis is bundled, so no font host is needed.
  - `connect-src 'self'`: the page may only call our own API.
  - `object-src 'none'`, `base-uri 'self'`, `form-action 'self'`: close three
    less common ways of injecting or redirecting content.
  - `frame-ancestors 'none'`: no site may show ours in a frame.
- **The other three:** `X-Content-Type-Options: nosniff` (the browser trusts
  the declared file type), `X-Frame-Options: DENY` (the older form of
  `frame-ancestors`, kept for old browsers) and
  `Referrer-Policy: strict-origin-when-cross-origin` (other sites learn our
  domain but not which page a visitor came from).
- **Testing it before deploy:** the headers exist only on Vercel, so
  `vite.config.ts` reads them from `vercel.json` and sends them in
  `npm run preview`. One copy of the headers, and a local way to see what the
  policy blocks.
- **Later:** Stripe's hosted Checkout is a full-page redirect and needs no
  change. An embedded Stripe form would need Stripe's domains added.

## 30. Time limits on the local server

- **Alternatives:** `http.ListenAndServe`, which has none.
- **Why:** without limits, a client that sends its request one byte at a time
  can hold a connection open forever (a "slowloris" attack). `cmd/server` now
  builds an `http.Server` with `ReadHeaderTimeout` 5s, `ReadTimeout` 10s,
  `WriteTimeout` 30s and `IdleTimeout` 60s. This matters for local runs and
  for Render. On Vercel the platform enforces its own limits and this file is
  not used.

## 31. `migrate -status` shows what would run without changing anything

- **Alternatives:** trusting the env file name.
- **Why:** before a production run I want to see which database the command
  is pointed at and which files are pending. `-status` connects, prints both
  and exits. It does not even create the bookkeeping table.

## 32. The homepage is assembled from a layout the backend sends

- **Alternatives:** a fixed page in React; one endpoint returning the layout
  and all block content together.
- **Why:** `GET /api/home` returns the blocks in page order (type, title,
  "see all" link, endpoint), read from the `home_layout` table. The frontend
  holds one map from block type to component and nothing else about the page.
  Reordering, renaming or hiding a block is an `UPDATE`, with no deploy. Each
  block then fetches its own endpoint, so blocks load side by side and one
  failing leaves the others standing. Unknown types are skipped on both
  sides: the backend drops types it cannot serve, the frontend drops types it
  cannot draw. This is the brief's "each block gets its content from its own
  backend service or database, nothing is hardcoded".

## 33. Best sellers reads its own table, `product_sales`

- **Alternatives:** summing paid orders, as first planned.
- **Why:** orders do not exist until the checkout phase, and the block should
  not wait for it. `product_sales` holds units sold per product. It is seeded
  with invented numbers; from the checkout phase on, paying for an order will
  add to it in the same transaction. The block's query is a plain read either
  way, and sold-out products are left out because a best seller nobody can
  buy does not belong on the front page. **The seeded figures are made up.**

## 34. A deal lowers the offer's real price

- **Alternatives:** keeping the offer's price and applying the discount when
  a price is shown.
- **Why:** a deal row points at one offer and stores the price it had before
  (`original_price_cents`) and when the deal ends. The offer's own
  `price_cents` is the lowered price. So there is one price everywhere: store
  card, product page, checkout. A discount applied at display time would have
  to be repeated in every query that touches a price, and one missed place
  would show a different number than the buyer pays.
- **Limitation:** when a deal ends, the block stops showing it but nothing
  raises the price again. That would need a scheduled job, which the free
  tier and this project's size do not justify.
- **The deals are invented**: no price was ever higher.

## 35. Deals end on real dates, spread over 60 days

- **Alternatives:** deals with no end; deals that all end the same week.
- **Why:** a countdown needs a real end time, and a real end time means the
  deal expires. The seed spreads eight deals from 5 to about 60 days out, so
  the block is not empty when the project is reviewed. Once all have ended
  the block has no items and the page leaves it out. `cmd/seed -reset`
  creates fresh ones.

## 36. A `title` column, so each game appears once in New releases

- **Alternatives:** showing the newest product rows as they are.
- **Why:** a product row is one edition (game, platform, region). The newest
  rows are the same game two or three times. `products.title` holds the game
  itself, and the block uses `DISTINCT ON (title)` to keep one edition per
  game: the cheapest one that has stock. The migration fills the column for
  existing rows from the name. The cover drawing now reads the title from the
  column, where before it cut it out of the name.

## 37. A static banner layout, with gradients and no images

- **Alternatives:** an auto-rotating carousel; banner image files.
- **Why:** a carousel that moves by itself changes while people read it and
  needs pause controls to be accessible. One large panel and two small ones
  show three promotions at once with no script. Banners are text on a
  gradient whose two colors come from the `banners` table, so there are no
  image assets to license, store or cache. A dark layer over the gradient
  keeps white text readable whatever colors are stored.

## 38. Homepage responses are cached at the edge for 60 seconds

- **Alternatives:** no caching; caching in the browser.
- **Why:** the homepage makes six API calls. `Cache-Control: public,
  max-age=0, s-maxage=60` lets Vercel's edge keep each answer for a minute
  and serve it without starting the function or waking the database.
  `max-age=0` keeps browsers from holding their own copy. The cost: a change
  to the homepage tables can take up to a minute to appear.

## 39. A failed or empty block draws nothing

- **Alternatives:** an error message inside the block.
- **Why:** on a shop's front page a missing row is better than an error box.
  If the layout itself fails to load, the page says so and links to the
  store. The failed request is still visible in the browser's network tab.

## 40. Vitest for frontend unit tests

- **Alternatives:** no frontend tests yet; Jest.
- **Why:** the countdown text ("3d 4h", "12m", ended) is logic with edges, and
  it belongs in a test. Vitest uses the Vite setup the project already has,
  so it needs no configuration. It is a development dependency only. CI runs
  it in the frontend job.

## 41. Covers shrink long titles instead of clipping them

- **Alternatives:** a fixed font size (the first version).
- **Why:** looking at a screenshot of the homepage showed titles such as
  "WITCHLIGHT" cut off at the cover's edge. The font size is now the largest
  one, up to 30, at which the longest line fits. The store page had the same
  flaw since phase 2. Browsers keep a cover for a day, so a returning visitor
  may see the old drawing until then.

