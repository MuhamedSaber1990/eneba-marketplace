# Design reference

Values read from eneba.com on 2026-11-26 with the browser's computed styles
(`getComputedStyle`), at a 1440 x 900 window. They become
`frontend/src/styles/tokens.css` in phase 1. Nothing here is an Eneba asset:
these are measurements, and the app uses its own wordmark and generated covers.

Pages measured: homepage (`/`), store (`/store/all`), one product page.

## Colors

| Role | Value | Where seen |
|---|---|---|
| Page background | `#4618AC` | `body` on every page |
| Card panel (text area under a cover) | `#1F0A4D` | product cards, 49 uses on the homepage |
| Raised panel | `#351183` | product page description box |
| Accent purple | `#7A4EDF` | badges, side panels on the product page |
| Primary button | `#FAD318`, black text | "Add to cart", "Accept all" |
| Mint | `#63E3C2` | small badges on cards |
| Region text | `#23C299` | "GLOBAL" on cards, uppercase |
| Lime text | `#84E916` | 36 uses on the homepage; role to confirm from screenshots |
| Text | `#FFFFFF` | default |
| Text, soft | `rgba(255,255,255,0.85)` | small labels on covers |
| Text, nav | `#DCD8E6` | category bar links |
| Text, muted | `#B3AAC9` | "From" above a price, filter inputs |
| Danger | `#EB002E` | one badge on the homepage |
| Overlay | `rgba(0,0,0,0.55)` | strip across the bottom of a cover |
| Input background | `rgba(0,0,0,0.2)` | store filter inputs |
| Footer background | `rgba(0,0,0,0.7)` | footer, over the purple page |

## Typography

Font family: `metropolis, Arial, Helvetica, sans-serif`, weights 400, 600 and
800 loaded (headings ask for 700). Metropolis is an open-source font, so the
app can self-host it.

| Use | Size / weight / line height |
|---|---|
| Body | 16px / 400 / 24px |
| Section heading (h2 on home, h1 on store) | 30px / 700 / 37.5px, margin 10px 0 30px |
| Product page title (h1) | 26px / 700 / 32.5px |
| Product page section heading | 18px / 700 |
| Header search input | 19px / 800 |
| Category bar link | 16px / 600, padding 10px |
| Log in / Register | 14px / 600 |
| Button label | 15px / 800 |
| Card title | 12px / 800 / 18px, two lines |
| Card region | 12px / 800, uppercase |
| Card "From" | 12px / 800, muted |
| Card price | 17.6px / 800 |
| Small label on cover | 10px / 800 |

## Shape and spacing

- **Square corners.** Buttons, cards, inputs and panels all have
  `border-radius: 0`. Only a handful of small elements use 2-5px.
- No borders or shadows on cards. Separation comes from the dark panel color
  against the purple page.
- Content width: `max-width: 1240px` with `padding: 0 20px` (1200px of content).
- Grid gap: 20px.
- Panel padding: 20px.

## Components

**Header.** Two rows inside the 1200px container, transparent over the page
color. Row 1 is 57px high: wordmark, search, language and currency, log in,
register, cart. Row 2 is 50px high: a centered bar of category links.

**Search box.** 550 x 57px, transparent, `0.8px solid #FFFFFF` border, no
radius, search icon on the left (input has 40px left padding). Placeholder:
"Search for games, top-ups and more".

**Buttons.** Height 35px, padding 0 10px, label 15px / 800.
- Primary: `#FAD318` background, black text, no border.
- Secondary: transparent, white text, `0.8px solid #FFFFFF`.
- Text link: `#FAD318` text, 12-14px / 400.

**Product card.** 224px wide.
- Homepage variant: cover 224 x 314, total height 458, with "Add to cart"
  (primary) and "View offers" buttons.
- Store variant: 225 x 440, cover with a 22px `rgba(0,0,0,0.55)` strip along
  its bottom edge, then a 230px `#1F0A4D` panel.
- Panel content, top to bottom: title (two lines), region in green uppercase,
  "From" in muted text, price.

**Store layout.** Filter column 220px wide on the left, then a 20px gap, then a
960px grid of 4 columns x 225px with 20px gaps, 20 products per page. Filters
seen: price range (two inputs), country, product type, operating system,
platforms, each option with a count. Sort options: price low to high, price
high to low, newest, oldest, most popular, A-Z, Z-A.

**Product page.** Cover 160 x 240 on the left, title to its right. Sections in
order: Recommended offers, Other offers, Gamers also viewed, Product
description (in a `#351183` panel with 20px padding), Other details. A 350px
column on the right holds `#7A4EDF` panels. "Load more" under the offers list
is a secondary button.

**Homepage sections seen** (for block ideas, the content is ours): a row of
tall promo banners (144 x 285 each), Popular gift cards, Recommended for you,
Top games, best sellers rows, Upcoming games, Categories.

## Not measured yet

- Cart and checkout pages (they need items in a cart).
- Hover and focus states.
- Phone widths.

Screenshots from the candidate will cover these in phase 8.
