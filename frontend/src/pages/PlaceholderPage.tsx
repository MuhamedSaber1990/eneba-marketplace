import { Link } from 'react-router-dom'
import './PlaceholderPage.css'

/**
 * Stands in for pages that are built in later phases, so header links lead
 * somewhere. Also used for unknown URLs.
 */
export function PlaceholderPage({ title, note }: { title: string; note: string }) {
  return (
    <div className="container placeholder">
      <h1 className="section-heading">{title}</h1>
      <p className="placeholder__note">{note}</p>
      <Link to="/" className="button button--secondary">
        Back to the homepage
      </Link>
    </div>
  )
}
