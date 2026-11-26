-- Checkout: orders, their items, and which order item holds which key.
--
-- An order starts as 'pending' with its keys reserved and a deadline. It ends
-- in exactly one of three states and never changes again:
--   paid     payment succeeded, the keys are sold and shown to the buyer
--   failed   payment was declined, the keys were released
--   expired  the deadline passed without payment, the keys were released

CREATE TABLE orders (
    -- A random UUID, not 1, 2, 3: the id appears in the address of the order
    -- page, and a counter would invite trying the next number and would show
    -- how many orders exist. gen_random_uuid() is built into PostgreSQL.
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     BIGINT NOT NULL REFERENCES users (id),
    status      TEXT NOT NULL DEFAULT 'pending'
                CHECK (status IN ('pending', 'paid', 'failed', 'expired')),
    total_cents INTEGER NOT NULL CHECK (total_cents > 0),
    -- Until when the keys are held for this order. Only meaningful while the
    -- order is pending.
    expires_at  TIMESTAMPTZ NOT NULL,
    -- The payment provider's id for this payment. Empty for simulated
    -- payments; the Stripe phase fills it.
    payment_ref TEXT,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    paid_at     TIMESTAMPTZ,
    -- An order is paid exactly when it has a payment time.
    CHECK ((status = 'paid') = (paid_at IS NOT NULL))
);

-- "My orders, newest first."
CREATE INDEX orders_user_idx ON orders (user_id, created_at DESC);

-- The cleanup looks for pending orders past their deadline. A partial index
-- holds only the pending ones, which are few at any moment.
CREATE INDEX orders_pending_idx ON orders (expires_at) WHERE status = 'pending';

CREATE TABLE order_items (
    id               BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    order_id         UUID NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    offer_id         BIGINT NOT NULL REFERENCES offers (id),
    -- Copies of what the buyer saw. An order is a record of a sale, so it
    -- must not change when a product is renamed or an offer repriced later.
    product_slug     TEXT NOT NULL,
    product_name     TEXT NOT NULL,
    seller_name      TEXT NOT NULL,
    unit_price_cents INTEGER NOT NULL CHECK (unit_price_cents > 0),
    quantity         INTEGER NOT NULL CHECK (quantity BETWEEN 1 AND 5)
);

CREATE INDEX order_items_order_idx ON order_items (order_id);

-- Which order item holds a key. NULL while the key is available.
ALTER TABLE keys ADD COLUMN order_item_id BIGINT REFERENCES order_items (id);

-- The rule that makes the states trustworthy, enforced by the database
-- itself: an available key belongs to no order, and a reserved or sold key
-- always belongs to one. No code path can leave a key half assigned.
ALTER TABLE keys ADD CONSTRAINT keys_status_matches_owner
    CHECK ((status = 'available') = (order_item_id IS NULL));

-- "The keys of this order item", for showing them and for releasing them.
CREATE INDEX keys_order_item_idx ON keys (order_item_id) WHERE order_item_id IS NOT NULL;
