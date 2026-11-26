import { describe, expect, it } from 'vitest'
import { formatPrice } from './money'

describe('formatPrice', () => {
  it.each([
    [4899, '€48.99'],
    [1000, '€10.00'],
    [859, '€8.59'],
    [5, '€0.05'],
    [123456, '€1,234.56'],
  ])('%d cents is %s', (cents, expected) => {
    expect(formatPrice(cents)).toBe(expected)
  })
})
