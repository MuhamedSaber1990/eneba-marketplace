import { useEffect, useState } from 'react'
import { ApiError } from '../api/client'

/** The three things a request can be: still running, finished, or failed. */
export type AsyncState<T> =
  | { status: 'loading' }
  | { status: 'ready'; data: T }
  | { status: 'error'; message: string; httpStatus?: number }

/**
 * Runs an async load function and tracks its state. It runs again whenever
 * `key` changes (for example the query string of the store page).
 *
 * If the key changes or the component goes away while a request is running,
 * that request is cancelled, so an old slow response can never overwrite a
 * newer one.
 */
export function useAsync<T>(load: (signal: AbortSignal) => Promise<T>, key: string): AsyncState<T> {
  // The outcome is stored together with the key it belongs to.
  const [outcome, setOutcome] = useState<{ key: string; state: AsyncState<T> } | null>(null)

  useEffect(() => {
    const controller = new AbortController()

    load(controller.signal)
      .then((data) => setOutcome({ key, state: { status: 'ready', data } }))
      .catch((error: unknown) => {
        if (controller.signal.aborted) return // cancelled on purpose, not a failure
        setOutcome({
          key,
          state: {
            status: 'error',
            message: error instanceof Error ? error.message : 'Something went wrong',
            httpStatus: error instanceof ApiError ? error.status : undefined,
          },
        })
      })

    return () => controller.abort()
    // `load` is a new function on every render, so only `key` decides when
    // to run again. The lint rule that wants `load` listed is switched off
    // for this line.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [key])

  // An outcome for an older key is stale: the request for the current key is
  // still on its way, so the state is "loading".
  if (outcome === null || outcome.key !== key) {
    return { status: 'loading' }
  }
  return outcome.state
}
