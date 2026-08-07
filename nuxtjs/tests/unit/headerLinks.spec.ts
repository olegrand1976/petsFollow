import { describe, expect, it } from 'vitest'
import {
  buildUnifiedRows,
  isCompleteCustom,
  isValidHttpsUrl,
  moveRow,
  prefsFromUnifiedRows,
  validateHeaderLinksCustoms,
  type HeaderLinkCatalogRow,
  type HeaderLinksPrefs,
} from '../../utils/headerLinks'

describe('headerLinks utils', () => {
  const catalog: HeaderLinkCatalogRow[] = [
    { id: 'be_ordre', label: 'Ordre', url: 'https://ordre.example/', enabled: true },
    { id: 'be_afmps', label: 'AFMPS', url: 'https://afmps.example/', enabled: false },
    { id: 'be_vetcompendium', label: 'Vetcomp', url: 'https://vetcomp.example/', enabled: true },
  ]

  it('rejects non-https urls', () => {
    expect(isValidHttpsUrl('http://lab.example/')).toBe(false)
    expect(isValidHttpsUrl('https://')).toBe(false)
    expect(isValidHttpsUrl('https://lab.example/path')).toBe(true)
  })

  it('validates incomplete and duplicate customs', () => {
    expect(validateHeaderLinksCustoms([{ id: 'custom_1', label: '', url: 'https://' }])).toBe('incompleteCustom')
    expect(validateHeaderLinksCustoms([{ id: 'custom_1', label: 'Lab', url: 'http://x' }])).toBe('urlInvalid')
    expect(validateHeaderLinksCustoms([
      { id: 'custom_1', label: 'A', url: 'https://lab.example/' },
      { id: 'custom_2', label: 'B', url: 'https://lab.example/' },
    ])).toBe('duplicateUrl')
    expect(validateHeaderLinksCustoms([
      { id: 'custom_1', label: 'Lab', url: 'https://lab.example/' },
    ])).toBeNull()
  })

  it('builds interleaved order from prefs.order', () => {
    const prefs: HeaderLinksPrefs = {
      configured: true,
      enabled: ['be_ordre', 'custom_lab', 'be_vetcompendium'],
      order: ['be_ordre', 'custom_lab', 'be_vetcompendium'],
      custom: [{ id: 'custom_lab', label: 'Lab', url: 'https://lab.example/' }],
    }
    // Sync catalog enabled flags from prefs for the test.
    const cat = catalog.map((c) => ({
      ...c,
      enabled: prefs.enabled.includes(c.id),
    }))
    const rows = buildUnifiedRows(cat, prefs)
    expect(rows.map((r) => r.id)).toEqual(['be_ordre', 'custom_lab', 'be_vetcompendium', 'be_afmps'])
    expect(rows[1].kind).toBe('custom')
  })

  it('prefsFromUnifiedRows preserves interleaved order of enabled+custom', () => {
    const rows = buildUnifiedRows(
      catalog.map((c) => ({ ...c, enabled: c.id !== 'be_afmps' })),
      {
        configured: true,
        enabled: ['be_ordre', 'custom_lab', 'be_vetcompendium'],
        order: ['be_ordre', 'custom_lab', 'be_vetcompendium'],
        custom: [{ id: 'custom_lab', label: 'Lab', url: 'https://lab.example/' }],
      },
    )
    // Move custom between catalog entries already in order; move afmps disabled stays in list.
    const moved = moveRow(rows, 1, 1) // custom after vetcompendium... wait order is ordre, custom, vetcomp, afmps
    // Move custom down: ordre, vetcomp, custom, afmps
    const prefs = prefsFromUnifiedRows(moved)
    expect(prefs.order).toEqual(['be_ordre', 'be_vetcompendium', 'custom_lab'])
    expect(prefs.custom).toEqual([{ id: 'custom_lab', label: 'Lab', url: 'https://lab.example/' }])
    expect(prefs.enabled).toContain('custom_lab')
    expect(prefs.enabled).not.toContain('be_afmps')
  })

  it('isCompleteCustom', () => {
    expect(isCompleteCustom({ label: 'Lab', url: 'https://lab.example/' })).toBe(true)
    expect(isCompleteCustom({ label: '', url: 'https://lab.example/' })).toBe(false)
  })
})
