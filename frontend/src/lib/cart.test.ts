import { describe, expect, it } from 'vitest'
import { addToCart, cartCount, MAX_LINES, MAX_QUANTITY, parseCart, quantityInCart, removeFromCart, setQuantity } from './cart'

describe('addToCart', () => {
  it('adds a new line with quantity 1', () => {
    expect(addToCart([], 7)).toEqual({ cart: [{ offerId: 7, quantity: 1 }], outcome: 'added' })
  })

  it('adds one more when the offer is already in the cart', () => {
    const { cart, outcome } = addToCart([{ offerId: 7, quantity: 2 }], 7)
    expect(cart).toEqual([{ offerId: 7, quantity: 3 }])
    expect(outcome).toBe('added')
  })

  it('stops at the limit per offer', () => {
    const full = [{ offerId: 7, quantity: MAX_QUANTITY }]
    expect(addToCart(full, 7)).toEqual({ cart: full, outcome: 'at_limit' })
  })

  it('refuses a new line when the cart has the maximum number of lines', () => {
    const full = Array.from({ length: MAX_LINES }, (_, index) => ({ offerId: index + 1, quantity: 1 }))
    expect(addToCart(full, 99)).toEqual({ cart: full, outcome: 'cart_full' })
    // An offer that is already in the full cart can still get one more.
    expect(addToCart(full, 3).outcome).toBe('added')
  })

  it('does not change the cart it was given', () => {
    const original = [{ offerId: 7, quantity: 1 }]
    addToCart(original, 7)
    addToCart(original, 8)
    expect(original).toEqual([{ offerId: 7, quantity: 1 }])
  })
})

describe('setQuantity', () => {
  const cart = [
    { offerId: 1, quantity: 2 },
    { offerId: 2, quantity: 1 },
  ]

  it('changes only the named line', () => {
    expect(setQuantity(cart, 1, 4)).toEqual([
      { offerId: 1, quantity: 4 },
      { offerId: 2, quantity: 1 },
    ])
  })

  it('caps the quantity at the limit', () => {
    expect(setQuantity(cart, 1, 99)[0].quantity).toBe(MAX_QUANTITY)
  })

  it('removes the line at zero or below', () => {
    expect(setQuantity(cart, 1, 0)).toEqual([{ offerId: 2, quantity: 1 }])
    expect(setQuantity(cart, 1, -3)).toEqual([{ offerId: 2, quantity: 1 }])
  })

  it('ignores an offer that is not in the cart', () => {
    expect(setQuantity(cart, 42, 3)).toEqual(cart)
  })
})

describe('removeFromCart, cartCount, quantityInCart', () => {
  const cart = [
    { offerId: 1, quantity: 2 },
    { offerId: 2, quantity: 3 },
  ]

  it('removes a line', () => {
    expect(removeFromCart(cart, 1)).toEqual([{ offerId: 2, quantity: 3 }])
  })

  it('counts keys over all lines', () => {
    expect(cartCount(cart)).toBe(5)
    expect(cartCount([])).toBe(0)
  })

  it('reports the quantity of one offer', () => {
    expect(quantityInCart(cart, 2)).toBe(3)
    expect(quantityInCart(cart, 9)).toBe(0)
  })
})

describe('parseCart', () => {
  it('reads a cart it saved itself', () => {
    const cart = [{ offerId: 4, quantity: 2 }]
    expect(parseCart(JSON.stringify(cart))).toEqual(cart)
  })

  it.each([
    ['nothing saved', null],
    ['an empty string', ''],
    ['text that is not JSON', '{oops'],
    ['JSON that is not a list', '{"offerId":4,"quantity":1}'],
    ['a list of the wrong things', '[1, "two", null, [3]]'],
  ])('gives an empty cart for %s', (_name, raw) => {
    expect(parseCart(raw)).toEqual([])
  })

  it('drops lines that make no sense and keeps the good ones', () => {
    const raw = JSON.stringify([
      { offerId: 4, quantity: 2 },
      { offerId: -1, quantity: 1 },
      { offerId: 5, quantity: 0 },
      { offerId: 6.5, quantity: 1 },
      { offerId: '7', quantity: 1 },
      { offerId: 8 },
      { offerId: 4, quantity: 1 }, // the same offer again
      { offerId: 9, quantity: 3 },
    ])
    expect(parseCart(raw)).toEqual([
      { offerId: 4, quantity: 2 },
      { offerId: 9, quantity: 3 },
    ])
  })

  it('brings an oversized quantity down to the limit', () => {
    expect(parseCart('[{"offerId":4,"quantity":500}]')).toEqual([{ offerId: 4, quantity: MAX_QUANTITY }])
  })

  it('keeps at most the maximum number of lines', () => {
    const many = Array.from({ length: 30 }, (_, index) => ({ offerId: index + 1, quantity: 1 }))
    expect(parseCart(JSON.stringify(many))).toHaveLength(MAX_LINES)
  })

  it('ignores extra fields, such as a price someone added by hand', () => {
    expect(parseCart('[{"offerId":4,"quantity":1,"price_cents":1}]')).toEqual([{ offerId: 4, quantity: 1 }])
  })
})
