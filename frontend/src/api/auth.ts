import { apiGet, apiPost } from './client'

/** A signed-in user, as the backend describes them. */
export type User = {
  id: number
  email: string
}

type UserResponse = { user: User }

export async function signUp(email: string, password: string): Promise<User> {
  return (await apiPost<UserResponse>('/api/auth/signup', { email, password })).user
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
