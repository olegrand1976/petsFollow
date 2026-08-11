import { describe, expect, it } from 'vitest'
import {
  canAdvanceSupportStatus,
  isValidSupportStatusTransition,
  SUPPORT_WORKFLOW_ORDER,
} from '~/utils/support-status'

describe('support-status', () => {
  it('autorise le parcours nominal et refuse les sauts', () => {
    expect(isValidSupportStatusTransition('open', 'in_progress')).toBe(true)
    expect(isValidSupportStatusTransition('in_progress', 'to_test')).toBe(true)
    expect(isValidSupportStatusTransition('to_test', 'done')).toBe(true)
    expect(isValidSupportStatusTransition('done', 'closed')).toBe(true)
    expect(isValidSupportStatusTransition('open', 'done')).toBe(false)
    expect(isValidSupportStatusTransition('done', 'to_test')).toBe(false)
    expect(isValidSupportStatusTransition('closed', 'open')).toBe(true)
  })

  it('expose les 5 étapes et les avances cliquables', () => {
    expect(SUPPORT_WORKFLOW_ORDER).toEqual(['open', 'in_progress', 'to_test', 'done', 'closed'])
    expect(canAdvanceSupportStatus('open', 'in_progress')).toBe(true)
    expect(canAdvanceSupportStatus('open', 'closed')).toBe(true)
    expect(canAdvanceSupportStatus('open', 'done')).toBe(false)
    expect(canAdvanceSupportStatus('open', 'open')).toBe(false)
  })
})
