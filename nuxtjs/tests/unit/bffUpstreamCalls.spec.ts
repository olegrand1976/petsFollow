import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'

/**
 * Une route BFF qui construit elle-même son URL amont (`apiBase()`) et appelle
 * `fetch` / `$fetch.raw` avec le cookie de session court-circuite le rejeu de
 * `proxyApi` / `proxyBinary` : sur un access token expiré, elle renvoie 401 au
 * lieu de rafraîchir. Le contournement « toucher une route JSON d'abord » ne
 * marche pas non plus, `apiHeaders` relisant le cookie de la *requête*.
 *
 * Exemptions : le refresh lui-même, qui ne peut pas se réentrer, et les deux
 * téléchargements publics par token — sans session à rafraîchir, et surtout en
 * `responseType: 'stream'` pour ne pas bufferiser un dossier médical complet en
 * RAM, ce que `proxyBinary` ferait. Ne pas les migrer.
 */

const API_ROOT = join(__dirname, '..', '..', 'server', 'api')

const EXEMPT = new Set([
  'auth/refresh.post.ts',
  'public/consultation/[token]/download.get.ts',
  'public/pet-dossier/[token]/download.get.ts',
])

function routeFiles(dir: string, acc: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) {
      routeFiles(full, acc)
      continue
    }
    if (entry.endsWith('.ts')) acc.push(full)
  }
  return acc
}

describe('routes BFF — appels amont', () => {
  it('aucune route authentifiée n’appelle l’API sans le rejeu après refresh', () => {
    const offenders = routeFiles(API_ROOT)
      .filter((f) => {
        const rel = f.slice(API_ROOT.length + 1)
        if (EXEMPT.has(rel)) return false
        return readFileSync(f, 'utf8').includes('apiBase(')
      })
      .map(f => f.slice(API_ROOT.length + 1))

    expect(offenders).toEqual([])
  })
})
