// Sellers are invented and have no logos. Each gets a round mark instead: the
// first letter of its name on one of a few colors. The color is worked out
// from the name, so a seller looks the same on every page and nothing has to
// be stored.

const MARK_COLORS = ['#ffb347', '#63e3c2', '#ff8fab', '#9bdeff', '#c5f277', '#d9b8ff']

export type SellerMark = { letter: string; color: string }

export function sellerMark(sellerName: string): SellerMark {
  const name = sellerName.trim()
  if (name === '') return { letter: '?', color: MARK_COLORS[0] }

  // Adding up the character codes turns the name into a number. The
  // remainder after dividing by the number of colors picks one of them.
  let sum = 0
  for (const character of name) {
    sum += character.codePointAt(0) ?? 0
  }
  return { letter: name[0].toUpperCase(), color: MARK_COLORS[sum % MARK_COLORS.length] }
}
