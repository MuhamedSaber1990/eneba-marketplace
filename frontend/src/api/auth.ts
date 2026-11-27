import { apiGet, apiPost } from './client'

/** A signed-in user, as the backend describes them. */
export type User = {
  id: number
  email: string
  name: string
  marketing_opt_in: boolean // the "send me news about deals" choice; nothing is ever sent
}

/** Everything the sign-up form sends. The server checks all of it again. */
export type SignUpForm = {
  name: string
  email: string
  password: string
  password_confirm: string
  accept_terms: boolean
  marketing_opt_in: boolean
}

type UserResponse = { user: User }

export async function signUp(form: SignUpForm): Promise<User> {
  return (await apiPost<UserResponse>('/api/auth/signup', form)).user
}

/** Changes the signed-in user's name and marketing choice, and returns the account as it now is. */
export async function updateAccount(name: string, marketingOptIn: boolean): Promise<User> {
  return (await apiPost<UserResponse>('/api/account', { name, marketing_opt_in: marketingOptIn })).user
}

export async function signIn(email: string, password: string): Promise<User> {
  return (await apiPost<UserResponse>('/api/auth/signin', { email, password })).user
}

export function signOut(): Promise<void> {
  return apiPost<void>('/api/auth/signout')
}

/**
 * Who is signed in, or null when nobody is. Being signed out is a normal
 * answer (200 with "user": null), not an error, so a visitor's browser
 * console stays clean.
 */
export async function getMe(signal?: AbortSignal): Promise<User | null> {
  return (await apiGet<{ user: User | null }>('/api/auth/me', signal)).user
}
