import { describe, expect, it } from 'vitest'
import { safeNextPath } from './safeNext'

describe('safeNextPath', () => {
  it.each([
    ['/account', '/account'],
    ['/store?platform=Steam&page=2', '/store?platform=Steam&page=2'],
    ['/product/ashen-crown-xbox-key-global', '/product/ashen-crown-xbox-key-global'],
    ['/', '/'],
  ])('keeps a path on this site: %s', (next, expected) => {
    expect(safeNextPath(next)).toBe(expected)
  })

  it.each([
    ['https://evil.example/login'],
    ['http://evil.example'],
    ['//evil.example'],
    ['/\\evil.example'],
    ['evil.example'],
    ['javascript:alert(1)'],
    ['/\t/evil.example'],
    ['/\n/evil.example'],
    [''],
  ])('falls back to the homepage for %j', (next) => {
    expect(safeNextPath(next)).toBe('/')
  })

  it('falls back to the homepage when there is no next at all', () => {
    expect(safeNextPath(null)).toBe('/')
  })
})
