import { test, expect } from '@playwright/test'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { loginAsAdmin } from '../helpers/auth'

/**
 * AFMPS CSV admin — navigation + triple contrôle UI (validate → revue → commit).
 * Tag @p1 @pharmacy. Ne coche pas deactivate-missing (préserve le catalogue seed).
 */
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

const FIXTURE = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '../fixtures/afmps-mini.csv',
)

/** Mini-CSV avec CNK uniques pour éviter collisions entre runs parallèles. */
function writeUniqueMiniCsv(): { filePath: string; cnk1: string; cnk2: string } {
  const stamp = `${Date.now()}`.slice(-5)
  const cnk1 = `29${stamp}`
  const cnk2 = `28${stamp}`
  const csv = [
    "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Code ATC;Usage Humain/Vétérinaire",
    `E2E AFMPS Alpha ${stamp};Gélule;10;${cnk1};Lab E2E;BE-E2E-A;Oui;QJ01CA04;Usage vétérinaire`,
    `E2E AFMPS Beta ${stamp};Comprimé;20;${cnk2};Lab E2E;BE-E2E-B;Oui;QH02CA01;Usage vétérinaire`,
    '',
  ].join('\n')
  const filePath = path.join(os.tmpdir(), `afmps-e2e-${stamp}.csv`)
  fs.writeFileSync(filePath, csv, 'utf8')
  return { filePath, cnk1, cnk2 }
}

test.describe('AFMPS admin imports', { tag: ['@p1', '@pharmacy'] }, () => {
  test.beforeAll(async ({ request }) => {
    const apiBase = (process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291')
      .replace(/\/$/, '')
    const res = await request.post(`${apiBase}/api/v1/auth/login`, {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    test.skip(res.status() !== 200, `admin.demo seed absent (HTTP ${res.status()})`)
  })

  test('D11c nav sidebar → /admin/afmps-imports', async ({ page }) => {
    test.setTimeout(60000)
    const pharmacyOn = process.env.NUXT_PUBLIC_PHARMACY_ENABLED
    test.skip(
      pharmacyOn === 'false' || pharmacyOn === '0',
      'NUXT_PUBLIC_PHARMACY_ENABLED off',
    )

    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin', { waitUntil: 'networkidle' })

    const nav = page.getByTestId('nav-admin-afmps-imports')
    await expect(nav).toBeVisible({ timeout: 15000 })
    await nav.click()

    await expect(page).toHaveURL(/\/admin\/afmps-imports\/?$/, { timeout: 15000 })
    await expect(page.getByTestId('admin-afmps-imports-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('afmps-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-new')).toBeVisible()

    await page.getByTestId('admin-afmps-new').click()
    await expect(page.getByTestId('admin-afmps-upload-card')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-file')).toBeVisible()
  })

  test('D11c triple contrôle UI validate → review → commit', async ({ page }) => {
    test.setTimeout(120000)
    const pharmacyOn = process.env.NUXT_PUBLIC_PHARMACY_ENABLED
    test.skip(
      pharmacyOn === 'false' || pharmacyOn === '0',
      'NUXT_PUBLIC_PHARMACY_ENABLED off',
    )

    // Garde-fou : fixture versionnée présente pour ops / docs.
    expect(fs.existsSync(FIXTURE)).toBeTruthy()

    const { filePath, cnk1 } = writeUniqueMiniCsv()
    test.info().annotations.push({ type: 'afmps-cnk', description: cnk1 })

    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin/afmps-imports', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-afmps-imports-page')).toBeVisible({ timeout: 15000 })

    await page.getByTestId('admin-afmps-new').click()
    await expect(page.getByTestId('admin-afmps-upload-card')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-afmps-file')).toBeVisible()
    await page.getByTestId('admin-afmps-file').setInputFiles(filePath)
    await page.getByTestId('admin-afmps-upload').click()

    await expect(page).toHaveURL(/\/admin\/afmps-imports\/[0-9a-f-]+/i, { timeout: 30000 })
    await expect(page.getByTestId('admin-afmps-import-detail')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-gates')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-job-status')).toHaveAttribute('data-status', 'validated', {
      timeout: 15000,
    })
    await expect(page.getByTestId('admin-afmps-preview')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-collision-filters')).toBeVisible()
    await expect(page.getByText(cnk1)).toBeVisible()

    // Gate 2
    await page.getByTestId('admin-afmps-review-ack-check').check()
    await page.getByTestId('admin-afmps-mark-reviewed').click()
    await expect(page.getByTestId('admin-afmps-commit-form')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-afmps-deactivate')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-job-status')).toHaveAttribute('data-status', 'reviewed')

    // Gate 3 — sans deactivate-missing
    await page.getByTestId('admin-afmps-confirm').fill('IMPORT_AFMPS')
    await page.getByTestId('admin-afmps-commit').click()
    await expect(page.getByTestId('admin-afmps-job-status')).toHaveAttribute('data-status', 'completed', {
      timeout: 30000,
    })
    await expect(page.getByTestId('admin-afmps-commit-form')).toHaveCount(0)

    try {
      fs.unlinkSync(filePath)
    }
    catch {
      /* ignore */
    }
  })
})
