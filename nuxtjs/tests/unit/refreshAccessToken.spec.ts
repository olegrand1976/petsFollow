import { describe, expect, it } from 'vitest'
import { classifyRefreshHttpStatus } from '../../server/utils/api'

describe('classifyRefreshHttpStatus', () => {
  it('traite 401/403 comme rejet de session (purge cookies)', () => {
    expect(classifyRefreshHttpStatus(401)).toBe('rejected')
    expect(classifyRefreshHttpStatus(403)).toBe('rejected')
  })

  it('traite 5xx / inconnu / réseau comme transient (conserver cookies)', () => {
    expect(classifyRefreshHttpStatus(500)).toBe('transient')
    expect(classifyRefreshHttpStatus(502)).toBe('transient')
    expect(classifyRefreshHttpStatus(503)).toBe('transient')
    expect(classifyRefreshHttpStatus(undefined)).toBe('transient')
  })
})
