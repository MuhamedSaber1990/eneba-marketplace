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
              style={bannerStyle(banner)}
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

/**
 * The banner's two colors and its picture come from the database. They are
 * handed to the stylesheet as CSS variables. A banner without a picture
 * leaves --banner-image unset, and the stylesheet then shows the gradient.
 */
function bannerStyle(banner: Banner): CSSProperties {
  const style: Record<string, string> = { '--banner-a': banner.color_a, '--banner-b': banner.color_b }
  if (banner.image_url) {
    style['--banner-image'] = `url("${banner.image_url}")`
  }
  return style as CSSProperties
}
