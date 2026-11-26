import { type FormEvent, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { safeNextPath } from '../lib/safeNext'
import { useAuth } from '../state/auth'
import './AuthPage.css'

const TEXT = {
  signin: {
    title: 'Log in',
    button: 'Log in',
    otherQuestion: 'New here?',
    otherLabel: 'Create an account',
    otherPath: '/signup',
  },
  signup: {
    title: 'Create an account',
    button: 'Create account',
    otherQuestion: 'Already have an account?',
    otherLabel: 'Log in',
    otherPath: '/signin',
  },
}

/** The sign-in page and the sign-up page. They differ only in wording and in which action the form calls. */
export function AuthPage({ mode }: { mode: 'signin' | 'signup' }) {
  const auth = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()
  const text = TEXT[mode]

  // Where to go afterwards. Only a path on this site is accepted.
  const next = safeNextPath(searchParams.get('next'))
  // The link to the other form keeps the destination.
  const otherLink = next === '/' ? text.otherPath : `${text.otherPath}?next=${encodeURIComponent(next)}`

  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    document.title = `${text.title} | Keybay`
  }, [text.title])

  // Someone already signed in has no use for this page.
  if (!auth.loading && auth.user && !submitting) {
    return <Navigate to={next} replace />
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    setError(null)
    setSubmitting(true)
    try {
      if (mode === 'signin') {
        await auth.signIn(email, password)
      } else {
        await auth.signUp(email, password)
      }
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
        <h1 className="auth__title">{text.title}</h1>

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
        <input
          id="auth-password"
          type="password"
          // Tells a password manager whether to offer a saved password or suggest a new one.
          autoComplete={mode === 'signin' ? 'current-password' : 'new-password'}
          required
          aria-describedby={mode === 'signup' ? 'auth-password-hint' : undefined}
          value={password}
          onChange={(event) => setPassword(event.target.value)}
        />
        {mode === 'signup' && (
          <p id="auth-password-hint" className="auth__hint">
            At least 8 characters. A longer phrase is better than symbols.
          </p>
        )}

        <button className="button auth__submit" disabled={submitting}>
          {submitting ? 'One moment…' : text.button}
        </button>

        <p className="auth__other">
          {text.otherQuestion} <Link to={otherLink}>{text.otherLabel}</Link>
        </p>
      </form>
    </div>
  )
}
