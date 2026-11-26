import { describe, expect, it } from 'vitest'
import { timeLeft } from './countdown'

const now = new Date('2026-12-01T12:00:00Z')
const after = (ms: number) => new Date(now.getTime() + ms)

const MINUTE = 60_000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

describe('timeLeft', () => {
  it.each([
    ['days and hours', 3 * DAY + 4 * HOUR + 59 * MINUTE, '3d 4h'],
    ['exactly one day', DAY, '1d 0h'],
    ['hours and minutes', 5 * HOUR + 12 * MINUTE, '5h 12m'],
    ['just under a day', DAY - MINUTE, '23h 59m'],
    ['minutes only', 12 * MINUTE + 30_000, '12m'],
    ['under a minute', 20_000, 'less than a minute'],
  ])('%s', (_name, ms, expected) => {
    expect(timeLeft(after(ms), now)).toBe(expected)
  })

  it('is null once the end has passed', () => {
    expect(timeLeft(after(-1000), now)).toBeNull()
    expect(timeLeft(now, now)).toBeNull()
  })

  it('is null for a date that could not be read', () => {
    expect(timeLeft(new Date('not a date'), now)).toBeNull()
  })
})
