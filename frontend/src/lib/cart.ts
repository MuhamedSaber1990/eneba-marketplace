// The cart is a list of offers and quantities kept in the browser. It holds
// no prices and no product names: those are looked up from the server every
// time the cart is shown, so they are never stale and cannot be tampered with
// to any effect. These functions are plain data in, data out, with no React
// and no storage, which is what makes them easy to test.

export type CartLine = {
  offerId: number
  quantity: number
}

/** The same limits the backend enforces (backend/checkout/cart.go). */
export const MAX_QUANTITY = 5
export const MAX_LINES = 10

/** What happened when something was added. */
export type AddOutcome = 'added' | 'at_limit' | 'cart_full'

/**
 * Reads a cart saved as text. Whatever is in storage could be damaged, from
 * an older version of the site, or edited by hand, so nothing about it is
 * trusted: anything that is not a sensible line is dropped.
 */
export function parseCart(raw: string | null): CartLine[] {
  if (!raw) return []

  let parsed: unknown
  try {
    parsed = JSON.parse(raw)
  } catch {
    return []
  }
  if (!Array.isArray(parsed)) return []

  const lines: CartLine[] = []
  for (const entry of parsed) {
    if (typeof entry !== 'object' || entry === null) continue
    const { offerId, quantity } = entry as Record<string, unknown>
    if (typeof offerId !== 'number' || !Number.isInteger(offerId) || offerId <= 0) continue
    if (typeof quantity !== 'number' || !Number.isInteger(quantity) || quantity < 1) continue
    if (lines.some((line) => line.offerId === offerId)) continue // a duplicate
    if (lines.length === MAX_LINES) break
    lines.push({ offerId, quantity: Math.min(quantity, MAX_QUANTITY) })
  }
  return lines
}

/** Adds one of an offer, or one more if it is already in the cart. */
export function addToCart(cart: CartLine[], offerId: number): { cart: CartLine[]; outcome: AddOutcome } {
  const existing = cart.find((line) => line.offerId === offerId)

  if (existing) {
    if (existing.quantity >= MAX_QUANTITY) return { cart, outcome: 'at_limit' }
    return { cart: setQuantity(cart, offerId, existing.quantity + 1), outcome: 'added' }
  }
  if (cart.length >= MAX_LINES) return { cart, outcome: 'cart_full' }
  return { cart: [...cart, { offerId, quantity: 1 }], outcome: 'added' }
}

/** Sets a line's quantity, kept between 1 and the limit. Zero or less removes the line. */
export function setQuantity(cart: CartLine[], offerId: number, quantity: number): CartLine[] {
  if (quantity < 1) return removeFromCart(cart, offerId)
  const clamped = Math.min(Math.floor(quantity), MAX_QUANTITY)
  return cart.map((line) => (line.offerId === offerId ? { ...line, quantity: clamped } : line))
}

export function removeFromCart(cart: CartLine[], offerId: number): CartLine[] {
  return cart.filter((line) => line.offerId !== offerId)
}

/** How many keys are in the cart, over all lines. This is the number on the cart icon. */
export function cartCount(cart: CartLine[]): number {
  return cart.reduce((sum, line) => sum + line.quantity, 0)
}

/** How many of one offer are in the cart. */
export function quantityInCart(cart: CartLine[], offerId: number): number {
  return cart.find((line) => line.offerId === offerId)?.quantity ?? 0
}
