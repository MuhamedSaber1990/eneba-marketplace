import { Link } from 'react-router-dom'

/**
 * Stands in for pages that are built in later phases, so header links lead
 * somewhere. Also used for unknown URLs.
 */
export function PlaceholderPage({ title, note }: { title: string; note: string }) {
  return (
    <div className="container" style={{ paddingTop: 'var(--space-4)' }}>
      <h1 className="section-heading">{title}</h1>
      <p style={{ marginBottom: 'var(--space-3)', color: 'var(--color-text-soft)' }}>{note}</p>
      <Link to="/" className="button button--secondary">
        Back to the homepage
      </Link>
    </div>
  )
}
