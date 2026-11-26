-- Homepage: the layout of the page and the data behind each of its blocks.
--
-- The homepage is a list of blocks (banner, deals, best sellers, ...). Which
-- blocks appear, in what order and under what title is stored in home_layout,
-- so the page can be rearranged with an UPDATE and no deploy. Each block then
-- reads from its own source, listed below.

-- A product row is one edition: "Thunder Circuit Kart Steam Key GLOBAL".
-- title is the game itself, "Thunder Circuit Kart", so editions of the same
-- game can be recognised (the New releases block shows each game once).
ALTER TABLE products ADD COLUMN title TEXT;

-- Fill it for the rows that already exist. Every name was built as
-- "<title> <platform> Key <REGION>", so the title is the part before
-- " <platform> Key ".
UPDATE products SET title = split_part(name, ' ' || platform || ' Key ', 1);

ALTER TABLE products ALTER COLUMN title SET NOT NULL;
CREATE INDEX products_title_idx ON products (title);

-- Which blocks the homepage shows, top to bottom.
CREATE TABLE home_layout (
    position    INTEGER PRIMARY KEY,           -- 1 is the top of the page
    block_type  TEXT NOT NULL,                 -- 'banner', 'deals', 'best_sellers', ...
    title       TEXT NOT NULL DEFAULT '',      -- heading above the block, may be empty
    see_all_url TEXT NOT NULL DEFAULT '',      -- optional "See all" link next to the heading
    enabled     BOOLEAN NOT NULL DEFAULT true  -- false hides the block without deleting the row
);

-- Source of the banner block. A banner is text on a gradient, no image file.
CREATE TABLE banners (
    id        BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title     TEXT NOT NULL,
    subtitle  TEXT NOT NULL,
    cta_label TEXT NOT NULL,                   -- the button text
    link      TEXT NOT NULL,                   -- where the button goes, a path on this site
    color_a   TEXT NOT NULL,
    color_b   TEXT NOT NULL,
    position  INTEGER NOT NULL,                -- 1 is the large banner
    active    BOOLEAN NOT NULL DEFAULT true
);

-- Source of the deals block. A deal belongs to one offer. The offer's own
-- price_cents is the price a buyer pays (already lowered); the deal remembers
-- what the price was before and until when the deal is shown.
CREATE TABLE deals (
    id                   BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    offer_id             BIGINT NOT NULL UNIQUE REFERENCES offers (id) ON DELETE CASCADE,
    original_price_cents INTEGER NOT NULL CHECK (original_price_cents > 0),
    starts_at            TIMESTAMPTZ NOT NULL DEFAULT now(),
    ends_at              TIMESTAMPTZ NOT NULL,
    CHECK (ends_at > starts_at)
);

-- The block only ever asks for deals that have not ended yet.
CREATE INDEX deals_ends_at_idx ON deals (ends_at);

-- Source of the best sellers block: how many units of each product were sold.
-- It is seeded with an invented history. From the checkout phase on, paying
-- for an order adds to it.
CREATE TABLE product_sales (
    product_id BIGINT PRIMARY KEY REFERENCES products (id) ON DELETE CASCADE,
    units_sold INTEGER NOT NULL DEFAULT 0 CHECK (units_sold >= 0)
);
