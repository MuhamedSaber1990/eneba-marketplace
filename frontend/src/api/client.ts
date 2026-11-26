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

/** Turns a response into its JSON body, or throws an ApiError. */
async function readResponse<T>(response: Response): Promise<T> {
  // 204 No Content has no body to parse.
  const body: unknown = response.status === 204 ? null : await response.json().catch(() => null)

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

/**
 * GET a JSON resource from the API.
 * Paths are relative ("/api/products"): the frontend and the API share a
 * domain on Vercel, and Vite forwards /api to the Go server in development.
 * Because it is the same site, the browser attaches the session cookie by
 * itself.
 */
export async function apiGet<T>(path: string, signal?: AbortSignal): Promise<T> {
  return readResponse<T>(await fetch(path, { signal }))
}

/**
 * POST JSON to the API. The backend only accepts changes sent as
 * application/json from this site's own pages.
 */
export async function apiPost<T>(path: string, body?: unknown): Promise<T> {
  const response = await fetch(path, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: body === undefined ? undefined : JSON.stringify(body),
  })
  return readResponse<T>(response)
}
