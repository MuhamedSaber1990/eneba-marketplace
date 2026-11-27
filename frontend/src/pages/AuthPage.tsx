import { type FormEvent, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { PasswordField } from '../components/PasswordField'
import { safeNextPath } from '../lib/safeNext'
import { useAuth } from '../state/auth'
import './AuthPage.css'

/** The sign-in page: an email, a password, nothing else. Sign-up is SignUpPage. */
export function AuthPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  // Where to go afterwards. Only a path on this site is accepted.
  const next = safeNextPath(searchParams.get('next'))
  // The link to the sign-up form keeps the destination.
  const signUpLink = next === '/' ? '/signup' : `/signup?next=${encodeURIComponent(next)}`

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    document.title = 'Log in | Keybay'
  }, [])

  // Someone already signed in has no use for this page.
  if (!auth.loading && auth.user && !submitting) {
    return <Navigate to={next} replace />
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      await auth.signIn(email, password)
      navigate(next, { replace: true })
    } catch (caught) {
      // The API's messages are written for the visitor ("Wrong email or
      // password.", "Too many attempts...").
      setError(caught instanceof Error ? caught.message : 'Something went wrong. Try again.')
      setSubmitting(false)
    }
  }

  return (
    <div className="container auth">
      <form className="auth__form panel" onSubmit={handleSubmit} noValidate>
        <h1 className="auth__title">Log in</h1>

        {/* role="alert" makes a screen reader read the error as soon as it appears. */}
        {error && (
          <p className="auth__error" role="alert">
            {error}
          </p>
        )}

        <label htmlFor="auth-email">Email</label>
        <input
          id="auth-email"
          type="email"
          autoComplete="email"
          required
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />

        <label htmlFor="auth-password">Password</label>
        <PasswordField id="auth-password" value={password} onChange={setPassword} autoComplete="current-password" />

        <button className="button auth__submit" disabled={submitting}>
          {submitting ? 'One moment…' : 'Log in'}
        </button>

        <p className="auth__other">
          New here? <Link to={signUpLink}>Create an account</Link>
        </p>
      </form>
    </div>
  )
}
