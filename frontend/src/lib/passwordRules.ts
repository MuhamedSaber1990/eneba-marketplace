// The rules for a new password, as the sign-up form shows them while the
// visitor types. The server checks the same rules again
// (backend/auth/password.go, ValidateNewPassword), and its answer is the one
// that counts: this file only saves the visitor a round trip.
//
// There is deliberately no "one capital, one number, one symbol" rule.
// Current guidance (NIST SP 800-63B) advises against such rules: people
// satisfy them with "Password1!" and the password is no harder to guess.
// Length is what helps.

export const MIN_PASSWORD_CHARS = 8
export const MAX_PASSWORD_BYTES = 72 // bcrypt reads no further than this

export type PasswordRule = {
  id: string
  label: string
  met: boolean
}

export function passwordRules(password: string, confirmation: string, email: string): PasswordRule[] {
  const typed = password !== ''
  return [
    {
      id: 'length',
      label: `At least ${MIN_PASSWORD_CHARS} characters`,
      // Spreading a string into an array counts characters the way a person
      // does. `.length` would count an emoji as two.
      met: [...password].length >= MIN_PASSWORD_CHARS,
    },
    {
      id: 'not-email',
      label: 'Not your email address',
      met: typed && password.trim().toLowerCase() !== email.trim().toLowerCase(),
    },
    {
      id: 'match',
      label: 'The two passwords match',
      met: typed && password === confirmation,
    },
  ]
}

/** True when the password is longer than the server accepts. Rare, so it is only mentioned when it happens. */
export function passwordTooLong(password: string): boolean {
  return new TextEncoder().encode(password).length > MAX_PASSWORD_BYTES
}
