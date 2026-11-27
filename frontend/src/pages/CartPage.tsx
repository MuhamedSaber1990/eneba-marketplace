import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { priceCart } from '../api/checkout'
import { CartLines } from '../components/CartLines'
import { CheckoutSteps } from '../components/CheckoutSteps'
import { BinIcon, MinusIcon, PlusIcon } from '../components/Icons'
import { MAX_QUANTITY } from '../lib/cart'
import { formatPrice } from '../lib/money'
import { useAsync } from '../lib/useAsync'
import { useCart } from '../state/cart'
import './Checkout.css'

/**
 * The cart. The browser only remembers which offers and how many. Names,
 * prices and stock are asked from the server each time the cart changes, so
 * what is shown is always current.
 */
export function CartPage() {
  const cart = useCart()
  // The cart as text is the key: a change to any line prices it again.
  const cartKey = JSON.stringify(cart.lines)
  const priced = useAsync(() => (cart.lines.length > 0 ? priceCart(cart.lines) : Promise.resolve(null)), cartKey)

  useEffect(() => {
    document.title = 'Cart | Keybay'
  }, [])

  // Offers that no longer exist are taken out of the saved cart.
  const missing = priced.status === 'ready' && priced.data ? priced.data.missing_offer_ids : []
  const missingKey = missing.join(',')
  useEffect(() => {
    missing.forEach((offerId) => cart.remove(offerId))
    // Runs when the list of missing offers changes, not on every render.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [missingKey])

  // Keys over all lines, as the server priced them: the "3 keys" line.
  const keyCount = priced.status === 'ready' && priced.data ? priced.data.lines.reduce((sum, line) => sum + line.quantity, 0) : 0

  if (cart.lines.length === 0) {
    return (
      <div className="container checkout">
        <CheckoutSteps current={1} />
        <div className="panel">
          <h1 className="checkout__heading">My cart</h1>
          <p>Your cart is empty.</p>
          <p className="checkout__note">Pick an offer on any product page to add it here.</p>
          <Link to="/store" className="button">
            Browse the store
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="container checkout">
      <CheckoutSteps current={1} />

      {priced.status === 'loading' && <p>Loading your cart…</p>}
      {priced.status === 'error' && (
        <div className="panel">
          <p>The cart could not be loaded: {priced.message}</p>
        </div>
      )}

      {priced.status === 'ready' && priced.data && (
        <div className="checkout__layout">
          <section className="cart-panel">
            <h1 className="checkout__heading">My cart</h1>
            <CartLines
              lines={priced.data.lines}
              renderControls={(line) => {
                // One more is possible below the per-order limit and while
                // the seller has keys left.
                const canAdd = line.quantity < Math.min(MAX_QUANTITY, line.stock)
                return (
                  <>
                    <div className="stepper" role="group" aria-label={`Quantity of ${line.product_name}`}>
                      <button
                        aria-label="One less"
                        disabled={line.quantity <= 1}
                        onClick={() => cart.setQuantity(line.offer_id, line.quantity - 1)}
                      >
                        <MinusIcon size={16} />
                      </button>
                      {/* aria-live makes a screen reader say the new number. */}
                      <span className="stepper__value" aria-live="polite">
                        {line.quantity}
                      </span>
                      <button
                        aria-label="One more"
                        disabled={!canAdd}
                        onClick={() => cart.setQuantity(line.offer_id, line.quantity + 1)}
                      >
                        <PlusIcon size={16} />
                      </button>
                    </div>
                    <button
                      className="cart-line__remove"
                      aria-label={`Remove ${line.product_name}`}
                      onClick={() => cart.remove(line.offer_id)}
                    >
                      <BinIcon />
                    </button>
                  </>
                )
              }}
            />
          </section>

          <aside className="panel checkout__summary">
            <h2>Summary</h2>
            <p className="checkout__subtotal">
              <span>
                {keyCount} {keyCount === 1 ? 'key' : 'keys'}
              </span>
              <span>{formatPrice(priced.data.total_cents)}</span>
            </p>
            <p className="checkout__total">
              <span>Total</span>
              <strong>{formatPrice(priced.data.total_cents)}</strong>
            </p>
            {priced.data.can_checkout ? (
              <Link to="/checkout" className="button checkout__action">
                Go to checkout
              </Link>
            ) : (
              <p className="checkout__warning" role="alert">
                Lower the quantity or remove the items marked above to continue.
              </p>
            )}
            <p className="checkout__note">Nothing is reserved yet. Keys are held for you when you place the order.</p>
          </aside>
        </div>
      )}
    </div>
  )
}
