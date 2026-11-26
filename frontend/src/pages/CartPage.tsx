import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { priceCart } from '../api/checkout'
import { CartLines } from '../components/CartLines'
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

  if (cart.lines.length === 0) {
    return (
      <div className="container checkout">
        <h1 className="section-heading">Cart</h1>
        <div className="panel">
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
      <h1 className="section-heading">Cart</h1>

      {priced.status === 'loading' && <p>Loading your cart…</p>}
      {priced.status === 'error' && (
        <div className="panel">
          <p>The cart could not be loaded: {priced.message}</p>
        </div>
      )}

      {priced.status === 'ready' && priced.data && (
        <div className="checkout__layout">
          <CartLines
            lines={priced.data.lines}
            renderControls={(line) => (
              <>
                <label className="visually-hidden" htmlFor={`quantity-${line.offer_id}`}>
                  Quantity of {line.product_name}
                </label>
                <select
                  id={`quantity-${line.offer_id}`}
                  value={line.quantity}
                  onChange={(event) => cart.setQuantity(line.offer_id, Number(event.target.value))}
                >
                  {Array.from({ length: MAX_QUANTITY }, (_, index) => index + 1).map((n) => (
                    <option key={n} value={n}>
                      {n}
                    </option>
                  ))}
                </select>
                <button className="cart-line__remove" onClick={() => cart.remove(line.offer_id)}>
                  Remove<span className="visually-hidden"> {line.product_name}</span>
                </button>
              </>
            )}
          />

          <aside className="panel checkout__summary">
            <h2>Summary</h2>
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
