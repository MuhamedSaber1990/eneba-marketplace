// Sellers are invented and have no logos. Each gets a round mark instead: the
// first letter of its name on one of a few colors. The color is worked out
// from the name, so nothing has to be stored.

const MARK_COLORS = ['#ffb347', '#63e3c2', '#ff8fab', '#9bdeff', '#c5f277', '#d9b8ff']

export type SellerMark = { letter: string; color: string }

/**
 * The marks for the sellers shown together on one page, in the same order.
 *
 * Each seller starts at the color its name points to. If a seller earlier in
 * the list already has that color, it moves on to the next free one. So two
 * sellers on a page only share a color when there are more sellers than
 * colors, and then the colors simply go round again.
 */
export function sellerMarks(sellerNames: string[]): SellerMark[] {
  const used = new Set<number>()

  return sellerNames.map((sellerName) => {
    const name = sellerName.trim()
    let index = colorIndex(name)

    if (used.size < MARK_COLORS.length) {
      while (used.has(index)) {
        index = (index + 1) % MARK_COLORS.length
      }
    }
    used.add(index)

    return { letter: name === '' ? '?' : name[0].toUpperCase(), color: MARK_COLORS[index] }
  })
}

/**
 * Adding up the character codes turns a name into a number. The remainder
 * after dividing by the number of colors picks one of them.
 */
function colorIndex(name: string): number {
  let sum = 0
  for (const character of name) {
    sum += character.codePointAt(0) ?? 0
  }
  return sum % MARK_COLORS.length
}
