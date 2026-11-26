import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { PricedLine } from '../api/checkout'
import { formatPrice } from '../lib/money'

type Props = {
  lines: PricedLine[]
  /**
   * What to show next to each line's price. The cart passes a quantity
   * picker and a remove button; the checkout passes nothing and shows the
   * quantity as text.
   */
  renderControls?: (line: PricedLine) => ReactNode
}

/** The list of cart lines, shared by the cart page and the checkout page. */
export function CartLines({ lines, renderControls }: Props) {
  return (
    <ul className="cart-lines">
      {lines.map((line) => (
        <li key={line.offer_id} className="cart-line">
          <img src={line.cover_url} alt="" width={60} height={80} />
          <div className="cart-line__info">
            <Link to={`/product/${line.product_slug}`} className="cart-line__name">
              {line.product_name}
            </Link>
            <span className="cart-line__seller">Sold by {line.seller_name}</span>
            <span className="cart-line__unit">{formatPrice(line.unit_price_cents)} each</span>
            {!line.available && (
              <span className="cart-line__problem" role="alert">
                {line.stock === 0 ? 'Sold out.' : `Only ${line.stock} left.`}
              </span>
            )}
          </div>
          <div className="cart-line__controls">
            {renderControls ? renderControls(line) : <span className="cart-line__quantity">× {line.quantity}</span>}
          </div>
          <strong className="cart-line__total">{formatPrice(line.line_total_cents)}</strong>
        </li>
      ))}
    </ul>
  )
}
