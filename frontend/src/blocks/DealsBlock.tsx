import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import type { Deal } from '../api/home'
import { Cover } from '../components/Cover'
import { ProductCardSkeleton } from '../components/ProductCard'
import { timeLeft } from '../lib/countdown'
import { formatPrice } from '../lib/money'
import { type BlockProps, BlockSection } from './BlockSection'
import { useBlockItems } from './useBlockItems'

/** Discounted offers, each with its old price and the time left. */
export function DealsBlock({ block }: BlockProps) {
  const { loading, items: deals } = useBlockItems<Deal>(block)

  // The countdown text depends on the current time. Updating `now` once a
  // minute redraws the cards with fresh "Ends in ..." text.
  const [now, setNow] = useState(() => new Date())
  useEffect(() => {
    const timer = setInterval(() => setNow(new Date()), 60_000)
    return () => clearInterval(timer)
  }, [])

  // A deal can run out while the page is open. Those are dropped here.
  const active = deals.filter((deal) => timeLeft(new Date(deal.ends_at), now) !== null)
  if (!loading && active.length === 0) return null

  return (
    <BlockSection block={block}>
      <div className="product-row">
        {loading
          ? Array.from({ length: 6 }, (_, index) => <ProductCardSkeleton key={index} />)
          : active.map((deal) => <DealCard key={deal.slug} deal={deal} now={now} />)}
      </div>
    </BlockSection>
  )
}

function DealCard({ deal, now }: { deal: Deal; now: Date }) {
  return (
    <Link to={`/product/${deal.slug}`} className="product-card">
      <div className="product-card__media">
        <Cover src={deal.cover_url} art={deal.cover_art} title={deal.title} platform={deal.platform} />
      </div>
      <div className="product-card__body">
        <h3 className="product-card__name">{deal.name}</h3>
        <p className="product-card__region">{deal.region}</p>
        <p className="product-card__price">
          {/* "From", the old price struck through, then the discount. */}
          <span className="product-card__from">
            From{' '}
            <s>
              <span className="visually-hidden">was </span>
              {formatPrice(deal.original_price_cents)}
            </s>{' '}
            <span className="product-card__discount">-{deal.discount_percent}%</span>
          </span>
          {formatPrice(deal.price_cents)}
        </p>
        <p className="deal__ends">Ends in {timeLeft(new Date(deal.ends_at), now)}</p>
      </div>
    </Link>
  )
}
