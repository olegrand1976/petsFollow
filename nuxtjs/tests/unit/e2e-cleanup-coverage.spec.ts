import fs from 'node:fs'
import path from 'node:path'
import { describe, expect, it } from 'vitest'

/**
 * Garde-fou anti-régression : tout email éphémère créé par les tests e2e / smoke
 * (uniqueE2EEmail, littéraux timestampés, smoke-test.sh) doit être purgé par
 * scripts/cleanup-staging-quality.sh — sinon il s'accumule en DB staging.
 * Nouveau préfixe e2e → ajouter le motif LIKE correspondant dans le script.
 */

const repoRoot = path.resolve(__dirname, '../../..')
const e2eDir = path.join(repoRoot, 'nuxtjs/tests/e2e')
const cleanupScript = path.join(repoRoot, 'scripts/cleanup-staging-quality.sh')
const smokeScript = path.join(repoRoot, 'scripts/smoke-test.sh')

function listTsFiles(dir: string): string[] {
  const out: string[] = []
  for (const entry of fs.readdirSync(dir, { withFileTypes: true })) {
    const full = path.join(dir, entry.name)
    if (entry.isDirectory()) out.push(...listTsFiles(full))
    else if (entry.name.endsWith('.ts')) out.push(full)
  }
  return out
}

type EmailSep = '+' | '-' | '.'
type EmailPattern = { prefix: string; sep: EmailSep; source: string }

function collectE2eEmailPatterns(): EmailPattern[] {
  const patterns = new Map<string, EmailPattern>()
  const add = (prefix: string, sep: EmailSep, source: string) => {
    patterns.set(`${prefix}${sep}`, { prefix, sep, source })
  }
  for (const file of listTsFiles(e2eDir)) {
    const src = fs.readFileSync(file, 'utf8')
    const rel = path.relative(repoRoot, file)
    // uniqueE2EEmail('prefix') → prefix+{ts}@petsfollow.test
    for (const m of src.matchAll(/uniqueE2EEmail\(\s*'([^']+)'\s*\)/g)) {
      add(m[1], '+', rel)
    }
    // uniqueE2EEmail() sans argument → défaut 'e2e'
    if (/uniqueE2EEmail\(\s*\)/.test(src)) add('e2e', '+', rel)
    // Littéraux template : `prefix{+|-|.}${stamp}@petsfollow.test`
    for (const m of src.matchAll(/`([a-z0-9.-]+?)([+.-])\$\{[^}]+\}@petsfollow\.test`/gi)) {
      add(m[1], m[2] as EmailSep, rel)
    }
  }
  // smoke-test.sh : smoke+$(date +%s)@petsfollow.test etc.
  const smokeSrc = fs.readFileSync(smokeScript, 'utf8')
  for (const m of smokeSrc.matchAll(/([a-z0-9-]+)\+\$\(date \+%s\)@petsfollow\.test/g)) {
    add(m[1], '+', path.relative(repoRoot, smokeScript))
  }
  return [...patterns.values()]
}

describe('e2e cleanup coverage', () => {
  const cleanupSql = fs.readFileSync(cleanupScript, 'utf8')
  const patterns = collectE2eEmailPatterns()

  it('trouve les préfixes emails éphémères dans les specs', () => {
    const keys = patterns.map((p) => `${p.prefix}${p.sep}`)
    // Sanity : les motifs historiques doivent être détectés (sinon le scan est cassé).
    expect(keys).toContain('smoke+')
    expect(keys).toContain('register+')
    expect(keys).toContain('pw-vet+')
    expect(keys).toContain('hr-dur-vet-')
  })

  it.each(patterns)(
    'purge le motif $prefix$sep%@petsfollow.test (créé par $source)',
    ({ prefix, sep }) => {
      expect(cleanupSql).toContain(`'${prefix}${sep}%@petsfollow.test'`)
    },
  )
})
