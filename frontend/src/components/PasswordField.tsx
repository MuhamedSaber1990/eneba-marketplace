import { useState } from 'react'

type Props = {
  id: string
  value: string
  onChange: (value: string) => void
  /** Tells a password manager whether to offer a saved password or suggest a new one. */
  autoComplete: 'current-password' | 'new-password'
  /** The id of text that describes the field, such as the list of rules. */
  describedBy?: string
  /** What the show button refers to, for screen readers: "password" or "confirmation". */
  what?: string
}

/**
 * A password input with a button to show what was typed. Showing it helps
 * people catch a typo, which matters most on a phone.
 *
 * The button is type="button" so that pressing it does not submit the form.
 * aria-pressed tells a screen reader whether the password is currently shown.
 */
export function PasswordField({ id, value, onChange, autoComplete, describedBy, what = 'password' }: Props) {
  const [shown, setShown] = useState(false)

  return (
    <div className="password-field">
      <input
        id={id}
        type={shown ? 'text' : 'password'}
        autoComplete={autoComplete}
        required
        aria-describedby={describedBy}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      />
      <button type="button" aria-pressed={shown} aria-label={`Show the ${what}`} onClick={() => setShown(!shown)}>
        {shown ? 'Hide' : 'Show'}
      </button>
    </div>
  )
}
