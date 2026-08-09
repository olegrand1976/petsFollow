import { readFileSync } from 'node:fs'
import { join } from 'node:path'

/**
 * Locales réellement servies par la face Pro. Doit rester le miroir de
 * `i18n.locales` (nuxt.config.ts) et de `SUPPORTED_LOCALES` (useLocaleSync).
 * `fr` est la source de vérité.
 *
 * Déclaré une seule fois : deux specs qui recopiaient la liste pouvaient diverger
 * en silence, et c'est précisément la divergence que ces tests surveillent.
 */
export const LIVE_LOCALES = ['fr', 'nl', 'en', 'es', 'et', 'it'] as const

export const TEMPLATE_LOCALE = 'fr'

export const localesDir = join(__dirname, '../../../locales')

/** Catalogue d'une locale. `family` : `.`, `presentation` ou `ai-flows`. */
export function readLocale(locale: string, family = '.'): Record<string, any> {
  return JSON.parse(readFileSync(join(localesDir, family, `${locale}.json`), 'utf8'))
}
