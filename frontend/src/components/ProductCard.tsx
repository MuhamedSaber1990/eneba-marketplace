import { Link } from 'react-router-dom'
import type { ProductSummary } from '../api/catalog'
import { formatPrice } from '../lib/money'
import { AddToCartButton } from './AddToCartButton'
import { Cover } from './Cover'
import './ProductCard.css'

type Props = {
  product: ProductSummary
  /** Adds a numbered badge, used by the best sellers row. */
  rank?: number
  /** Adds "Add to cart" and "View offers" under the price, used by the store. */
  withActions?: boolean
}

/**
 * One product in a grid: cover on top, name, region and price underneath.
 *
 * The cover and the text are one link to the product page. The buttons are
 * outside that link, because a button inside a link is not valid HTML and a
 * click on it would also follow the link.
 */
export function ProductCard({ product, rank, withActions = false }: Props) {
  const soldOut = product.from_price_cents === null
  const productUrl = `/product/${product.slug}`

  return (
    <article className="product-card">
      <Link to={productUrl} className="product-card__link">
        <div className="product-card__media">
          <Cover src={product.cover_url} art={product.cover_art} title={product.title} platform={product.platform} />
          {rank !== undefined && (
            <span className="product-card__badge">
              <span className="visually-hidden">Rank </span>
              {rank}
            </span>
          )}
        </div>
        <div className="product-card__body">
          <h3 className="product-card__name">{product.name}</h3>
          <p className="product-card__region">{product.region}</p>
          {soldOut ? (
            <p className="product-card__sold-out">Sold out</p>
          ) : (
            <p className="product-card__price">
              <span className="product-card__from">From</span>
              {formatPrice(product.from_price_cents!)}
            </p>
          )}
        </div>
      </Link>

      {withActions && (
        <div className="product-card__actions">
          {/* "Add to cart" adds the cheapest offer, the one behind the "From"
              price. A sold-out product has no such offer and no button. */}
          {product.from_offer_id !== null && (
            <AddToCartButton offerId={product.from_offer_id} context={`: ${product.name}`} />
          )}
          <Link to={productUrl} className="button button--secondary">
            View offers<span className="visually-hidden"> for {product.name}</span>
          </Link>
        </div>
      )}
    </article>
  )
}

/** A placeholder with the card's shape (cover, two lines, price), shown while products load. */
export function ProductCardSkeleton() {
  return (
    <div className="product-card product-card--skeleton" aria-hidden="true">
      <div className="skeleton skeleton--cover" />
      <div className="product-card__body">
        <div className="skeleton skeleton--line" />
        <div className="skeleton skeleton--line skeleton--short" />
        <div className="skeleton skeleton--price" />
      </div>
    </div>
  )
}
