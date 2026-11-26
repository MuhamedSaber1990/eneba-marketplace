/** "2026-11-26T22:14:35Z" becomes "26 Nov 2026, 22:14" in the visitor's own time zone. */
export function formatDateTime(iso: string): string {
  return new Date(iso).toLocaleString('en-GB', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  })
}
