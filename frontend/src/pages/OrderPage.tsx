import { useEffect, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import { getOrder, type Order, type OrderStatus, payOrder, startCheckout } from '../api/checkout'
import { ApiError } from '../api/client'
import { CheckoutSteps } from '../components/CheckoutSteps'
import { clockLeft } from '../lib/countdown'
import { formatDateTime } from '../lib/dates'
import { formatPrice } from '../lib/money'
import { useCart } from '../state/cart'
import './Checkout.css'

const STATUS_TEXT: Record<OrderStatus, string> = {
  pending: 'Waiting for payment',
  paid: 'Paid',
  failed: 'Payment declined',
  expired: 'Expired',
}

// A loaded order comes with the moment its reservation ends, worked out when
// the answer arrived: this device's time then, plus the seconds the server
// said were left. Only the length of time comes from the server, so a
// device whose clock is wrong still counts down correctly.
type Load =
  | { status: 'loading' }
  | { status: 'ready'; order: Order; deadline: Date }
  | { status: 'missing' }
  | { status: 'error'; message: string }

function loaded(order: Order): Load {
  return { status: 'ready', order, deadline: new Date(Date.now() + order.seconds_left * 1000) }
}

/**
 * One order. While it is pending it shows the payment step and how long the
 * keys are still held. Once paid it shows the keys. The order comes from the
 * server every time, so reloading the page, or opening it days later from
 * the order history, shows the same thing.
 */
export function OrderPage() {
  const { id = '' } = useParams()
  const cart = useCart()
  // Stripe sends the buyer back here with ?checkout=done after paying.
  const [searchParams] = useSearchParams()
  const returnedFromPayment = searchParams.get('checkout') === 'done'

  const [load, setLoad] = useState<Load>({ status: 'loading' })
  const [paying, setPaying] = useState(false)
  const [message, setMessage] = useState<string | null>(null)

  // Fetch the order, again whenever the id in the address changes.
  useEffect(() => {
    const controller = new AbortController()
    getOrder(id, controller.signal)
      .then((order) => setLoad(loaded(order)))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return
        if (error instanceof ApiError && error.status === 404) {
          setLoad({ status: 'missing' })
        } else {
          setLoad({ status: 'error', message: error instanceof Error ? error.message : 'Something went wrong' })
        }
      })
    return () => controller.abort()
  }, [id])

  const order = load.status === 'ready' ? load.order : null

  useEffect(() => {
    document.title = 'Your order | Keybay'
  }, [])

  // Back from the payment page, the order is usually still "pending" for a
  // moment: the payment is confirmed by a message from Stripe to our server
  // (the webhook), which travels separately from the buyer. So the page asks
  // again every two seconds until the order has left "pending", and gives up
  // after a minute.
  const [checksLeft, setChecksLeft] = useState(30)
  const confirming = returnedFromPayment && order?.status === 'pending' && checksLeft > 0
  useEffect(() => {
    if (!confirming) return
    const timer = setTimeout(() => {
      getOrder(id)
        .then((fresh) => setLoad(loaded(fresh)))
        .catch(() => {}) // a failed check is simply tried again
        .finally(() => setChecksLeft((n) => n - 1))
    }, 2000)
    return () => clearTimeout(timer)
  }, [confirming, id, checksLeft])

  // Paid: the cart has done its job, whichever way the payment was made.
  const isPaid = order?.status === 'paid'
  useEffect(() => {
    if (isPaid) cart.clear()
    // Only when the order becomes paid, not whenever the cart object changes.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [isPaid])

  // A clock that ticks once a second while the order is pending.
  const [now, setNow] = useState(() => new Date())
  const isPending = order?.status === 'pending'
  useEffect(() => {
    if (!isPending) return
    const timer = setInterval(() => setNow(new Date()), 1000)
    return () => clearInterval(timer)
  }, [isPending])

  const timeLeft = load.status === 'ready' && isPending ? clockLeft(load.deadline, now) : null
  // The countdown reached zero while the page was open: for the buyer the
  // order is expired from this moment, whatever the last answer said.
  const shownStatus: OrderStatus | null = order ? (isPending && timeLeft === null ? 'expired' : order.status) : null

  async function pay(outcome: 'approve' | 'decline') {
    if (!order) return
    setPaying(true)
    setMessage(null)
    try {
      const updated = await payOrder(order.id, outcome)
      setLoad(loaded(updated))
    } catch (caught) {
      // "Not payable" comes with the order's final state. Show that state.
      const body = caught instanceof ApiError ? caught.body : null
      if (body !== null && typeof body === 'object' && 'order' in body) {
        setLoad(loaded(body.order as Order))
      } else {
        setMessage(caught instanceof Error ? caught.message : 'The payment could not be sent. Try again.')
      }
    } finally {
      setPaying(false)
    }
  }

  /** Stripe mode: get the address of the payment page and go there. */
  async function goToPayment() {
    if (!order) return
    setPaying(true)
    setMessage(null)
    try {
      // Leaving this site for Stripe's page. The button stays disabled.
      window.location.assign(await startCheckout(order.id))
    } catch (caught) {
      setMessage(caught instanceof Error ? caught.message : 'The payment page could not be opened. Try again.')
      setPaying(false)
      // The order may have expired in the meantime: show its current state.
      getOrder(order.id)
        .then((fresh) => setLoad(loaded(fresh)))
        .catch(() => {})
    }
  }

  if (load.status === 'loading') {
    return (
      <div className="container checkout">
        <p>Loading your order…</p>
      </div>
    )
  }
  if (load.status === 'missing' || load.status === 'error' || !order || !shownStatus) {
    return (
      <div className="container checkout">
        <h1 className="section-heading">Order not found</h1>
        <p className="checkout__note">
          {load.status === 'error' ? `The order could not be loaded: ${load.message}` : 'There is no order of yours at this address.'}
        </p>
        <Link to="/orders" className="button button--secondary">
          Your orders
        </Link>
      </div>
    )
  }

  return (
    <div className="container checkout">
      {/* Paid is the last step. Anything else is still about paying. */}
      <CheckoutSteps current={shownStatus === 'paid' ? 3 : 2} />
      <h1 className="section-heading">Your order</h1>
      <p className="order__meta">
        <span className={`order__status order__status--${shownStatus}`}>{STATUS_TEXT[shownStatus]}</span>
        <span>Placed {formatDateTime(order.created_at)}</span>
        <span>Total {formatPrice(order.total_cents)}</span>
      </p>

      {message && (
        <p className="checkout__warning" role="alert">
          {message}
        </p>
      )}

      {shownStatus === 'pending' && confirming && (
        <section className="panel order__payment" aria-live="polite">
          <h2>Confirming your payment</h2>
          <p className="checkout__note">
            Stripe is telling us about your payment. This usually takes a few seconds, and the page updates by itself.
          </p>
        </section>
      )}

      {shownStatus === 'pending' && !confirming && (
        <section className="panel order__payment" aria-live="polite">
          <h2>Payment</h2>
          <p>
            Your keys are held for <strong className="order__clock">{timeLeft}</strong> more minutes.
          </p>

          {returnedFromPayment && (
            <p className="checkout__warning" role="alert">
              We have not heard from Stripe about a payment yet. If you paid, reload this page in a moment.
            </p>
          )}

          {order.payment_mode === 'stripe' ? (
            <>
              <p className="checkout__note">
                You pay on Stripe's page, in test mode: no real money is taken. Use the test card{' '}
                <strong>4242 4242 4242 4242</strong> with any future date and any three digits.
              </p>
              <div className="order__buttons">
                <button className="button" disabled={paying} onClick={goToPayment}>
                  {paying ? 'Opening Stripe…' : `Pay ${formatPrice(order.total_cents)} with card`}
                </button>
              </div>
            </>
          ) : (
            <>
              <p className="checkout__note">
                This is a test payment. No card is asked for and no money is taken. Choose how the payment should turn
                out.
              </p>
              <div className="order__buttons">
                <button className="button" disabled={paying} onClick={() => pay('approve')}>
                  {paying ? 'One moment…' : `Pay ${formatPrice(order.total_cents)} (test)`}
                </button>
                <button className="button button--secondary" disabled={paying} onClick={() => pay('decline')}>
                  Simulate a declined card
                </button>
              </div>
            </>
          )}
        </section>
      )}

      {shownStatus === 'failed' && (
        <section className="panel order__payment">
          <h2>The payment was declined</h2>
          <p className="checkout__note">Nothing was charged and the keys were released. Your cart is still filled.</p>
          <Link to="/cart" className="button">
            Back to the cart
          </Link>
        </section>
      )}

      {shownStatus === 'expired' && (
        <section className="panel order__payment">
          <h2>This order ran out of time</h2>
          {order.refunded ? (
            <p className="checkout__note">
              Your payment arrived after the order had expired, so it was refunded in full and no keys were delivered.
              Your cart is still filled.
            </p>
          ) : (
            <p className="checkout__note">It was not paid in time, so the keys were released. Your cart is still filled.</p>
          )}
          <Link to="/cart" className="button">
            Back to the cart
          </Link>
        </section>
      )}

      <ul className="order__items">
        {order.items.map((item) => (
          <li key={item.product_slug + item.seller_name} className="order-item">
            <img src={item.cover_url} alt="" width={60} height={80} />
            <div className="order-item__info">
              <Link to={`/product/${item.product_slug}`} className="cart-line__name">
                {item.product_name}
              </Link>
              <span className="cart-line__seller">
                Sold by {item.seller_name} · {item.quantity} × {formatPrice(item.unit_price_cents)}
              </span>

              {shownStatus === 'paid' && (
                <ul className="order-item__keys" aria-label={`Keys for ${item.product_name}`}>
                  {item.keys.map((key) => (
                    <KeyRow key={key} code={key} />
                  ))}
                </ul>
              )}
            </div>
          </li>
        ))}
      </ul>

      <p className="order__links">
        <Link to="/orders">All your orders</Link>
        <Link to="/store">Back to the store</Link>
      </p>
    </div>
  )
}

/** One key with a button that copies it. */
function KeyRow({ code }: { code: string }) {
  const [copied, setCopied] = useState(false)

  async function copy() {
    try {
      await navigator.clipboard.writeText(code)
      setCopied(true)
      setTimeout(() => setCopied(false), 2000)
    } catch {
      // Copying can be refused by the browser. The key is on screen and can
      // be selected by hand, so nothing more is needed.
    }
  }

  return (
    <li className="key-row">
      <code>{code}</code>
      <button className="button button--secondary" onClick={copy}>
        {copied ? 'Copied' : 'Copy'}
        <span className="visually-hidden"> key {code}</span>
      </button>
    </li>
  )
}
