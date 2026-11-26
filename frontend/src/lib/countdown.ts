const MINUTE = 60 * 1000
const HOUR = 60 * MINUTE
const DAY = 24 * HOUR

/**
 * Says how long is left until a deal ends, using the two largest units:
 * "3d 4h", "5h 12m", "12m". Returns null when the time has passed.
 */
export function timeLeft(endsAt: Date, now: Date): string | null {
  const ms = endsAt.getTime() - now.getTime()
  if (Number.isNaN(ms) || ms <= 0) return null

  const days = Math.floor(ms / DAY)
  const hours = Math.floor((ms % DAY) / HOUR)
  const minutes = Math.floor((ms % HOUR) / MINUTE)

  if (days > 0) return `${days}d ${hours}h`
  if (hours > 0) return `${hours}h ${minutes}m`
  if (minutes > 0) return `${minutes}m`
  return 'less than a minute'
}
