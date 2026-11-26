import type { ProductSummary } from './catalog'
import { apiGet } from './client'

// These types mirror the JSON written by the Go package backend/home.

/** One entry of the homepage: which block to draw and where its content is. */
export type HomeBlock = {
  type: string
  title: string
  see_all_url: string
  endpoint: string
}

export type Banner = {
  title: string
  subtitle: string
  cta_label: string
  link: string
  color_a: string
  color_b: string
}

export type Deal = {
  slug: string
  name: string
  platform: string
  region: string
  cover_url: string
  seller_name: string
  price_cents: number
  original_price_cents: number
  discount_percent: number
  ends_at: string // a timestamp like 2026-12-24T18:00:00Z
}

export type BestSeller = ProductSummary & { units_sold: number }

export type PlatformTile = {
  name: string
  product_count: number
}

/** The list of blocks on the homepage, in page order. */
export async function getHomeLayout(signal?: AbortSignal): Promise<HomeBlock[]> {
  const body = await apiGet<{ blocks: HomeBlock[] }>('/api/home', signal)
  return body.blocks
}

/**
 * The content of one block. Every block endpoint answers {"items": [...]},
 * so one function serves them all; T is the type of one item.
 */
export async function getBlockItems<T>(endpoint: string, signal?: AbortSignal): Promise<T[]> {
  const body = await apiGet<{ items: T[] }>(endpoint, signal)
  return body.items
}
