import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { getHealth, type Health } from '../api/health'
import './HomePage.css'

// What the status panel can be showing.
type HealthState = { phase: 'loading' } | { phase: 'loaded'; health: Health } | { phase: 'failed' }

/**
 * Phase 1 homepage: a placeholder that proves the frontend, the Go API and
 * the database are connected. Phase 3 replaces it with blocks loaded from
 * the backend.
 */
export function HomePage() {
  const [state, setState] = useState<HealthState>({ phase: 'loading' })

  useEffect(() => {
    // AbortController cancels the request if the page is left before it ends.
    const controller = new AbortController()
    getHealth(controller.signal)
      .then((health) => setState({ phase: 'loaded', health }))
      .catch((error: unknown) => {
        if (!controller.signal.aborted) {
          console.error(error)
          setState({ phase: 'failed' })
        }
      })
    return () => controller.abort()
  }, [])

  return (
    <div className="container home">
      <h1 className="section-heading">Digital game keys, from many sellers</h1>
      <p className="home__lead">
        The shop is being built. The catalog, search and checkout arrive in the next phases.
      </p>
      <Link to="/store" className="button">
        Browse the store
      </Link>

      <section className="panel home__status" aria-live="polite">
        <h2>System status</h2>
        <StatusRows state={state} />
      </section>
    </div>
  )
}

function StatusRows({ state }: { state: HealthState }) {
  if (state.phase === 'loading') {
    return <p>Checking the API…</p>
  }
  if (state.phase === 'failed') {
    return <p className="home__bad">The API did not answer.</p>
  }
  const { health } = state
  return (
    <dl>
      <div>
        <dt>API</dt>
        <dd className="home__good">reachable</dd>
      </div>
      <div>
        <dt>Database</dt>
        <dd className={health.database === 'ok' ? 'home__good' : 'home__bad'}>{health.database}</dd>
      </div>
    </dl>
  )
}
