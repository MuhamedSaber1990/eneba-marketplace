import { apiGet } from './client'

// These types mirror the JSON written by backend/catalog/catalog.go.

export type ProductSummary = {
  slug: string
  name: string
  platform: string
  region: string
  genre: string
  release_date: string // YYYY-MM-DD
  from_price_cents: number | null // null means sold out
  offer_count: number
  cover_url: string
}

export type ProductList = {
  items: ProductSummary[]
  total: number
  page: number
  page_size: number
  total_pages: number
}

export type Offer = {
  id: number
  seller_name: string
  seller_rating: number
  seller_sales_count: number
  price_cents: number
  stock: number
}

export type ProductDetail = ProductSummary & {
  description: string
  offers: Offer[]
}

export type Filters = {
  platforms: string[]
  regions: string[]
  min_price_cents: number
  max_price_cents: number
}

/**
 * Loads one page of the store. The query string is passed through as it is in
 * the browser's address bar, because the page URL and the API use the same
 * parameter names (platform, region, min_price, max_price, q, sort, page).
 */
export function listProducts(queryString: string, signal?: AbortSignal) {
  return apiGet<ProductList>(`/api/products?${queryString}`, signal)
}

export function getProduct(slug: string, signal?: AbortSignal) {
  return apiGet<ProductDetail>(`/api/products/${encodeURIComponent(slug)}`, signal)
}

export function getFilters(signal?: AbortSignal) {
  return apiGet<Filters>('/api/filters', signal)
}
