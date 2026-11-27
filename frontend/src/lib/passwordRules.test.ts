import { describe, expect, it } from 'vitest'
import { passwordRules, passwordTooLong } from './passwordRules'

/** The rules as { id: met }, which is easier to compare. */
function met(password: string, confirmation: string, email = 'layla@example.com') {
  return Object.fromEntries(passwordRules(password, confirmation, email).map((rule) => [rule.id, rule.met]))
}

describe('passwordRules', () => {
  it('meets nothing while the field is empty', () => {
    expect(met('', '')).toEqual({ length: false, 'not-email': false, match: false })
  })

  it('accepts a good password typed twice', () => {
    expect(met('correct horse battery', 'correct horse battery')).toEqual({ length: true, 'not-email': true, match: true })
  })

  it('wants 8 characters, and counts characters, not bytes', () => {
    expect(met('1234567', '1234567').length).toBe(false)
    expect(met('12345678', '12345678').length).toBe(true)
    // Seven emoji are seven characters, although they take 28 bytes.
    expect(met('😀😀😀😀😀😀😀', '').length).toBe(false)
  })

  it('refuses the email address as the password, however it is typed', () => {
    expect(met('layla@example.com', 'layla@example.com')['not-email']).toBe(false)
    expect(met('LAYLA@Example.com', '')['not-email']).toBe(false)
    expect(met(' layla@example.com ', '')['not-email']).toBe(false)
    expect(met('layla@example.com1', '')['not-email']).toBe(true)
  })

  it('wants the confirmation to match exactly', () => {
    expect(met('correct horse', 'correct horse').match).toBe(true)
    expect(met('correct horse', 'correct Horse').match).toBe(false)
    expect(met('correct horse', '').match).toBe(false)
  })
})

describe('passwordTooLong', () => {
  it('allows 72 bytes and refuses 73', () => {
    expect(passwordTooLong('a'.repeat(72))).toBe(false)
    expect(passwordTooLong('a'.repeat(73))).toBe(true)
  })

  it('counts bytes, so fewer characters fit in other alphabets', () => {
    // Each of these letters takes two bytes: 37 of them are 74 bytes.
    expect(passwordTooLong('é'.repeat(36))).toBe(false)
    expect(passwordTooLong('é'.repeat(37))).toBe(true)
  })
})
