import { type FormEvent, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { useAuth } from '../state/auth'
import './AuthPage.css'

/** The signed-in user's page. It is wrapped in RequireAuth, so `user` is set. */
export function AccountPage() {
  const { user, signOut, updateProfile } = useAuth()
  const navigate = useNavigate()
  const [signingOut, setSigningOut] = useState(false)

  // The form starts with what the account holds now.
  const [name, setName] = useState(user?.name ?? '')
  const [marketing, setMarketing] = useState(user?.marketing_opt_in ?? false)
  const [saving, setSaving] = useState(false)
  // What happened to the last save: nothing yet, saved, or the server's message.
  const [result, setResult] = useState<{ ok: boolean; message: string } | null>(null)

  useEffect(() => {
    document.title = 'Your account | Keybay'
  }, [])

  const changed = name.trim() !== user?.name || marketing !== user?.marketing_opt_in

  async function handleSave(event: FormEvent) {
    event.preventDefault()
    if (name.trim() === '') {
      setResult({ ok: false, message: 'Enter your name.' })
      return
    }
    setSaving(true)
    setResult(null)
    try {
      await updateProfile(name, marketing)
      // The server tidies the name (trims it, single spaces). Show it the
      // same way, so the field matches what was stored.
      setName(name.trim().replace(/\s+/g, ' '))
      setResult({ ok: true, message: 'Saved.' })
    } catch (caught) {
      setResult({ ok: false, message: caught instanceof Error ? caught.message : 'That could not be saved. Try again.' })
    }
    setSaving(false)
  }

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

      <form className="panel account__form" onSubmit={handleSave} noValidate>
        <h2>Your details</h2>

        <label htmlFor="account-name">Name</label>
        <input
          id="account-name"
          type="text"
          autoComplete="name"
          maxLength={60}
          required
          value={name}
          onChange={(event) => {
            setName(event.target.value)
            setResult(null)
          }}
        />

        <span className="account__label">Email</span>
        <span className="account__email">{user?.email}</span>
        <p className="auth__hint">The email cannot be changed here, because this demo cannot verify a new one.</p>

        <label className="auth__check">
          <input
            type="checkbox"
            className="checkbox"
            checked={marketing}
            onChange={(event) => {
              setMarketing(event.target.checked)
              setResult(null)
            }}
          />
          <span>Send me news about deals</span>
        </label>
        <p className="auth__hint">
          This stores your choice. No emails are sent by this demo: see the <Link to="/privacy">Privacy Policy</Link>.
        </p>

        <button className="button" disabled={saving || !changed}>
          {saving ? 'Saving…' : 'Save changes'}
        </button>
        {/* role="status" makes a screen reader read the outcome without moving focus. */}
        <p className={result && !result.ok ? 'auth__error' : 'account__saved'} role="status">
          {result?.message}
        </p>
      </form>

      <div className="panel">
        <h2>Your orders</h2>
        <p className="account__note">The keys you bought are kept with your orders.</p>
        <Link to="/orders" className="button">
          Your orders
        </Link>
        <button className="button button--secondary" onClick={handleSignOut} disabled={signingOut}>
          {signingOut ? 'Logging out…' : 'Log out'}
        </button>
      </div>
    </div>
  )
}
