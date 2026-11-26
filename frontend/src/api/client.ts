// One place for talking to the backend. Every API file builds on apiGet, so
// error handling is written once.

/** An error response from the API, with its HTTP status. */
export class ApiError extends Error {
  status: number

  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

/**
 * GET a JSON resource from the API.
 * Paths are relative ("/api/products"): the frontend and the API share a
 * domain on Vercel, and Vite forwards /api to the Go server in development.
 */
export async function apiGet<T>(path: string, signal?: AbortSignal): Promise<T> {
  const response = await fetch(path, { signal })
  const body: unknown = await response.json().catch(() => null)

  if (!response.ok) {
    // The API sends {"error": "..."} on failure.
    const message =
      body !== null && typeof body === 'object' && 'error' in body
        ? String(body.error)
        : `Request failed with status ${response.status}`
    throw new ApiError(response.status, message)
  }
  return body as T
}
