-- Payments through a provider (Stripe).
--
-- orders.payment_ref, added earlier, holds the provider's id for the payment
-- page of an order (a Stripe Checkout Session id). Two more columns:

-- The address of that payment page. An order gets one page. If the buyer
-- leaves it and comes back, they are sent to the same page again, so there
-- are never two open pages that could both be paid for one order.
ALTER TABLE orders ADD COLUMN payment_url TEXT;

-- The provider's id of a refund. It is set in the one case where money
-- arrives for an order that can no longer be fulfilled (it expired while the
-- buyer was on the payment page): the payment is sent straight back.
ALTER TABLE orders ADD COLUMN refund_ref TEXT;
