import { type ReactNode, useEffect, useState } from 'react'
import { addToCart, type CartLine, cartCount, parseCart, removeFromCart, setQuantity } from '../lib/cart'
import { CartContext, type CartState } from './cart'

const STORAGE_KEY = 'keybay-cart'

/**
 * Keeps the cart in React state and mirrors it to localStorage, so it
 * survives a reload and a sign-in. The cart exists before anyone has an
 * account, which is why it lives in the browser and not on the server.
 */
export function CartProvider({ children }: { children: ReactNode }) {
  // The function form runs once, on the first render, to read the saved cart.
  const [lines, setLines] = useState<CartLine[]>(() => parseCart(localStorage.getItem(STORAGE_KEY)))

  // Save after every change.
  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(lines))
  }, [lines])

  // Another tab of the site changed the cart: the browser fires a "storage"
  // event in every other tab, and this one follows.
  useEffect(() => {
    function handleStorage(event: StorageEvent) {
      if (event.key === STORAGE_KEY) setLines(parseCart(event.newValue))
    }
    window.addEventListener('storage', handleStorage)
    return () => window.removeEventListener('storage', handleStorage)
  }, [])

  const value: CartState = {
    lines,
    count: cartCount(lines),
    add(offerId) {
      const result = addToCart(lines, offerId)
      setLines(result.cart)
      return result.outcome
    },
    setQuantity(offerId, quantity) {
      setLines(setQuantity(lines, offerId, quantity))
    },
    remove(offerId) {
      setLines(removeFromCart(lines, offerId))
    },
    clear() {
      setLines([])
    },
  }

  return <CartContext.Provider value={value}>{children}</CartContext.Provider>
}
