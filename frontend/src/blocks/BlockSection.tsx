import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { HomeBlock } from '../api/home'

/** What every block component receives: its entry from the layout. */
export type BlockProps = { block: HomeBlock }

/**
 * The frame around a homepage block: the heading and the optional "See all"
 * link, both taken from the layout the backend sent.
 */
export function BlockSection({ block, children }: BlockProps & { children: ReactNode }) {
  return (
    <section className="home-block">
      {(block.title || block.see_all_url) && (
        <header className="home-block__header">
          {block.title && <h2 className="section-heading">{block.title}</h2>}
          {block.see_all_url && (
            <Link to={block.see_all_url} className="home-block__see-all">
              See all
            </Link>
          )}
        </header>
      )}
      {children}
    </section>
  )
}
