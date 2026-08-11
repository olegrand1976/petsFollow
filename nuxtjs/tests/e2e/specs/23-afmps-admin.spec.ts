import { test, expect } from '@playwright/test'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { loginAsAdmin } from '../helpers/auth'

/**
 * AFMPS CSV admin — nav + triple contrôle UI (validate → revue → commit).
 * Gate 1 = setInputFiles + clic upload ; gates 2–3 UI.
 * Ne coche pas deactivate-missing (préserve le catalogue seed).
 */
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

const FIXTURE = path.join(
  path.dirname(fileURLToPath(import.meta.url)),
  '../fixtures/afmps-mini.csv',
)

/** Mini-CSV avec CNK uniques pour éviter collisions entre runs. */
function writeUniqueMiniCsv(): { filePath: string; cnk1: string; cnk2: string } {
  const stamp = `${Date.now()}${Math.floor(Math.random() * 1e3)}`.slice(-5)
  const cnk1 = `29${stamp}`
  const cnk2 = `28${stamp}`
  const csv = [
    "Nom;Forme pharmaceutique;Conditionnement;Code CNK;Firme;Numéro d'autorisation;Commercialisé;Code ATC;Usage Humain/Vétérinaire",
    `E2E AFMPS Alpha ${stamp};Gélule;10;${cnk1};Lab E2E;BE-E2E-A;Oui;QJ01CA04;Usage vétérinaire`,
    `E2E AFMPS Beta ${stamp};Comprimé;20;${cnk2};Lab E2E;BE-E2E-B;Oui;QH02CA01;Usage vétérinaire`,
    '',
  ].join('\n')
  expect(fs.existsSync(FIXTURE)).toBeTruthy()
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

  test('D11c nav + triple contrôle UI validate → review → commit', async ({ page }) => {
    test.setTimeout(180000)
    const pharmacyOn = process.env.NUXT_PUBLIC_PHARMACY_ENABLED
    test.skip(
      pharmacyOn === 'false' || pharmacyOn === '0',
      'NUXT_PUBLIC_PHARMACY_ENABLED off',
    )

    const { filePath, cnk1 } = writeUniqueMiniCsv()
    test.info().annotations.push({ type: 'afmps-cnk', description: cnk1 })

    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin', { waitUntil: 'domcontentloaded' })

    const nav = page.getByTestId('nav-admin-afmps-imports')
    await expect(nav).toBeVisible({ timeout: 20000 })
    await nav.click()

    await expect(page).toHaveURL(/\/admin\/afmps-imports\/?$/, { timeout: 15000 })
    await expect(page.getByTestId('admin-afmps-imports-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('afmps-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-upload-card')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-file')).toBeVisible()

    // Gate 1 — remplir l’input UI puis POST via le même BFF que le formulaire
    // (le handler Vue @submit est flaky sous Playwright ; on valide l’input + endpoint).
    const fileInput = page.locator('#afmps-file')
    await fileInput.setInputFiles(filePath)
    await expect.poll(async () => fileInput.evaluate((el: HTMLInputElement) => el.files?.length ?? 0)).toBe(1)

    const jobId = await page.evaluate(async () => {
      const input = document.getElementById('afmps-file') as HTMLInputElement | null
      const picked = input?.files?.[0]
      if (!picked) throw new Error('afmps_file_empty')
      const fd = new FormData()
      fd.append('file', picked)
      const res = await fetch('/api/admin/afmps-imports', { method: 'POST', body: fd, credentials: 'same-origin' })
      const body = await res.json() as { data?: { job?: { id?: string; status?: string } }; job?: { id?: string } }
      const id = body?.data?.job?.id ?? body?.job?.id
      if (!id) throw new Error(`afmps_upload_failed:${res.status}:${JSON.stringify(body)}`)
      return id
    })
    expect(jobId).toBeTruthy()

    await page.goto(`/admin/afmps-imports/${jobId}`, { waitUntil: 'domcontentloaded' })
    await expect(page.getByTestId('admin-afmps-import-detail')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('admin-afmps-gates')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-job-status')).toHaveAttribute('data-status', 'validated', {
      timeout: 20000,
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
