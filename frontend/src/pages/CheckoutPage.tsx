import { useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { createOrder, priceCart } from '../api/checkout'
import { ApiError } from '../api/client'
import { CartLines } from '../components/CartLines'
import { CheckoutSteps } from '../components/CheckoutSteps'
import { formatPrice } from '../lib/money'
import { useAsync } from '../lib/useAsync'
import { useAuth } from '../state/auth'
import { useCart } from '../state/cart'
import './Checkout.css'

/**
 * The last look before buying. "Place order" creates the order on the
 * server, which reserves the keys, and moves on to the order page to pay.
 * This page is wrapped in RequireAuth: only a signed-in user gets here.
 */
export function CheckoutPage() {
  const cart = useCart()
  const { user } = useAuth()
  const navigate = useNavigate()

  // `attempt` is part of the key, so raising it prices the cart again. That
  // is how the page refreshes itself after the server reports a new price.
  const [attempt, setAttempt] = useState(0)
  const priced = useAsync(
    () => (cart.lines.length > 0 ? priceCart(cart.lines) : Promise.resolve(null)),
    `${JSON.stringify(cart.lines)}#${attempt}`,
  )

  const [placing, setPlacing] = useState(false)
  const [problem, setProblem] = useState<{ message: string; code: string | null } | null>(null)

  useEffect(() => {
    document.title = 'Checkout | Keybay'
  }, [])

  async function placeOrder(totalCents: number) {
    setPlacing(true)
    setProblem(null)
    try {
      const order = await createOrder(cart.lines, totalCents)
      // The cart is kept until the order is paid, so a declined payment or
      // an expired order does not cost the buyer their selection.
      navigate(`/orders/${order.id}`)
    } catch (caught) {
      const error = caught instanceof ApiError ? caught : null
      setProblem({
        message: error?.message ?? 'The order could not be placed. Try again.',
        code: error?.code ?? null,
      })
      setPlacing(false)
      // A price or the stock changed: show the cart as it is now.
      if (error?.code === 'price_changed' || error?.code === 'out_of_stock' || error?.code === 'offer_gone') {
        setAttempt((n) => n + 1)
      }
    }
  }

  if (cart.lines.length === 0) {
    return (
      <div className="container checkout">
        <CheckoutSteps current={2} />
        <div className="panel">
          <h1 className="checkout__heading">Checkout</h1>
          <p>There is nothing to check out.</p>
          <Link to="/store" className="button">
            Browse the store
          </Link>
        </div>
      </div>
    )
  }

  return (
    <div className="container checkout">
      <CheckoutSteps current={2} />

      {problem && (
        <div className="checkout__warning" role="alert">
          <p>{problem.message}</p>
          {problem.code === 'too_many_pending' && (
            <p>
              <Link to="/orders">See your orders</Link>
            </p>
          )}
          {problem.code === 'out_of_stock' && (
            <p>
              <Link to="/cart">Change your cart</Link>
            </p>
          )}
        </div>
      )}

      {priced.status === 'loading' && <p>Loading your order…</p>}
      {priced.status === 'error' && (
        <div className="panel">
          <p>The order could not be loaded: {priced.message}</p>
        </div>
      )}

      {priced.status === 'ready' && priced.data && (
        <div className="checkout__layout">
          <section className="cart-panel">
            <h1 className="checkout__heading">Checkout</h1>
            <CartLines lines={priced.data.lines} />
          </section>

          <aside className="panel checkout__summary">
            <h2>Your order</h2>
            <p className="checkout__note">Buying as {user?.email}</p>
            <p className="checkout__total">
              <span>Total</span>
              <strong>{formatPrice(priced.data.total_cents)}</strong>
            </p>
            {priced.data.can_checkout ? (
              <button
                className="button checkout__action"
                disabled={placing}
                onClick={() => placeOrder(priced.data!.total_cents)}
              >
                {placing ? 'Placing your order…' : 'Place order'}
              </button>
            ) : (
              <p className="checkout__warning" role="alert">
                Some items are no longer available. <Link to="/cart">Change your cart</Link>
              </p>
            )}
            <p className="checkout__note">
              Placing the order holds your keys while you pay. No real money is taken on this site.
            </p>
          </aside>
        </div>
      )}
    </div>
  )
}
