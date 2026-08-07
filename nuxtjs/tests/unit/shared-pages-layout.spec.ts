import { readFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'
import { describe, expect, it } from 'vitest'

const here = dirname(fileURLToPath(import.meta.url))

/**
 * Pages liées depuis plusieurs shells (default / admin / commercial / manager) :
 * sans middleware de layout, un admin qui clique voit la sidebar véto (layout `default`).
 */
const SHARED_PAGES = ['produits.vue', 'nouveautes.vue'] as const

describe('pages partagées multi-layouts', () => {
  for (const page of SHARED_PAGES) {
    it(`${page} pose le layout selon le rôle (middleware shared-pro-layout)`, () => {
      const src = readFileSync(join(here, '../../pages', page), 'utf8')
      expect(src).toContain("'shared-pro-layout'")
    })
  }
})
