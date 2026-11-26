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
- **One card per game (changed before release).** The first version ranked
  product rows, and since a product is one edition, the row showed "Quiet
  Harbor" twice (two regions of the Xbox edition). Now the query keeps one
  edition per title, the one that sold the most among those in stock, and
  ranks those. It is the same `DISTINCT ON (title)` approach as New releases
  (entry 36), with a different rule for which edition wins: most sold here,
  cheapest there.

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

## 42. Search lives in `/api/products`, not in a separate `/api/search`

- **What PLAN.md said:** two endpoints, `GET /api/search?q=` for ranked
  results and `GET /api/search/suggest?q=` for the dropdown.
- **What was built:** search results come from `GET /api/products?q=...`.
  Only `/api/search/suggest` is separate.
- **Why:** a search result page is the store listing with one more condition.
  It needs the same platform, region and price filters, the same paging and
  the same card data. A second endpoint would have meant the same query
  twice, kept in step by hand. With one endpoint the search line is one more
  `AND` in the listing's `WHERE`, and every filter works with it for free.
  The store page's own "Name" field got typo tolerance at the same time,
  with no extra work.
- **The suggest endpoint stays separate** because it answers a different
  question: the best six matches, no filters, no count, cached at the edge.
  It reuses the same matching and ranking code (`queryPage`).

## 43. Each word of a search is matched on its own

- **Alternatives:** comparing the whole text with the whole name; PostgreSQL
  full-text search; Neon's BM25 and vector search extensions.
- **Why:** a product matches when every word matches its name, either
  because the name contains the word or because the word is similar to part
  of the name. This is what makes "kart xbox" find "Thunder Circuit Kart Xbox
  Key", where the words are not adjacent, and what makes "thunder zzzz" find
  nothing. Full-text search matches whole words and does nothing for typos
  or half-typed words. Neon's search extensions rank documents and exist
  only on Neon, so the CI container could not run the tests.
- **Only letters and digits reach the query.** The text is split wherever
  there is anything else, so `%` and `_`, which are wildcards in a `LIKE`
  pattern, never get there. A search for "%" is a search with no words and
  finds nothing.
- **Ranking (the order of results), best first:** the name starts with the
  whole text; the name contains it; the name contains every word; a word
  matched only by similarity. Inside each group, products that can be bought
  come before sold-out ones, then the closer match, then units sold.
- **In stock before sold out was added after release.** On the live site
  "witchlite" listed a sold-out edition first, because inside a group the
  order went straight to units sold. A shopper wants what they can buy, so
  stock now comes first within each group. It does not cross groups: an exact
  match that is sold out still ranks above a typo match that is in stock.

## 44. Similarity threshold 0.45, and no typo matching under 4 letters

- **How it was chosen:** I measured `word_similarity` for about 45 words
  against the seeded names. Real typos scored 0.50 or more: "legnds" 0.50,
  "obsidain" 0.56, "raly" 0.60, "witchlite" 0.70, "circut" 0.71. Unrelated
  words scored 0.40 or less: "kart" against "Garage" 0.40, "tied" against
  "Tide" 0.40, "racing" against "Rally" 0.29. 0.45 sits in the gap.
- **The 4-letter rule:** a 3-letter word has four trigrams, and one wrong
  letter changes most of them, so its similarity to the right word ("vwo"
  to "Vow": 0.25) is no higher than to wrong ones. Short words must be
  contained exactly.
- **Known limits, each pinned by a test:** "vwo" (typo in a 3-letter word),
  "krat" and "rouge" (swapped letters in a short word, 0.20 and 0.33) and
  "glbal" (0.44, just under) find nothing. Search looks at names only, so
  "racing" does not find a racing game.
- **The threshold fits this catalog.** It was tuned on 24 invented titles. A
  real catalog would need the measurement repeated.

## 45. No trigram index, although the plan promised one

- **What the plan said:** a GIN trigram index on `products.name`.
- **What was built:** the extension only.
- **Why:** an index that the query cannot use is dead weight, and this query
  cannot use one. It calls `word_similarity(word, name) >= 0.45` as a
  function inside a per-word check, and PostgreSQL only uses a trigram index
  for the operator forms (`%`, `<%`, `LIKE`) written directly against the
  column. With 56 rows, reading the whole table is no work at all, so
  nothing is lost today. I did not measure it, because there is no indexed
  version to compare against.
- **What a large catalog would need:** the per-word condition rewritten with
  the `<%` operator and `ILIKE` directly on `name` (one fixed slot per word),
  the threshold set with `SET LOCAL pg_trgm.word_similarity_threshold` inside
  a transaction, and then the GIN index. I chose the readable query for the
  size this project has.

## 46. The results page is the store page

- **Alternatives:** a separate search results component.
- **Why:** `/search?q=...` renders `StorePage` with `mode="search"`. It
  changes the heading ("Results for ..."), adds a "Best match" sort that is
  the default, keeps the search text when filters are cleared, and on no
  results shows a hint and the best sellers row. Grid, filters and paging are
  the same code, so they cannot drift apart.
- **No "did you mean":** with a typo the page shows the matches directly.
  Printing a corrected word would need a second step that decides which word
  was meant.

## 47. The search box: debounce, cancel, and only fresh suggestions

- **Debounce:** `debounce()` (about 20 lines, in `lib/debounce.ts`, with tests)
  waits until typing has paused for 250 ms. Typing "thund" sends one request.
- **Cancel:** `useAsync` aborts the previous request when the text changes,
  so a slow answer for "thu" cannot replace the answer for "thunder".
- **Fresh only:** the list is shown only when it belongs to what is in the
  box right now. Without that rule the old suggestions stayed visible for a
  moment under new text, with a "See all results" line that did not match.
- **Keyboard:** arrow keys move through the options and wrap around, Enter
  opens the highlighted option or the results page, Escape closes the list
  and keeps the text. The markup follows the ARIA combobox pattern, so a
  screen reader announces the highlighted suggestion.
- **Two bugs the browser test caught:** Enter on a highlighted suggestion
  went to the results page, because pressing Enter in a form also activates
  its submit button and that button reset the highlight (fixed by making it a
  plain button). And Escape cleared the text, which is a search field's
  default behavior (now prevented).

## 48. The filter panel no longer rebuilds on every URL change

- **Before:** the panel was keyed by the whole query string, so ticking a
  platform rebuilt it. Keyboard focus was lost on every tick.
- **Now:** it is keyed only by the three text values (name, minimum and
  maximum price). Those inputs keep their own state while typing and need a
  rebuild when the URL changes from outside. Checkboxes read the URL directly
  and need none.

## 49. Passwords: bcrypt at cost 12, length rules only

- **Alternatives:** argon2id; scrypt; a plain or salted SHA hash (never).
- **Why bcrypt:** it is slow on purpose and salts every hash by itself, so a
  stolen `users` table cannot be guessed through quickly and two users with
  the same password do not get the same hash. argon2id is the newer first
  recommendation, but it needs memory settings tuned to the host, and a
  serverless function's memory is small. bcrypt has one number to explain.
- **Rules:** at least 8 characters, at most 72 bytes. bcrypt reads only the
  first 72 bytes, so a longer password is refused: cutting it silently would
  mean two different passwords open the same account. No "must contain a
  symbol" rule, because length is what makes a password hard to guess.
- **Never exposed:** the `User` type the rest of the program passes around
  has no hash field, so the hash cannot leak into a response by accident.
- **Shortcut:** no check against lists of leaked passwords.

## 50. The session cookie and what is stored for it

- **The cookie:** `__Host-session`, with `HttpOnly` (JavaScript cannot read
  it), `Secure` (HTTPS only), `SameSite=Lax` (not attached to other sites'
  background requests), `Path=/`, no `Domain`, 7 days. The `__Host-` prefix
  makes the browser itself refuse the cookie unless it is Secure, host-only
  and site-wide.
- **The token:** 32 bytes from the system's secure random source. Nothing is
  encoded in it.
- **In the database:** only the SHA-256 hash of the token. A fast hash is
  right here, unlike for passwords, because 32 random bytes cannot be
  guessed however fast the hash is. A copy of the `sessions` table does not
  let anyone sign in.
- **Alternatives rejected:** a JWT in `localStorage` (readable by any script
  on the page) and a self-contained JWT cookie (cannot be revoked before it
  expires, so "sign out" would not really end anything).
- **Local development:** plain `http://localhost` cannot use `Secure` or the
  prefix. `INSECURE_COOKIES=1` in `.env` switches both off. The default is
  the secure form, and the variable is not set on Vercel.

## 51. CSRF: three layers and no token

- **The attack:** another site makes the visitor's browser send a request
  here, with their cookie attached.
- **Layer 1:** `SameSite=Lax` keeps the cookie off other sites' POSTs.
- **Layer 2:** every route that changes something is wrapped in `sameOrigin`,
  which refuses a request whose `Origin` header is missing or names another
  host. Browsers set that header themselves and a page cannot fake it.
- **Layer 3:** those routes accept only `application/json`. A plain HTML form
  cannot send that, and a script can only try after a preflight request that
  this server never approves.
- **No GET changes anything.** Sign out is a POST.
- **Why no CSRF token:** with one domain for frontend and API and the three
  layers above, a token would add state to keep in sync and stop nothing
  more.
- **A consequence for local development:** the dev proxy must pass the
  browser's `Host` header through unchanged (`changeOrigin: false` in
  `vite.config.ts`). Vite's short proxy form rewrites it, and then every
  sign-in is refused as cross-site. The browser test found this.

## 52. Sessions end on the server

- **Expiry:** `expires_at` is 7 days after sign in, and it is checked in the
  query on every request. An expired session stops working even if the
  browser still sends its cookie.
- **Sign out:** deletes the session row and clears the cookie. A copied token
  is dead from that moment.
- **A new token at every sign in,** so a token planted beforehand is useless.
- **Cleanup:** a user's expired rows are deleted when they sign in. No
  scheduled job.
- **`Cache-Control: no-store`** on every response about an account.
- **Shortcuts:** no "sign out everywhere", no list of active sessions, and
  the 7 days neither extend with use nor shorten when idle.

## 53. Sign-in attempts are counted in the database

- **Why the database:** serverless function instances share no memory.
- **Limits:** 5 failed sign-ins per email in 15 minutes; 30 failed sign-ins
  per IP address in 15 minutes; 10 sign-ups per IP address per hour. Over a
  limit the answer is 429.
- **Limits apply to any email, registered or not,** so being limited does not
  reveal whether an account exists.
- **A successful sign-in does not reset the count,** so an attacker cannot
  clear it with an account of their own.
- **The IP address** is the first entry of `X-Forwarded-For`. Vercel
  overwrites that header and does not pass on what a visitor sent, so it
  cannot be forged there (Vercel's documentation says so; I did not test it
  myself). On a host that passes the header through, `clientIP` would need
  changing.
- **Gaps:** anyone who knows an email can lock its sign-in for 15 minutes by
  failing 5 times. No CAPTCHA, no growing delays. Old rows are deleted when a
  later attempt is recorded, not on a schedule.

## 54. The demo account is not locked per email

- **The problem:** the demo account's password is printed in the README. With
  the per-email limit, anyone could lock it for 15 minutes by failing 5
  times, and a reviewer could run into that.
- **What was built:** `users.is_demo` marks the account, and sign-in skips the
  per-email limit for it. The per-IP limit still applies.
- **Why that is safe:** the limit exists to stop password guessing, and this
  password is public. Locking the account protects nothing.
- **Also:** the README tells reviewers they can create their own account
  with any email, which does not depend on the demo account at all.

## 55. Sign-in does not reveal who has an account; sign-up does

- **Sign-in:** an unknown email and a wrong password get the same status and
  the same message. They also do the same work: for an unknown email the
  password is still compared against a fixed dummy bcrypt hash, so the
  response time does not give it away. The dummy hash is a constant in the
  code and not computed at startup, which would slow every cold start.
- **Sign-up:** an email that already has an account is answered with "an
  account with that email already exists". That does reveal it.
- **Why the shortcut:** the proper pattern answers "check your inbox" either
  way and lets an email tell the owner. This project sends no email.
  Pretending success would leave a real person unable to understand why they
  cannot sign in. The per-IP sign-up limit keeps this from being used to
  test a long list of addresses quickly.
- **Uniqueness is enforced by the database,** not by a `SELECT` first. Two
  sign-ups for the same address at the same instant cannot both succeed.

## 56. The frontend asks who is signed in; it never holds the session

- **How:** `AuthProvider` calls `GET /api/auth/me` when the page loads and
  keeps the answer. The cookie is invisible to JavaScript, so there is no
  token in React state or `localStorage` to steal.
- **`RequireAuth` is a convenience, not the protection.** It sends a
  signed-out visitor to the sign-in page. The data behind a protected page is
  refused by the backend's `requireUser` without a valid session, so
  bypassing the component in the browser gains nothing.
- **Returning after sign-in:** the wanted address travels in `?next=`. It is
  accepted only if it is a path on this site (`safeNextPath`, with tests), so
  a crafted link cannot bounce someone to another site right after they
  typed their password.

## 57. Order and key states, and what enforces them

- **A key** is `available`, `reserved` (held for one order item) or `sold`.
  It moves available → reserved → sold, or back from reserved to available
  when the order fails or runs out of time. A sold key never changes.
- **An order** is `pending` (keys reserved, a deadline 15 minutes away),
  `paid`, `failed` (payment declined) or `expired` (deadline passed). The
  three end states never change.
- **The database enforces two rules itself:** an available key belongs to no
  order item and a reserved or sold key always belongs to one
  (`keys_status_matches_owner`), and an order is paid exactly when it has a
  payment time. No code path can leave a key half assigned.
- **Reservations last 15 minutes,** where `PLAN.md` said 30. Stock held by
  abandoned orders comes back sooner.

## 58. A key cannot be sold twice: `FOR UPDATE SKIP LOCKED`

- **The statement** (`reserveKeysSQL` in `backend/checkout/orders.go`): an
  `UPDATE` whose inner `SELECT` picks available keys of the offer and locks
  those rows until the transaction ends. `SKIP LOCKED` makes a second
  transaction pass over rows another one has locked and look for others. Two
  transactions can therefore never pick the same row.
- **Alternatives:** the strictest isolation level (`SERIALIZABLE`), which the
  Go skill suggests for inventory; a stock counter on the offer with a
  `CHECK (stock >= 0)`; an application-level lock.
- **Why this one:** serializable transactions fail whenever they overlap and
  have to be retried, which needs a retry loop around every order. A stock
  counter says how many are left but not which key each buyer gets, and here
  each buyer needs a specific key. A lock in the application does not work
  across serverless instances. The row lock says exactly which rows are
  protected and needs nothing else.
- **All or nothing:** creating an order is one transaction. If any line
  cannot be filled, the whole order is rolled back, including the lines that
  could. Lines are sorted by offer id first, so two orders wanting the same
  two offers always take them in the same order.

## 59. The trade-off of `SKIP LOCKED`: "sold out" can be premature

- **What can happen:** two buyers go for the last key at the same moment. The
  first one's transaction holds it. The second is told "not enough stock"
  immediately. If the first buyer then has the payment declined, or never
  pays and the order expires, the key comes back, and the second buyer was
  turned away from a key that ended up unsold.
- **Why it is accepted:** the answer was true when it was given, since the
  key was held by someone else. The buyer is told at once and can try again
  a little later. Nothing is oversold, which is the property that must not
  break.
- **The alternative** is plain `FOR UPDATE` without `SKIP LOCKED`: the second
  buyer's request waits until the first transaction ends and then sees the
  result. That only covers the split second in which the first order is
  being created. The first buyer's reservation lasts up to 15 minutes after
  that, and no request can wait that long. To really close the gap a shop
  needs a waiting list: record the second buyer's interest and offer them
  the key when it is released. That is a feature of its own (a queue, a
  notification, a time limit on the offer) and is not built here.
- **A smaller step that would help:** tell the refused buyer how many keys
  are held by unpaid orders, so "try again in a few minutes" is an informed
  suggestion. Not built.

## 60. Expiry is cleaned up through the order row

- **The problem:** an unpaid order does not release its keys by itself, and
  the free tier has no scheduled jobs.
- **What was built:** `releaseExpiredSQL` runs at the start of every order
  creation and every payment. One statement marks pending orders past their
  deadline as expired and releases the keys of exactly those orders.
- **Why it goes through the order row:** a payment locks its order row while
  it works. The cleanup selects orders with `FOR UPDATE SKIP LOCKED`, so it
  passes over an order that is being paid at that moment. A key is therefore
  never released while it is being sold. That order's fate is decided by the
  payment, which checks the deadline itself after taking the lock.
- **What the buyer sees:** reading an order reports "pending and past its
  deadline" as expired, so the page is right even before any cleanup ran.
- **Alternative:** a cron job. More timely, but one more moving part, and the
  result for a buyer is the same.

## 61. Paying is safe to repeat, and is the only place an order is fulfilled

- **`Pay` locks the order row first.** A second call for the same order (a
  double click, two tabs, a webhook delivered twice) waits, then finds the
  order already paid and changes nothing. Keys are sold once and the sale is
  counted once.
- **One fulfilment step:** `markPaid` marks the keys sold, the order paid and
  adds the quantities to `product_sales` for the best sellers block. The
  simulated payment calls it now. The Stripe phase will call the same
  function from its webhook, and never from the page the buyer lands on.
- **A last check inside it:** the number of keys just sold must equal the
  quantity ordered. If not, the transaction fails and rolls back, which is
  better than handing over fewer keys than were paid for.
- **Payment is simulated with two buttons, not a card form.** A form that
  looks like it wants a card number invites someone to type a real one.

## 62. What the buyer is protected from

- **A price they did not see:** the browser sends offer ids and quantities
  only, and the server reads every price itself. The browser also sends the
  total it showed. If the server's total differs, the order is refused with
  the new prices and nothing is reserved.
- **Other people reading their keys:** an order belongs to its user. Someone
  else's order is answered with 404, the same as one that does not exist, so
  the answer does not confirm it exists. Keys are selected only for a paid
  order: both conditions are in the SQL, not in an `if` that could be
  forgotten.
- **Guessable order addresses:** order ids are random UUIDs, not 1, 2, 3.
- **One account holding all the stock:** at most 5 of an offer per order, 10
  lines per order and 3 unpaid orders per user. This is a soft limit: two
  orders placed by one user at the same instant could both pass the count.

## 63. The cart lives in the browser and holds no prices

- **What is stored:** offer ids and quantities, in `localStorage`. Names,
  prices and stock are asked from the server every time the cart is shown.
- **Why:** the cart has to exist before sign-in, and prices kept in the
  browser would be stale and would invite tampering. Editing the saved cart
  by hand can change nothing that the server does not check.
- **`parseCart` trusts nothing it reads back:** a damaged or hand-edited
  saved cart is reduced to the lines that make sense.
- **The cart is emptied only when an order is paid,** so a declined payment
  or an expired order does not cost the buyer their selection.

## 64. The countdown runs on the server's clock

- **The bug:** the order page first compared the order's deadline with the
  device's own clock, and showed 15:04 for a 15-minute reservation because
  this PC's clock was a few seconds behind the database's. A device an hour
  off would have shown the order as expired at once.
- **The fix:** the API sends `seconds_left`, worked out by the database. The
  page counts down from that length of time and never compares two clocks.

## 65. Concurrency tests, and how I know they can fail

- **The tests** (`backend/checkout/concurrency_test.go`) run against a real
  database, locally and in CI:
  1. 20 buyers, 1 key: exactly 1 order, 19 refused.
  2. 20 buyers wanting 2 each, 5 keys: exactly 2 orders, nobody holds a
     partial order.
  3. One order paid 10 times at once: sold once, counted once.
  4. A payment and the cleanup colliding at the deadline, 12 rounds with
     shifting timing: every round ends either paid with its keys sold or
     expired with its keys released, never a mixture.
  They check the tables afterwards, not only what the calls returned.
- **They were shown to fail.** With the reservation swapped for a naive
  version (read the available keys, then update them, no lock), test 1
  reported 20 orders for 1 key and test 2 reported 20 orders for 5 keys.
- **My first version of test 1 could not fail,** and I only found out by
  doing that swap. A connection pool opens its connections one at a time, so
  one buyer got a connection about 200 ms before the others and finished
  before they started. The test passed without two transactions ever
  meeting, even with the naive query. `warmUp` now opens every connection
  before the gate is released.
- **What the tests are and are not:** they are strong evidence that the
  property held in the runs made (tests 1 to 3 passed 10 times each in a
  row). They are not a proof. The guarantee itself is PostgreSQL's row lock.
- **`-p 1`:** the Go packages' tests share one database. The checkout tests
  add a temporary product while they run, and the catalog tests count
  products, so packages are run one at a time.
- **The race detector** (`go test -race`) runs in CI. It could not run on
  this PC, which has no C compiler.

## 66. Shortcuts in checkout

- Keys are stored as plain text in the database. A real shop would encrypt
  them at rest.
- No email with the keys. They stay on the order page.
- No refunds, no cancel button, no invoices, no tax, no guest checkout.
- The simulated payment endpoint lets the buyer choose the outcome. That is
  the point of a simulation, and the Stripe phase replaces it.

## 67. The live demo restocks itself

- **The problem:** on the public site paying costs nothing (payment is
  simulated, or uses free test cards once Stripe is in), and the demo
  account's password is published. Anyone could buy every key and leave the
  store sold out before the project is reviewed.
- **What was built:** when a sale leaves an offer with fewer than 3 available
  keys, the same transaction adds 5 newly generated ones
  (`backend/checkout/restock.go`). It runs only where `DEMO_RESTOCK=1` is
  set, which is the Vercel project and nowhere else.
- **Why this and not something else:**
  - A cap on purchases per account does nothing: anyone can register another.
  - Making the demo account's purchases free of consequences would put a
    special case in the middle of the checkout logic.
  - Reseeding by hand before the interview works, but it deletes every
    order, including a reviewer's, and has to be remembered.
- **What it keeps:** stock still goes down after a purchase, reserving the
  last key still behaves correctly, and the three products seeded as sold
  out stay sold out, because nothing can be sold from them.
- **What it costs:** a real shop cannot invent stock, so this is a demo
  device and is described as one in the README. The tests that check exact
  stock run with it off, which is the default. It has its own tests, which
  switch it on.
- **Two payments for the same offer at the same moment** could both top it
  up, adding 10 keys where 5 were meant. That is harmless here.

