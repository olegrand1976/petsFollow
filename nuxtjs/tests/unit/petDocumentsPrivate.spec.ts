import { readFileSync, readdirSync, statSync } from 'node:fs'
import { join } from 'node:path'
import { describe, expect, it } from 'vitest'
import { petDocumentHref } from '../../utils/petDocuments'

describe('petDocumentHref', () => {
  it('pointe sur la route BFF authentifiée, jamais sur le stockage', () => {
    const href = petDocumentHref('pet-1', 'doc-9')
    expect(href).toBe('/api/pets/pet-1/documents/doc-9/download')
    expect(href).not.toContain('storage.googleapis.com')
    expect(href).not.toContain('/media/')
  })

  it('encode les identifiants pour éviter une évasion de chemin', () => {
    expect(petDocumentHref('a/../b', 'c d')).toBe('/api/pets/a%2F..%2Fb/documents/c%20d/download')
  })
})

const ROOTS = ['pages', 'components', 'composables', 'server', 'utils']
const EXTS = ['.vue', '.ts', '.js']

function sourceFiles(dir: string, acc: string[] = []): string[] {
  for (const entry of readdirSync(dir)) {
    const full = join(dir, entry)
    if (statSync(full).isDirectory()) {
      sourceFiles(full, acc)
      continue
    }
    if (EXTS.some(e => entry.endsWith(e))) acc.push(full)
  }
  return acc
}

/**
 * L'API ne renvoie plus `fileUrl` pour les documents animaux (PHI) : le champ
 * est marqué `json:"-"` côté Go et la colonne reste vide. Toute réapparition
 * dans la face Pro signale qu'on a re-branché une URL de stockage.
 */
describe('documents animaux — pas d’URL de stockage côté Pro', () => {
  it('aucune source Nuxt ne consomme fileUrl', () => {
    const root = join(__dirname, '..', '..')
    const offenders = ROOTS
      .flatMap(r => sourceFiles(join(root, r)))
      .filter(f => readFileSync(f, 'utf8').includes('fileUrl'))
      .map(f => f.slice(root.length + 1))

    expect(offenders).toEqual([])
  })
})
