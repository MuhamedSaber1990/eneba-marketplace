import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { debounce } from './debounce'

describe('debounce', () => {
  // Fake timers let the test move time forward instantly.
  beforeEach(() => vi.useFakeTimers())
  afterEach(() => vi.useRealTimers())

  it('does not run before the delay has passed', () => {
    const fn = vi.fn()
    debounce(fn, 250).call('a')
    vi.advanceTimersByTime(249)
    expect(fn).not.toHaveBeenCalled()
    vi.advanceTimersByTime(1)
    expect(fn).toHaveBeenCalledWith('a')
  })

  it('runs once, with the last value, when calls come quickly', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 250)
    for (const text of ['t', 'th', 'thu', 'thun']) {
      debounced.call(text)
      vi.advanceTimersByTime(100) // typing faster than the delay
    }
    vi.advanceTimersByTime(250)
    expect(fn).toHaveBeenCalledTimes(1)
    expect(fn).toHaveBeenCalledWith('thun')
  })

  it('runs again for a call made after a pause', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 250)
    debounced.call('kart')
    vi.advanceTimersByTime(250)
    debounced.call('kart xbox')
    vi.advanceTimersByTime(250)
    expect(fn).toHaveBeenCalledTimes(2)
  })

  it('cancel drops the waiting call', () => {
    const fn = vi.fn()
    const debounced = debounce(fn, 250)
    debounced.call('a')
    debounced.cancel()
    vi.advanceTimersByTime(1000)
    expect(fn).not.toHaveBeenCalled()
  })
})
