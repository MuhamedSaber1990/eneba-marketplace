import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getOrder, type Order, type OrderStatus, payOrder } from '../api/checkout'
import { ApiError } from '../api/client'
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
      if (updated.status === 'paid') {
        cart.clear() // bought: the cart has done its job
      }
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

      {shownStatus === 'pending' && (
        <section className="panel order__payment" aria-live="polite">
          <h2>Payment</h2>
          <p>
            Your keys are held for <strong className="order__clock">{timeLeft}</strong> more minutes.
          </p>
          <p className="checkout__note">
            This is a test payment. No card is asked for and no money is taken. Choose how the payment should turn out.
          </p>
          <div className="order__buttons">
            <button className="button" disabled={paying} onClick={() => pay('approve')}>
              {paying ? 'One moment…' : `Pay ${formatPrice(order.total_cents)} (test)`}
            </button>
            <button className="button button--secondary" disabled={paying} onClick={() => pay('decline')}>
              Simulate a declined card
            </button>
          </div>
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
          <p className="checkout__note">It was not paid within 15 minutes, so the keys were released. Your cart is still filled.</p>
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
