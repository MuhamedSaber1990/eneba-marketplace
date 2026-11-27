/**
 * The best sellers block as other pages use it: under an empty search and
 * under the cart, so neither page is a dead end. It is the homepage's block
 * pointed at the same endpoint.
 */
export const BEST_SELLERS_BLOCK = {
  type: 'best_sellers',
  title: 'Best sellers',
  see_all_url: '/store',
  endpoint: '/api/home/best-sellers',
}
