import { useEffect } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getProduct, type Offer, type ProductDetail } from '../api/catalog'
import { formatPrice } from '../lib/money'
import { useAsync } from '../lib/useAsync'
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

  return (
    <div className="container product">
      <nav className="product__breadcrumb" aria-label="Breadcrumb">
        <Link to="/store">Store</Link>
        <span aria-hidden="true">/</span>
        <Link to={`/store?platform=${encodeURIComponent(product.platform)}`}>{product.platform}</Link>
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

        <div className="product__summary">
          {best ? (
            <>
              <p className="product__summary-label">Best price</p>
              <p className="product__summary-price">{formatPrice(best.price_cents)}</p>
              <p className="product__summary-seller">from {best.seller_name}</p>
            </>
          ) : (
            <p className="product__summary-price">Sold out</p>
          )}
          <p className="product__summary-count">
            {product.offers.length} {product.offers.length === 1 ? 'offer' : 'offers'} to compare
          </p>
        </div>
      </div>

      <section className="product__section">
        <h2>Offers</h2>
        <div className="offers-scroll">
          <table className="offers">
            <thead>
              <tr>
                <th scope="col">Seller</th>
                <th scope="col">Rating</th>
                <th scope="col" className="offers__optional">
                  Sales
                </th>
                <th scope="col">Stock</th>
                <th scope="col" className="offers__price">
                  Price
                </th>
              </tr>
            </thead>
            <tbody>
              {product.offers.map((offer) => (
                <OfferRow key={offer.id} offer={offer} isBest={offer.id === best?.id} />
              ))}
            </tbody>
          </table>
        </div>
      </section>

      <section className="product__section">
        <h2>Product description</h2>
        <div className="panel">
          <p>{product.description}</p>
        </div>
      </section>
    </div>
  )
}

function OfferRow({ offer, isBest }: { offer: Offer; isBest: boolean }) {
  const soldOut = offer.stock === 0
  return (
    <tr className={soldOut ? 'offers__row--sold-out' : undefined}>
      <th scope="row">
        {offer.seller_name}
        {isBest && <span className="offers__badge">Best price</span>}
      </th>
      <td>{offer.seller_rating.toFixed(1)} / 5</td>
      <td className="offers__optional">{offer.seller_sales_count.toLocaleString('en')}</td>
      <td>{soldOut ? 'Sold out' : `${offer.stock} in stock`}</td>
      <td className="offers__price">{formatPrice(offer.price_cents)}</td>
    </tr>
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
