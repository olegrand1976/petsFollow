import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'
import fr from '../../locales/fr.json'
import en from '../../locales/en.json'
import nl from '../../locales/nl.json'
import es from '../../locales/es.json'
import et from '../../locales/et.json'
import itLocale from '../../locales/it.json'

const SECTION_KEYS = [
  'avatar',
  'profile',
  'research',
  'availability',
  'sites',
  'schedule',
  'vacations',
  'visitTypes',
  'notifications',
  'password',
  'language',
  'twoFa',
  'privacy',
  'legal',
] as const

describe('ProAccordionSection', () => {
  const src = readFileSync(
    join(dirname(fileURLToPath(import.meta.url)), '../../components/pro/ProAccordionSection.vue'),
    'utf8',
  )

  it('declares title, description, open and dataTestid props', () => {
    expect(src).toMatch(/title:\s*string/)
    expect(src).toMatch(/description\?:\s*string/)
    expect(src).toMatch(/open\?:\s*boolean/)
    expect(src).toMatch(/dataTestid\?:\s*string/)
  })

  it('uses native details/summary with visible description and expand icon', () => {
    expect(src).toContain('<details')
    expect(src).toContain('<summary')
    expect(src).toContain('pro-accordion__description')
    expect(src).toContain('name="expand_more"')
  })

  it('applies initial open state on mount without locking the toggle', () => {
    expect(src).toContain('onMounted')
    expect(src).toContain('ensureOpen')
    expect(src).toContain('watch(() => props.open, ensureOpen)')
    expect(src).toContain('detailsEl.value.open = true')
  })

  it('uses an h2 title for heading hierarchy', () => {
    expect(src).toContain('<h2 class="pro-accordion__title">')
  })
})

describe('settings.sections i18n', () => {
  const locales = { fr, en, nl, es, et, it: itLocale }

  it.each(Object.entries(locales))('%s has all section descriptions', (_code, locale) => {
    const sections = (locale as typeof fr).settings.sections
    expect(sections).toBeTruthy()
    for (const key of SECTION_KEYS) {
      expect(sections[key]?.description?.length, key).toBeGreaterThan(10)
    }
  })
})
