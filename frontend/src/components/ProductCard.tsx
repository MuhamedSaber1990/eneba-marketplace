import { Link } from 'react-router-dom'
import type { ProductSummary } from '../api/catalog'
import { formatPrice } from '../lib/money'
import './ProductCard.css'

/**
 * One product in a grid: cover on top, name, region and price underneath.
 * `rank` adds a numbered badge, used by the best sellers row.
 */
export function ProductCard({ product, rank }: { product: ProductSummary; rank?: number }) {
  const soldOut = product.from_price_cents === null

  return (
    <Link to={`/product/${product.slug}`} className="product-card">
      {/* width and height reserve the space, so the grid does not jump while covers load */}
      <div className="product-card__media">
        <img src={product.cover_url} alt="" width={225} height={300} loading="lazy" />
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
  )
}

/** A grey placeholder with the card's shape, shown while products load. */
export function ProductCardSkeleton() {
  return <div className="product-card product-card--skeleton" aria-hidden="true" />
}
