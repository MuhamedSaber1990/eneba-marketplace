import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'
import type { HomeBlock } from '../api/home'

/** What every block component receives: its entry from the layout. */
export type BlockProps = { block: HomeBlock }

/**
 * The frame around a homepage block: the heading on top and the optional
 * "Show all" button underneath, both taken from the layout the backend sent.
 *
 * The section spans the whole window so it can have its own background
 * color. The content inside sits in the same 1200px column as everything else.
 */
export function BlockSection({ block, children }: BlockProps & { children: ReactNode }) {
  return (
    <section className="home-block">
      <div className="container">
        {block.title && <h2 className="section-heading">{block.title}</h2>}
        {children}
        {block.see_all_url && (
          <Link to={block.see_all_url} className="button button--secondary home-block__show-all">
            Show all<span className="visually-hidden"> {block.title}</span>
          </Link>
        )}
      </div>
    </section>
  )
}
