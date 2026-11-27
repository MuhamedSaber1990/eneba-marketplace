import { type FormEvent, useEffect, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import { PasswordField } from '../components/PasswordField'
import { passwordRules, passwordTooLong } from '../lib/passwordRules'
import { safeNextPath } from '../lib/safeNext'
import { useAuth } from '../state/auth'
import './AuthPage.css'

/**
 * The sign-up form. It checks what was typed before sending, to save a round
 * trip, and shows the password rules as they are met. None of that is the
 * protection: the server checks every field again, and a request sent
 * without this form gets the same answers.
 */
export function SignUpPage() {
  const auth = useAuth()
  const navigate = useNavigate()
  const [searchParams] = useSearchParams()

  // Where the visitor was going. Only a path on this site is accepted.
  const next = safeNextPath(searchParams.get('next'))
  const nextQuery = next === '/' ? '' : `?next=${encodeURIComponent(next)}`

  const [name, setName] = useState('')
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [confirmation, setConfirmation] = useState('')
  const [acceptTerms, setAcceptTerms] = useState(false)
  const [marketing, setMarketing] = useState(false) // off unless the visitor switches it on
  const [error, setError] = useState<string | null>(null)
  const [submitting, setSubmitting] = useState(false)

  useEffect(() => {
    document.title = 'Create an account | Keybay'
  }, [])

  // Someone already signed in has no use for this page.
  if (!auth.loading && auth.user && !submitting) {
    return <Navigate to={next} replace />
  }

  const rules = passwordRules(password, confirmation, email)
  const tooLong = passwordTooLong(password)

  /** The first thing wrong with the form, or null when it can be sent. */
  function firstProblem(): string | null {
    if (name.trim() === '') return 'Enter your name.'
    if (email.trim() === '') return 'Enter your email address.'
    if (tooLong) return 'That password is too long. Use at most 72 characters.'
    if (rules.some((rule) => !rule.met)) return 'The password does not meet the rules below yet.'
    if (!acceptTerms) return 'You need to agree to the Terms and the Privacy Policy to create an account.'
    return null
  }

  async function handleSubmit(event: FormEvent) {
    event.preventDefault()
    const problem = firstProblem()
    if (problem) {
      setError(problem)
      return
    }

    setError(null)
    setSubmitting(true)
    try {
      await auth.signUp({
        name,
        email,
        password,
        password_confirm: confirmation,
        accept_terms: acceptTerms,
        marketing_opt_in: marketing,
      })
      // A short welcome first, then on to wherever the visitor was going.
      navigate(`/welcome${nextQuery}`, { replace: true })
    } catch (caught) {
      setError(caught instanceof Error ? caught.message : 'Something went wrong. Try again.')
      setSubmitting(false)
    }
  }

  return (
    <div className="container auth">
      <form className="auth__form panel" onSubmit={handleSubmit} noValidate>
        <h1 className="auth__title">Create an account</h1>

        {/* role="alert" makes a screen reader read the error as soon as it appears. */}
        {error && (
          <p className="auth__error" role="alert">
            {error}
          </p>
        )}

        <label htmlFor="signup-name">Name</label>
        <input
          id="signup-name"
          type="text"
          autoComplete="name"
          maxLength={60}
          required
          value={name}
          onChange={(event) => setName(event.target.value)}
        />

        <label htmlFor="signup-email">Email</label>
        <input
          id="signup-email"
          type="email"
          autoComplete="email"
          required
          aria-describedby="signup-email-hint"
          value={email}
          onChange={(event) => setEmail(event.target.value)}
        />
        <p id="signup-email-hint" className="auth__hint">
          Nothing is sent to it. This demo has no email service.
        </p>

        <label htmlFor="signup-password">Password</label>
        <PasswordField
          id="signup-password"
          value={password}
          onChange={setPassword}
          autoComplete="new-password"
          describedBy="signup-rules"
        />

        <label htmlFor="signup-confirmation">Confirm password</label>
        <PasswordField
          id="signup-confirmation"
          value={confirmation}
          onChange={setConfirmation}
          autoComplete="new-password"
          describedBy="signup-rules"
          what="confirmation"
        />

        {/* Each rule gets a tick when it is met. The words "done" and "to do"
            are there for screen readers, so the state is not shown by color
            and a symbol alone. */}
        <ul id="signup-rules" className="rules" aria-label="Password rules">
          {rules.map((rule) => (
            <li key={rule.id} className={rule.met ? 'rules__rule rules__rule--met' : 'rules__rule'}>
              <span className="rules__mark" aria-hidden="true" />
              {rule.label}
              <span className="visually-hidden">{rule.met ? ', done' : ', to do'}</span>
            </li>
          ))}
          {tooLong && <li className="rules__rule rules__rule--problem">At most 72 characters</li>}
        </ul>

        <label className="auth__check">
          <input
            type="checkbox"
            className="checkbox"
            checked={acceptTerms}
            onChange={(event) => setAcceptTerms(event.target.checked)}
          />
          <span>
            I agree to the{' '}
            <Link to="/terms" target="_blank">
              Terms
            </Link>{' '}
            and the{' '}
            <Link to="/privacy" target="_blank">
              Privacy Policy
            </Link>
            <span className="visually-hidden"> (each opens in a new tab)</span>
          </span>
        </label>

        <label className="auth__check">
          <input
            type="checkbox"
            className="checkbox"
            checked={marketing}
            onChange={(event) => setMarketing(event.target.checked)}
          />
          <span>Send me news about deals (optional)</span>
        </label>

        <button className="button auth__submit" disabled={submitting}>
          {submitting ? 'One moment…' : 'Create account'}
        </button>

        <p className="auth__other">
          Already have an account? <Link to={`/signin${nextQuery}`}>Log in</Link>
        </p>
      </form>
    </div>
  )
}
