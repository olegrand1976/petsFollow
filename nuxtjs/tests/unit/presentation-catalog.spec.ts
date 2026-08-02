import { describe, expect, it } from 'vitest'
import {
  PRESENTATION_STEPS,
  parsePresentationStepId,
} from '../../data/presentation/catalog'
import {
  aiFlowsCatalog,
  parseAiFlowProfileId,
  diagramsForProfile,
  fillMermaidTemplate,
} from '../../data/ai-flows/catalog'

describe('presentation catalog', () => {
  it('expose 10 steps stables', () => {
    expect(PRESENTATION_STEPS).toHaveLength(10)
    expect(PRESENTATION_STEPS[0]).toBe('welcome')
    expect(PRESENTATION_STEPS).toContain('ai_in_app')
    expect(PRESENTATION_STEPS).toContain('ai_automation')
  })

  it('parsePresentationStepId canonise', () => {
    expect(parsePresentationStepId('vetpro')).toBe('vetpro')
    expect(parsePresentationStepId('nope')).toBeNull()
    expect(parsePresentationStepId(['pain', 'x'])).toBe('pain')
    expect(parsePresentationStepId(undefined)).toBeNull()
  })
})

describe('ai-flows catalog', () => {
  it('met le profil vet en premier avec 3 diagrammes', () => {
    expect(aiFlowsCatalog.profiles[0]?.id).toBe('vet')
    expect(diagramsForProfile('vet')).toHaveLength(3)
  })

  it('parseAiFlowProfileId', () => {
    expect(parseAiFlowProfileId('vet')).toBe('vet')
    expect(parseAiFlowProfileId('nope')).toBeNull()
    expect(parseAiFlowProfileId(['commercial'])).toBe('commercial')
  })

  it('fillMermaidTemplate remplace les placeholders', () => {
    const out = fillMermaidTemplate('A["{{start}}"] --> B["{{end}}"]', {
      start: 'Hello "world"',
      end: 'Done',
    })
    expect(out).toBe('A["Hello \'world\'"] --> B["Done"]')
  })
})
