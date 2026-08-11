import { describe, expect, it } from 'vitest'
import {
  isValidSupportStatusTransition,
  supportStatusSelectOptions,
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

  it('expose current + next pour le select', () => {
    expect(supportStatusSelectOptions('done')).toEqual(['done', 'closed'])
    expect(supportStatusSelectOptions('open')).toEqual(['open', 'in_progress', 'closed'])
  })
})
