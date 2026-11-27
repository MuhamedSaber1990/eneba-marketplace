import { useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getProduct, listProducts, type Offer, type ProductDetail } from '../api/catalog'
import { AddToCartButton } from '../components/AddToCartButton'
import { ChevronIcon, StarIcon } from '../components/Icons'
import { ProductCard } from '../components/ProductCard'
import { formatPrice } from '../lib/money'
import { type SellerMark, sellerMarks } from '../lib/sellerMark'
import { useAsync } from '../lib/useAsync'
import '../blocks/blocks.css'
import './ProductPage.css'

/** One product: its details and every seller's offer, for comparing. */
export function ProductPage() {
  const { slug = '' } = useParams()
  const product = useAsync((signal) => getProduct(slug, signal), slug)

  const name = product.status === 'ready' ? product.data.name : null
  useEffect(() => {
    document.title = name ? `${name} | Keybay` : 'Keybay'
  }, [name])

  if (product.status === 'loading') {
    return (
      <div className="container product">
        <p>Loading…</p>
      </div>
    )
  }

  if (product.status === 'error') {
    const notFound = product.httpStatus === 404
    return (
      <div className="container product">
        <h1 className="section-heading">{notFound ? 'Product not found' : 'Something went wrong'}</h1>
        <p className="product__message">
          {notFound ? 'There is no product at this address.' : `The product could not be loaded: ${product.message}`}
        </p>
        <Link to="/store" className="button button--secondary">
          Back to the store
        </Link>
      </div>
    )
  }

  return <ProductView product={product.data} />
}

function ProductView({ product }: { product: ProductDetail }) {
  // The API sorts offers: in stock first, cheapest first. So the first offer
  // is the best price, if it has stock at all.
  const best = product.offers.find((offer) => offer.stock > 0)
  const others = product.offers.filter((offer) => offer.id !== best?.id)

  // The marks are worked out for all of the page's sellers together, so no
  // two of them get the same color. markOf looks one up by offer.
  const marks = sellerMarks(product.offers.map((offer) => offer.seller_name))
  const markOf = (offer: Offer) => marks[product.offers.indexOf(offer)]

  return (
    <div className="container product">
      <nav className="product__breadcrumb" aria-label="Breadcrumb">
        <Link to="/store">Store</Link>
        <ChevronIcon />
        <Link to={`/store?platform=${encodeURIComponent(product.platform)}`}>{product.platform}</Link>
        <ChevronIcon />
        <span aria-current="page">{product.name}</span>
      </nav>

      <div className="product__top">
        <img className="product__cover" src={product.cover_url} alt="" width={225} height={300} />

        <div className="product__info">
          <h1 className="product__title">{product.name}</h1>
          <dl className="product__facts">
            <div>
              <dt>Platform</dt>
              <dd>{product.platform}</dd>
            </div>
            <div>
              <dt>Region</dt>
              <dd className="product__region">{product.region}</dd>
            </div>
            <div>
              <dt>Genre</dt>
              <dd>{product.genre}</dd>
            </div>
            <div>
              <dt>Released</dt>
              <dd>{formatDate(product.release_date)}</dd>
            </div>
          </dl>
        </div>

        {/* The price box holds the page's one yellow button. The buttons in
            the offer rows below are outlined, so it is clear which is which. */}
        <div className="product__summary">
          {best ? (
            <>
              <p className="product__summary-label">Best price</p>
              <p className="product__summary-price">{formatPrice(best.price_cents)}</p>
              <p className="product__summary-seller">from {best.seller_name}</p>
              <AddToCartButton
                offerId={best.id}
                stock={best.stock}
                context={` from ${best.seller_name}, the best price`}
                className="product__summary-button"
              />
            </>
          ) : (
            <p className="product__summary-price">Sold out</p>
          )}
          <p className="product__summary-count">
            {product.offers.length} {product.offers.length === 1 ? 'offer' : 'offers'} to compare
          </p>
        </div>
      </div>

      {best && (
        <section className="product__section">
          <h2>Best offer</h2>
          <ul className="offers">
            <OfferRow offer={best} mark={markOf(best)} isBest />
          </ul>
        </section>
      )}

      {others.length > 0 && (
        <section className="product__section">
          <h2>
            {best ? (
              <>
                <span className="product__count">{others.length}</span> other {others.length === 1 ? 'offer' : 'offers'}
              </>
            ) : (
              'Offers'
            )}
          </h2>
          <ul className="offers">
            {others.map((offer) => (
              <OfferRow key={offer.id} offer={offer} mark={markOf(offer)} />
            ))}
          </ul>
        </section>
      )}

      <section className="product__section">
        <h2>Product description</h2>
        <div className="panel">
          <p>{product.description}</p>
        </div>
      </section>

      {/* The key rebuilds the row when the visitor opens another product. */}
      <MoreOnPlatform key={product.slug} platform={product.platform} currentSlug={product.slug} />
    </div>
  )
}

function OfferRow({ offer, mark, isBest = false }: { offer: Offer; mark: SellerMark; isBest?: boolean }) {
  const soldOut = offer.stock === 0

  return (
    <li className={soldOut ? 'offer offer--sold-out' : 'offer'}>
      {/* The mark's color depends on the seller, so it is set here and not
          in the stylesheet. */}
      <span className="offer__mark" style={{ backgroundColor: mark.color }} aria-hidden="true">
        {mark.letter}
      </span>
      <div className="offer__seller">
        <span className="offer__name">
          {offer.seller_name}
          {isBest && <span className="offer__badge">Best price</span>}
        </span>
        <span className="offer__rating">
          <StarIcon />
          <strong>{offer.seller_rating.toFixed(1)} / 5</strong> rating
          <span className="offer__sales"> · {offer.seller_sales_count.toLocaleString('en')} sales</span>
        </span>
      </div>
      <span className="offer__stock">{soldOut ? 'Sold out' : `${offer.stock} in stock`}</span>
      <strong className="offer__price">{formatPrice(offer.price_cents)}</strong>
      <div className="offer__action">
        {!soldOut && (
          <AddToCartButton
            offerId={offer.id}
            stock={offer.stock}
            context={` from ${offer.seller_name}`}
            variant="secondary"
          />
        )}
      </div>
    </li>
  )
}

/**
 * One row of other products on the same platform, so the page does not end
 * at the description. It uses the store's list request with the platform
 * filter. If the request fails or nothing else is on that platform, the row
 * is left out: it is an extra, not something to show an error about.
 */
function MoreOnPlatform({ platform, currentSlug }: { platform: string; currentSlug: string }) {
  const list = useAsync((signal) => listProducts(`platform=${encodeURIComponent(platform)}`, signal), platform)
  if (list.status !== 'ready') return null

  const others = list.data.items.filter((item) => item.slug !== currentSlug).slice(0, 6)
  if (others.length === 0) return null

  return (
    <section className="product__section">
      <h2>More on {platform}</h2>
      <div className="product-row">
        {others.map((item) => (
          <ProductCard key={item.slug} product={item} />
        ))}
      </div>
    </section>
  )
}

/** "2024-03-14" becomes "14 Mar 2024". */
function formatDate(isoDate: string): string {
  return new Date(`${isoDate}T00:00:00`).toLocaleDateString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  })
}
