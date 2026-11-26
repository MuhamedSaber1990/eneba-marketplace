/**
 * After signing in, a visitor is sent back to the page they wanted
 * (/signin?next=/account). The `next` value comes from the URL, so anyone can
 * put anything there. This returns it only if it is a path on this site, and
 * the homepage otherwise.
 *
 * Without the check, a link like /signin?next=https://evil.example would
 * send someone to another site right after they typed their password here,
 * which is a convincing start for a phishing page.
 */
export function safeNextPath(next: string | null): string {
  if (!next) return '/'

  // Must start with exactly one slash. "//evil.example" and "/\evil.example"
  // are both read by browsers as another site.
  if (!next.startsWith('/') || next.startsWith('//') || next.startsWith('/\\')) return '/'

  // Control characters such as a newline or tab have no place in a path.
  // Browsers strip some of them, which could turn "/\t/evil" into "//evil".
  for (const character of next) {
    if (character.charCodeAt(0) < 32) return '/'
  }

  return next
}
