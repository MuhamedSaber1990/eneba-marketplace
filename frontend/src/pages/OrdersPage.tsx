import { useEffect } from 'react'
import { Link } from 'react-router-dom'
import { listOrders, type OrderStatus } from '../api/checkout'
import { formatDateTime } from '../lib/dates'
import { formatPrice } from '../lib/money'
import { useAsync } from '../lib/useAsync'
import './Checkout.css'

const STATUS_TEXT: Record<OrderStatus, string> = {
  pending: 'Waiting for payment',
  paid: 'Paid',
  failed: 'Declined',
  expired: 'Expired',
}

/** The signed-in user's order history, newest first. */
export function OrdersPage() {
  const orders = useAsync(listOrders, 'orders')

  useEffect(() => {
    document.title = 'Your orders | Keybay'
  }, [])

  return (
    <div className="container checkout">
      <h1 className="section-heading">Your orders</h1>

      {orders.status === 'loading' && <p>Loading your orders…</p>}
      {orders.status === 'error' && (
        <div className="panel">
          <p>Your orders could not be loaded: {orders.message}</p>
        </div>
      )}
      {orders.status === 'ready' && orders.data.length === 0 && (
        <div className="panel">
          <p>You have not ordered anything yet.</p>
          <Link to="/store" className="button">
            Browse the store
          </Link>
        </div>
      )}
      {orders.status === 'ready' && orders.data.length > 0 && (
        <ul className="order-list">
          {orders.data.map((order) => (
            <li key={order.id}>
              <Link to={`/orders/${order.id}`} className="order-list__row">
                <span className="order-list__what">
                  <strong>{order.first_item}</strong>
                  {order.line_count > 1 && <span> and {order.line_count - 1} more</span>}
                  <span className="order-list__keys">
                    {order.key_count} {order.key_count === 1 ? 'key' : 'keys'}
                  </span>
                  <span className="order-list__date">{formatDateTime(order.created_at)}</span>
                </span>
                <span className={`order__status order__status--${order.status}`}>{STATUS_TEXT[order.status]}</span>
                <strong className="order-list__total">{formatPrice(order.total_cents)}</strong>
              </Link>
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
