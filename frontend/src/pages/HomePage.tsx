import { type ComponentType, useEffect } from 'react'
import { Link } from 'react-router-dom'
import { getHomeLayout } from '../api/home'
import { BannerBlock } from '../blocks/BannerBlock'
import type { BlockProps } from '../blocks/BlockSection'
import { DealsBlock } from '../blocks/DealsBlock'
import { PlatformsBlock } from '../blocks/PlatformsBlock'
import { BestSellersBlock, ProductRowBlock } from '../blocks/ProductRowBlock'
import { useAsync } from '../lib/useAsync'
import '../blocks/blocks.css'

/**
 * Which component draws which block type. This map is the only thing the
 * frontend decides about the homepage. Which blocks appear, their order,
 * their headings and their content all come from the backend.
 */
const BLOCK_COMPONENTS: Record<string, ComponentType<BlockProps>> = {
  banner: BannerBlock,
  deals: DealsBlock,
  best_sellers: BestSellersBlock,
  new_releases: ProductRowBlock,
  platforms: PlatformsBlock,
}

export function HomePage() {
  const layout = useAsync(getHomeLayout, 'home-layout')

  useEffect(() => {
    document.title = 'Keybay: digital game keys'
  }, [])

  return (
    // Not a container: each block is a band across the whole window and
    // brings its own 1200px column (see BlockSection).
    <div className="home">
      <h1 className="visually-hidden">Keybay: digital game keys from many sellers</h1>

      {layout.status === 'loading' && (
        <div className="container home__state">
          <div className="banners banners--loading" aria-hidden="true" />
        </div>
      )}

      {layout.status === 'error' && (
        <div className="container home__state">
          <div className="panel">
            <p>The homepage could not be loaded. The store is still there.</p>
            <p>
              <Link to="/store" className="home__link">
                Go to the store
              </Link>
            </p>
          </div>
        </div>
      )}

      {layout.status === 'ready' &&
        layout.data.map((block) => {
          const Block = BLOCK_COMPONENTS[block.type]
          // A block type this version of the frontend does not know is skipped.
          if (!Block) return null
          // Two blocks can share a type, so the key includes the endpoint.
          return <Block key={`${block.type}:${block.endpoint}:${block.title}`} block={block} />
        })}
    </div>
  )
}
