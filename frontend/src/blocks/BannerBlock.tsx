import type { CSSProperties } from 'react'
import { Link } from 'react-router-dom'
import type { Banner } from '../api/home'
import { type BlockProps, BlockSection } from './BlockSection'
import { useBlockItems } from './useBlockItems'

/** Up to three promotion panels: the first is large, the others sit beside it. */
export function BannerBlock({ block }: BlockProps) {
  const { loading, items: banners } = useBlockItems<Banner>(block)
  if (!loading && banners.length === 0) return null

  return (
    <BlockSection block={block}>
      {loading ? (
        <div className="banners banners--loading" aria-hidden="true" />
      ) : (
        <div className="banners">
          {banners.map((banner, index) => (
            <Link
              key={banner.title}
              to={banner.link}
              className={index === 0 ? 'banner banner--large' : 'banner'}
              // The two gradient colors come from the database. They are
              // handed to the stylesheet as CSS variables.
              style={{ '--banner-a': banner.color_a, '--banner-b': banner.color_b } as CSSProperties}
            >
              <span className="banner__title">{banner.title}</span>
              <span className="banner__subtitle">{banner.subtitle}</span>
              <span className="button banner__cta">{banner.cta_label}</span>
            </Link>
          ))}
        </div>
      )}
    </BlockSection>
  )
}
