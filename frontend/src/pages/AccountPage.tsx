import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useAuth } from '../state/auth'
import './AuthPage.css'

/** The signed-in user's page. It is wrapped in RequireAuth, so `user` is set. */
export function AccountPage() {
  const { user, signOut } = useAuth()
  const navigate = useNavigate()
  const [signingOut, setSigningOut] = useState(false)

  useEffect(() => {
    document.title = 'Your account | Keybay'
  }, [])

  async function handleSignOut() {
    setSigningOut(true)
    // Leave this page first. It is a protected page, so if the user were
    // cleared while still on it, the route guard would send them to the
    // sign-in page, which is an odd place to land after logging out.
    navigate('/', { replace: true })
    try {
      await signOut()
    } catch {
      // The request failed, so the session is still alive and the header
      // keeps showing the account. Nothing else to do.
    }
  }

  return (
    <div className="container account">
      <h1 className="section-heading">Your account</h1>
      <div className="panel">
        <span className="account__label">Signed in as</span>
        <span className="account__email">{user?.email}</span>
        <p className="account__note">Your orders and keys will be listed here once checkout is built.</p>
        <button className="button button--secondary" onClick={handleSignOut} disabled={signingOut}>
          {signingOut ? 'Logging out…' : 'Log out'}
        </button>
      </div>
    </div>
  )
}
