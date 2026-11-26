const euros = new Intl.NumberFormat('en-IE', { style: 'currency', currency: 'EUR' })

/** Formats a price stored in cents: 4899 becomes "€48.99". */
export function formatPrice(cents: number): string {
  return euros.format(cents / 100)
}
