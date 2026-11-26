import { Link } from 'react-router-dom'
import type { PlatformTile } from '../api/home'
import { type BlockProps, BlockSection } from './BlockSection'
import { useBlockItems } from './useBlockItems'

/** One tile per platform, linking to the store filtered by that platform. */
export function PlatformsBlock({ block }: BlockProps) {
  const { loading, items: platforms } = useBlockItems<PlatformTile>(block)
  if (!loading && platforms.length === 0) return null

  return (
    <BlockSection block={block}>
      <div className="platform-tiles">
        {loading
          ? Array.from({ length: 4 }, (_, index) => <div key={index} className="platform-tile platform-tile--loading" aria-hidden="true" />)
          : platforms.map((platform) => (
              <Link key={platform.name} to={`/store?platform=${encodeURIComponent(platform.name)}`} className="platform-tile">
                <span className="platform-tile__name">{platform.name}</span>
                <span className="platform-tile__count">
                  {platform.product_count} {platform.product_count === 1 ? 'product' : 'products'}
                </span>
              </Link>
            ))}
      </div>
    </BlockSection>
  )
}
