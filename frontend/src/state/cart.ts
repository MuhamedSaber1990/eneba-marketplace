import { createContext, useContext } from 'react'
import type { AddOutcome, CartLine } from '../lib/cart'

/** The cart and the actions on it, shared with every component. */
export type CartState = {
  lines: CartLine[]
  /** Keys in the cart over all lines: the number on the cart icon. */
  count: number
  add: (offerId: number) => AddOutcome
  setQuantity: (offerId: number, quantity: number) => void
  remove: (offerId: number) => void
  clear: () => void
}

export const CartContext = createContext<CartState | null>(null)

export function useCart(): CartState {
  const state = useContext(CartContext)
  if (state === null) {
    throw new Error('useCart must be used inside <CartProvider>')
  }
  return state
}
