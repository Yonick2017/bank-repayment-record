import { describe, expect, it } from 'vitest'
import { sha256Hex, sha256HexFallback } from './sha256'

describe('sha256Hex', () => {
  it('matches known SHA-256 digests', async () => {
    expect(await sha256Hex('change-me')).toBe(
      'e2186dbdb1bb4193608605e84f33208765b5693b55edd4f730a719a100eeea6f',
    )
    expect(await sha256Hex('')).toBe(
      'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855',
    )
  })

  it('fallback matches SubtleCrypto when available', async () => {
    const value = 'shared-password-gate'
    expect(sha256HexFallback(value)).toBe(await sha256Hex(value))
  })

  it('uses fallback when crypto.subtle is unavailable', async () => {
    const original = globalThis.crypto
    Object.defineProperty(globalThis, 'crypto', {
      configurable: true,
      value: { subtle: undefined },
    })
    try {
      await expect(sha256Hex('change-me')).resolves.toBe(
        'e2186dbdb1bb4193608605e84f33208765b5693b55edd4f730a719a100eeea6f',
      )
    } finally {
      Object.defineProperty(globalThis, 'crypto', {
        configurable: true,
        value: original,
      })
    }
  })
})
