-- Catalog: what the store sells and who sells it.
--
-- A product is one edition of a game for one platform and region, for example
-- "Ashen Crown Steam Key GLOBAL". Several sellers can offer the same product
-- at different prices. Each offer owns a pile of keys, and the number of
-- available keys is the offer's stock.

CREATE TABLE products (
    id            BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    slug          TEXT NOT NULL UNIQUE,          -- used in URLs: /product/ashen-crown-steam-key-global
    name          TEXT NOT NULL,
    description   TEXT NOT NULL,
    platform      TEXT NOT NULL,                 -- 'Steam', 'PlayStation', ...
    region        TEXT NOT NULL,                 -- 'Global', 'Europe', ...
    genre         TEXT NOT NULL,
    release_date  DATE NOT NULL,
    cover_color_a TEXT NOT NULL,                 -- the generated cover is a gradient
    cover_color_b TEXT NOT NULL,                 -- between these two colors
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- The store filters by these two columns.
CREATE INDEX products_platform_idx ON products (platform);
CREATE INDEX products_region_idx ON products (region);

CREATE TABLE sellers (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name        TEXT NOT NULL UNIQUE,
    rating      NUMERIC(2, 1) NOT NULL CHECK (rating BETWEEN 0 AND 5),
    sales_count INTEGER NOT NULL DEFAULT 0 CHECK (sales_count >= 0)
);

CREATE TABLE offers (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    product_id  BIGINT NOT NULL REFERENCES products (id) ON DELETE CASCADE,
    seller_id   BIGINT NOT NULL REFERENCES sellers (id) ON DELETE CASCADE,
    price_cents INTEGER NOT NULL CHECK (price_cents > 0),  -- money is whole cents, never a float
    UNIQUE (product_id, seller_id)                         -- one offer per seller per product
);

-- UNIQUE (product_id, seller_id) already gives an index that starts with
-- product_id, which serves "all offers of a product". This one serves the
-- other direction and the foreign key.
CREATE INDEX offers_seller_idx ON offers (seller_id);

CREATE TABLE keys (
    id       BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    offer_id BIGINT NOT NULL REFERENCES offers (id) ON DELETE CASCADE,
    code     TEXT NOT NULL UNIQUE,
    status   TEXT NOT NULL DEFAULT 'available'
             CHECK (status IN ('available', 'reserved', 'sold'))
);

-- Stock is "how many available keys does this offer have". A partial index
-- holds only the available rows, so it stays small as keys get sold.
CREATE INDEX keys_available_idx ON keys (offer_id) WHERE status = 'available';
