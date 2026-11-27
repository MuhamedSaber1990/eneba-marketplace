import { useEffect } from 'react'
import { Link, useSearchParams } from 'react-router-dom'
import { safeNextPath } from '../lib/safeNext'
import { useAuth } from '../state/auth'
import './AuthPage.css'

/**
 * Shown once, right after signing up. It says what the account is for and
 * then sends the visitor on to wherever they were going.
 *
 * It is not a verification step. Nothing was emailed and nothing needs to be
 * confirmed: the account works from the moment it was created. The page is
 * wrapped in RequireAuth, so `user` is set.
 */
export function WelcomePage() {
  const { user } = useAuth()
  const [searchParams] = useSearchParams()
  const next = safeNextPath(searchParams.get('next'))

  useEffect(() => {
    document.title = 'Welcome | Keybay'
  }, [])

  return (
    <div className="container auth">
      <div className="auth__form panel welcome">
        <h1 className="auth__title">Welcome, {user?.name}</h1>
        <p>Your account is ready. Here is what it does:</p>
        <ul className="welcome__list">
          <li>Your orders are kept in one place.</li>
          <li>The keys you buy stay on their order page, to copy whenever you need them.</li>
          <li>Checkout goes straight to payment, with no forms to fill in again.</li>
        </ul>
        <Link to={next} replace className="button auth__submit">
          {next === '/' ? 'Start browsing' : 'Continue'}
        </Link>
      </div>
    </div>
  )
}
