import { createContext, useContext } from 'react'
import type { User } from '../api/auth'

/**
 * Who is signed in, shared with every component.
 *
 * The session itself lives in an HttpOnly cookie that JavaScript cannot
 * read. So the frontend does not "have" the session: it asks the backend who
 * the cookie belongs to (GET /api/auth/me) when the page loads, and
 * AuthProvider remembers the answer.
 */
export type AuthState = {
  /** True until the first "who am I" request has been answered. */
  loading: boolean
  /** The signed-in user, or null. */
  user: User | null
  signIn: (email: string, password: string) => Promise<void>
  signUp: (email: string, password: string) => Promise<void>
  signOut: () => Promise<void>
}

export const AuthContext = createContext<AuthState | null>(null)

/** Read the signed-in user and the sign in, sign up and sign out actions. */
export function useAuth(): AuthState {
  const state = useContext(AuthContext)
  if (state === null) {
    throw new Error('useAuth must be used inside <AuthProvider>')
  }
  return state
}
