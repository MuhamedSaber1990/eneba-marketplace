import { describe, expect, it } from 'vitest'
import { sellerMark } from './sellerMark'

describe('sellerMark', () => {
  it('uses the first letter, in upper case', () => {
    expect(sellerMark('lootLantern').letter).toBe('L')
    expect(sellerMark('  GG Depot').letter).toBe('G')
  })

  it('gives the same seller the same color every time', () => {
    expect(sellerMark('GG Depot')).toEqual(sellerMark('GG Depot'))
  })

  it('does not give every seller the same color', () => {
    const names = ['GG Depot', 'LootLantern', 'Arcadia Keys', 'NorthGate Games', 'PixelPort', 'KeyHarbor']
    const colors = new Set(names.map((name) => sellerMark(name).color))
    expect(colors.size).toBeGreaterThan(1)
  })

  it('copes with an empty name', () => {
    expect(sellerMark('').letter).toBe('?')
  })
})
