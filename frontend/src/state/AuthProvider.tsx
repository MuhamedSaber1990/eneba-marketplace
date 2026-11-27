import { type ReactNode, useEffect, useState } from 'react'
import * as authApi from '../api/auth'
import type { User } from '../api/auth'
import { AuthContext, type AuthState } from './auth'

/**
 * Asks the backend once who is signed in, and makes the answer and the sign
 * in, sign up and sign out actions available to everything inside it.
 */
export function AuthProvider({ children }: { children: ReactNode }) {
  const [loading, setLoading] = useState(true)
  const [user, setUser] = useState<User | null>(null)

  useEffect(() => {
    const controller = new AbortController()
    authApi
      .getMe(controller.signal)
      .then(setUser)
      // The answer is the user, or null for a visitor who is not signed in.
      // If the request itself fails, the visitor is treated as signed out.
      .catch(() => setUser(null))
      .finally(() => {
        if (!controller.signal.aborted) setLoading(false)
      })
    return () => controller.abort()
  }, [])

  const value: AuthState = {
    loading,
    user,
    async signIn(email, password) {
      setUser(await authApi.signIn(email, password))
    },
    async signUp(form) {
      setUser(await authApi.signUp(form))
    },
    async updateProfile(name, marketingOptIn) {
      // The server's answer replaces the remembered user, so the header
      // shows the new name at once.
      setUser(await authApi.updateAccount(name, marketingOptIn))
    },
    async signOut() {
      await authApi.signOut()
      setUser(null)
    },
  }

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>
}
