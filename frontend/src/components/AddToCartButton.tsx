import { MAX_QUANTITY, quantityInCart } from '../lib/cart'
import { useCart } from '../state/cart'

type Props = {
  offerId: number
  /** Keys the seller has left, when the caller knows it. */
  stock?: number
  /** Read out after the label by screen readers, for example " from GG Depot". */
  context: string
  /** The yellow button is the main action of a page; the outlined one is not. */
  variant?: 'primary' | 'secondary'
  className?: string
}

/**
 * Adds one key of an offer to the cart. Once the offer is in the cart the
 * label shows how many, and the button stops at the per-order limit or at
 * the seller's stock, whichever is lower.
 */
export function AddToCartButton({ offerId, stock = MAX_QUANTITY, context, variant = 'primary', className = '' }: Props) {
  const cart = useCart()
  const inCart = quantityInCart(cart.lines, offerId)
  const atLimit = inCart >= Math.min(MAX_QUANTITY, stock)
  const classes = ['button', variant === 'secondary' ? 'button--secondary' : '', className].filter(Boolean).join(' ')

  return (
    <button className={classes} disabled={atLimit} onClick={() => cart.add(offerId)}>
      {inCart > 0 ? `In cart (${inCart})` : 'Add to cart'}
      <span className="visually-hidden">{context}</span>
    </button>
  )
}
