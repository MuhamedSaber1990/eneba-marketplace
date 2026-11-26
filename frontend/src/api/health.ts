/** Shape of GET /api/health (see backend/httpapi/health.go). */
export type Health = {
  status: 'ok' | 'degraded'
  database: 'ok' | 'not configured' | 'unreachable'
}

/**
 * Reads the health check. The endpoint answers 503 with the same JSON body
 * when the database is down, so this uses fetch directly instead of apiGet,
 * which would turn the 503 into an exception and lose the body.
 */
export async function getHealth(signal?: AbortSignal): Promise<Health> {
  const response = await fetch('/api/health', { signal })
  return (await response.json()) as Health
}
