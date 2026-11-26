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

/** Who is signed in. Throws an ApiError with status 401 when nobody is. */
export async function getMe(signal?: AbortSignal): Promise<User> {
  return (await apiGet<UserResponse>('/api/auth/me', signal)).user
}
