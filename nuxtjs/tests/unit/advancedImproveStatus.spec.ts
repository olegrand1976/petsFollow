import { describe, expect, it } from 'vitest'
import {
  advancedImproveStateFromStep,
  isAdvancedImproveControlStep,
} from '~/utils/advancedImproveStatus'

describe('advancedImproveStatus', () => {
  it('detects control steps by stable state', () => {
    expect(isAdvancedImproveControlStep({ state: 'crew_warming' })).toBe(true)
    expect(isAdvancedImproveControlStep({ state: 'crew_ready' })).toBe(true)
    expect(isAdvancedImproveControlStep({ state: 'running' })).toBe(true)
    expect(isAdvancedImproveControlStep({})).toBe(false)
    expect(isAdvancedImproveControlStep({ state: undefined })).toBe(false)
    expect(isAdvancedImproveControlStep({ state: 'other' })).toBe(false)
  })

  it('maps step.state to AdvancedImproveState', () => {
    expect(advancedImproveStateFromStep({ state: 'crew_warming' })).toBe('crew_warming')
    expect(advancedImproveStateFromStep({ state: 'running' })).toBe('running')
    expect(advancedImproveStateFromStep({})).toBeNull()
    expect(advancedImproveStateFromStep({ state: 'bogus' })).toBeNull()
  })
})
