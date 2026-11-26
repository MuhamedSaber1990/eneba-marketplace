/**
 * Wraps a function so it runs only after calls have stopped for `delayMs`.
 * Each new call restarts the wait and replaces the earlier arguments.
 *
 * The search box uses it so that typing "thunder" sends one request, for the
 * finished word, and not seven.
 */
export function debounce<Args extends unknown[]>(fn: (...args: Args) => void, delayMs: number) {
  let timer: ReturnType<typeof setTimeout> | undefined

  return {
    /** Schedules fn with these arguments, dropping any call still waiting. */
    call(...args: Args) {
      clearTimeout(timer)
      timer = setTimeout(() => fn(...args), delayMs)
    },
    /** Drops the waiting call, if there is one. */
    cancel() {
      clearTimeout(timer)
    },
  }
}
