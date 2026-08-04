import { readFileSync, readdirSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

/**
 * Garde-fou de complétude des catalogues i18n Nuxt.
 *
 * Il manquait totalement : c'est ainsi que `nav.adminInvoicing`, le bloc
 * `invoicing.*` et `pharmacy.daf.context*` avaient disparu de certaines locales
 * sans que rien n'échoue — l'UI affichait simplement la clé brute ou du
 * français. Les tests par écran (reading-type, productFlows, presentation) ne
 * couvrent qu'une poignée de clés énumérées à la main.
 */

const localesDir = join(__dirname, '../../locales')

/**
 * Locales réellement servies par la face Pro. Doit rester le miroir de
 * `i18n.locales` (nuxt.config.ts) et de `SUPPORTED_LOCALES` (useLocaleSync).
 * `fr` est la source de vérité.
 */
const LIVE_LOCALES = ['fr', 'nl', 'en', 'es', 'et', 'it'] as const
const TEMPLATE = 'fr'

/** Les trois familles de catalogues déclarées par locale dans nuxt.config.ts. */
const FAMILIES = ['.', 'presentation', 'ai-flows'] as const

type Json = Record<string, unknown>

function read(family: string, locale: string): Json {
  return JSON.parse(readFileSync(join(localesDir, family, `${locale}.json`), 'utf8'))
}

function exists(family: string, locale: string): boolean {
  try {
    readFileSync(join(localesDir, family, `${locale}.json`))
    return true
  } catch {
    return false
  }
}

/** Aplatit en chemins pointés ; les tableaux sont indexés pour comparer leur longueur. */
function flatten(value: unknown, prefix = ''): Map<string, unknown> {
  const out = new Map<string, unknown>()
  if (Array.isArray(value)) {
    value.forEach((v, i) => {
      for (const [k, val] of flatten(v, `${prefix}[${i}]`)) out.set(k, val)
    })
  } else if (value !== null && typeof value === 'object') {
    for (const [k, v] of Object.entries(value)) {
      const key = prefix ? `${prefix}.${k}` : k
      for (const [kk, val] of flatten(v, key)) out.set(kk, val)
    }
  } else {
    out.set(prefix, value)
  }
  return out
}

/** Jetons d'interpolation vue-i18n : `{n}`, `{email}`… (hors `{'@'}` échappé). */
function placeholders(value: unknown): string[] {
  return [...`${value}`.matchAll(/\{(\w+)\}/g)].map((m) => m[1]).sort()
}

/** Locales présentes sur disque mais pas encore servies (traduction en cours). */
function stagedLocales(): string[] {
  return readdirSync(localesDir)
    .filter((f) => f.endsWith('.json'))
    .map((f) => f.replace(/\.json$/, ''))
    .filter((loc) => !(LIVE_LOCALES as readonly string[]).includes(loc))
    .sort()
}

describe.each(FAMILIES)('catalogue %s', (family) => {
  const template = flatten(read(family, TEMPLATE))

  it('le template fr n\'est pas vide', () => {
    expect(template.size).toBeGreaterThan(0)
  })

  it.each(LIVE_LOCALES.filter((l) => l !== TEMPLATE))(
    '%s a exactement les mêmes clés que fr',
    (locale) => {
      const keys = new Set(flatten(read(family, locale)).keys())
      const missing = [...template.keys()].filter((k) => !keys.has(k))
      const extra = [...keys].filter((k) => !template.has(k))
      expect({ missing, extra }).toEqual({ missing: [], extra: [] })
    },
  )

  it.each(LIVE_LOCALES.filter((l) => l !== TEMPLATE))(
    '%s conserve les placeholders de fr',
    (locale) => {
      const cat = flatten(read(family, locale))
      const diverging: string[] = []
      for (const [key, frValue] of template) {
        const value = cat.get(key)
        if (value === undefined) continue
        if (placeholders(value).join(',') !== placeholders(frValue).join(',')) diverging.push(key)
      }
      expect(diverging).toEqual([])
    },
  )

  it.each(LIVE_LOCALES)('%s n\'a aucune valeur vide là où fr est renseigné', (locale) => {
    const cat = flatten(read(family, locale))
    const empty: string[] = []
    for (const [key, frValue] of template) {
      if (typeof frValue !== 'string' || frValue.trim() === '') continue
      const value = cat.get(key)
      if (typeof value === 'string' && value.trim() === '') empty.push(key)
    }
    expect(empty).toEqual([])
  })
})

/**
 * uk / ru : traduits partiellement et volontairement absents de nuxt.config.ts
 * (cf. useLocaleSync). On n'exige donc pas la parité — mais on interdit les
 * dérives qui rendraient la reprise fausse : clé inventée ou placeholder cassé.
 */
describe('locales en préparation (hors nuxt.config.ts)', () => {
  const staged = stagedLocales()

  it('ne sont pas servies par la face Pro', () => {
    for (const locale of staged) {
      expect(LIVE_LOCALES).not.toContain(locale)
    }
  })

  it.each(staged.length ? staged : ['(aucune)'])('%s : aucune clé inconnue, placeholders intacts', (locale) => {
    if (locale === '(aucune)') return
    for (const family of FAMILIES) {
      if (!exists(family, locale)) continue
      const template = flatten(read(family, TEMPLATE))
      const cat = flatten(read(family, locale))
      const extra = [...cat.keys()].filter((k) => !template.has(k))
      expect(extra, `${family}/${locale}.json contient des clés absentes de fr`).toEqual([])
      const diverging: string[] = []
      for (const [key, value] of cat) {
        const frValue = template.get(key)
        if (frValue === undefined) continue
        if (placeholders(value).join(',') !== placeholders(frValue).join(',')) diverging.push(key)
      }
      expect(diverging, `${family}/${locale}.json casse des placeholders`).toEqual([])
    }
  })
})
