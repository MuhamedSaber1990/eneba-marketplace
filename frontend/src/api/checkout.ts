import type { CartLine } from '../lib/cart'
import { apiGet, apiPost } from './client'

// These types mirror the JSON written by the Go package backend/checkout.

export type PricedLine = {
  offer_id: number
  quantity: number
  product_slug: string
  product_name: string
  region: string
  seller_name: string
  cover_url: string
  unit_price_cents: number
  line_total_cents: number
  stock: number
  available: boolean // stock covers the quantity
}

export type PricedCart = {
  lines: PricedLine[]
  missing_offer_ids: number[]
  total_cents: number
  can_checkout: boolean
}

export type OrderStatus = 'pending' | 'paid' | 'failed' | 'expired'

export type OrderItem = {
  product_slug: string
  product_name: string
  seller_name: string
  cover_url: string
  unit_price_cents: number
  quantity: number
  keys: string[] // empty until the order is paid
}

export type Order = {
  id: string
  status: OrderStatus
  total_cents: number
  created_at: string
  expires_at: string
  /** Time left on a pending order's reservation, by the server's clock. */
  seconds_left: number
  paid_at: string | null
  /** A payment arrived too late to be fulfilled and was sent back. */
  refunded: boolean
  /** Which payment step the order page shows. */
  payment_mode: 'simulated' | 'stripe'
  items: OrderItem[]
}

export type OrderSummary = {
  id: string
  status: OrderStatus
  total_cents: number
  created_at: string
  key_count: number
  line_count: number
  first_item: string
}

/** The browser's cart lines in the shape the API expects. Only ids and quantities are sent. */
function toItems(lines: CartLine[]) {
  return lines.map((line) => ({ offer_id: line.offerId, quantity: line.quantity }))
}

/** Looks up current prices and stock for a cart. Reserves nothing. */
export function priceCart(lines: CartLine[]): Promise<PricedCart> {
  return apiPost<PricedCart>('/api/cart/price', { items: toItems(lines) })
}

/**
 * Places the order. expectedTotalCents is the total the buyer was looking
 * at: if a price changed in the meantime the API refuses, so nobody pays a
 * price they did not see.
 */
export async function createOrder(lines: CartLine[], expectedTotalCents: number): Promise<Order> {
  const body = await apiPost<{ order: Order }>('/api/orders', {
    items: toItems(lines),
    expected_total_cents: expectedTotalCents,
  })
  return body.order
}

/** The simulated payment: the buyer picks the outcome. */
export async function payOrder(orderId: string, outcome: 'approve' | 'decline'): Promise<Order> {
  const body = await apiPost<{ order: Order }>(`/api/orders/${encodeURIComponent(orderId)}/pay`, { outcome })
  return body.order
}

/**
 * Asks for the address of the payment provider's page for this order. The
 * browser is then sent there. Used when payment_mode is "stripe".
 */
export async function startCheckout(orderId: string): Promise<string> {
  const body = await apiPost<{ url: string }>(`/api/orders/${encodeURIComponent(orderId)}/checkout`, {})
  return body.url
}

export async function getOrder(orderId: string, signal?: AbortSignal): Promise<Order> {
  const body = await apiGet<{ order: Order }>(`/api/orders/${encodeURIComponent(orderId)}`, signal)
  return body.order
}

export async function listOrders(signal?: AbortSignal): Promise<OrderSummary[]> {
  const body = await apiGet<{ orders: OrderSummary[] }>('/api/orders', signal)
  return body.orders
}
