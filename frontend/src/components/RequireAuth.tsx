import type { ReactNode } from 'react'
import { Navigate, useLocation } from 'react-router-dom'
import { useAuth } from '../state/auth'

/**
 * Wraps a page that only signed-in users may see. A visitor who is not
 * signed in is sent to the sign-in page, with the address they wanted in
 * `next` so they come back here afterwards.
 *
 * This is a convenience for the visitor, not the protection itself. The
 * backend refuses the data behind such a page without a valid session, so
 * skipping this component in the browser gains nothing.
 */
export function RequireAuth({ children }: { children: ReactNode }) {
  const { loading, user } = useAuth()
  const location = useLocation()

  // Wait for "who am I" before deciding, or a signed-in user reloading the
  // page would be bounced to sign in for a moment.
  if (loading) {
    return (
      <div className="container auth-wait">
        <p>Loading…</p>
      </div>
    )
  }

  if (!user) {
    const next = encodeURIComponent(location.pathname + location.search)
    return <Navigate to={`/signin?next=${next}`} replace />
  }

  return children
}
