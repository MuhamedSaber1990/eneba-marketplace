import './Cover.css'

type Props = {
  src: string
  /**
   * True when the picture is an illustration with no text in it. The title
   * and the platform are then written over it. False for the cover the API
   * draws itself, which already contains both.
   */
  art: boolean
  title: string
  platform: string
  /** Covers further down a page load when scrolled to. The first ones should not wait. */
  lazy?: boolean
}

/**
 * A product's cover at card size or larger. Small covers (search dropdown,
 * cart, orders) are plain <img> elements: at 40 pixels wide a title would not
 * be readable.
 *
 * The image is decoration: the product's name is always in text next to it,
 * so the alt text is empty and a screen reader does not read the name twice.
 */
export function Cover({ src, art, title, platform, lazy = true }: Props) {
  return (
    <div className="cover">
      {/* width and height reserve the space, so the page does not jump while covers load */}
      <img src={src} alt="" width={450} height={600} loading={lazy ? 'lazy' : undefined} />
      {art && (
        <div className="cover__overlay" aria-hidden="true">
          <span className="cover__title">{title}</span>
          {/* data-platform lets the stylesheet give each platform a thin
              colored line, so editions of one game can be told apart. */}
          <span className="cover__platform" data-platform={platform}>
            {platform}
          </span>
        </div>
      )}
    </div>
  )
}
