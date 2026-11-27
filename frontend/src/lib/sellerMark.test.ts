import { describe, expect, it } from 'vitest'
import { sellerMarks } from './sellerMark'

describe('sellerMarks', () => {
  it('uses the first letter, in upper case', () => {
    expect(sellerMarks(['lootLantern', '  GG Depot']).map((mark) => mark.letter)).toEqual(['L', 'G'])
  })

  it('gives the same list the same colors every time', () => {
    const names = ['GG Depot', 'LootLantern', 'Arcadia Keys']
    expect(sellerMarks(names)).toEqual(sellerMarks(names))
  })

  it('gives two sellers different colors even when their names point to the same one', () => {
    // These two were the same green on a product page before this rule.
    const [first, second] = sellerMarks(['GG Depot', 'LootLantern'])
    expect(first.color).not.toBe(second.color)

    // The same name twice is the plainest clash there is.
    const [a, b] = sellerMarks(['Same Name', 'Same Name'])
    expect(a.color).not.toBe(b.color)
  })

  it('keeps six sellers on six different colors', () => {
    const names = ['GG Depot', 'LootLantern', 'Arcadia Keys', 'NorthGate Games', 'PixelPort', 'KeyHarbor']
    const colors = new Set(sellerMarks(names).map((mark) => mark.color))
    expect(colors.size).toBe(6)
  })

  it('still answers when there are more sellers than colors', () => {
    const marks = sellerMarks(['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H'])
    expect(marks).toHaveLength(8)
    expect(marks.every((mark) => mark.color.startsWith('#'))).toBe(true)
  })

  it('copes with an empty name', () => {
    expect(sellerMarks([''])[0].letter).toBe('?')
  })
})
