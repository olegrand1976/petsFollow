import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/** C2.17 — historique /consultations (walk-in date DESC + filtres + ouvrir CR). */

async function dismissProModals(page: Page) {
  await page.getByTestId('pro-modal').first()
    .waitFor({ state: 'visible', timeout: 1500 })
    .catch(() => undefined)
  for (let i = 0; i < 3; i++) {
    const modal = page.getByTestId('pro-modal')
    if ((await modal.count()) === 0) return
    const close = page.getByTestId('pro-modal-close')
    if ((await close.count()) > 0) {
      await close.first().click({ force: true })
    }
    else {
      await page.keyboard.press('Escape')
    }
    await expect(modal).toHaveCount(0, { timeout: 5000 }).catch(() => undefined)
  }
}

async function createWalkInWithReport(page: Page): Promise<string> {
  await page.goto('/clients', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
  await dismissProModals(page)

  const search = page.getByPlaceholder(/nom ou email|name or email|naam of e-mail/i)
  await search.fill('Sophie')
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })

  const cta = page.locator('[data-testid^="new-consultation-"]').first()
  await expect(cta).toBeVisible({ timeout: 10000 })
  try {
    await cta.click({ timeout: 5000 })
  }
  catch {
    await dismissProModals(page)
    await cta.click()
  }

  await expect(page.getByTestId('consultation-modal')).toBeVisible({ timeout: 10000 })
  const petSelect = page.getByTestId('consultation-pet-select')
  await expect(petSelect).toBeEnabled({ timeout: 10000 })
  const options = petSelect.locator('option:not([disabled])')
  await options.first().waitFor({ state: 'attached', timeout: 10000 })
  const value = await options.first().getAttribute('value')
  if (value) await petSelect.selectOption(value)

  const createRes = page.waitForResponse(
    (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('consultation-start').click()
  const created = await createRes
  expect([200, 201]).toContain(created.status())
  const body = await created.json().catch(() => null) as any
  const visitId = String((body?.data ?? body)?.id || '')
  expect(visitId).toBeTruthy()

  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
  const reportBody = page.getByTestId('visit-report-body')
  await reportBody.fill(`E2E history CR ${Date.now()}`)
  await page.getByTestId('visit-report-save').click()
  await expect(page.getByTestId('consultation-cta-done')).toBeVisible({ timeout: 15000 })
  await page.getByTestId('consultation-cta-done').click()
  await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
  return visitId
}

test.describe('historique consultations', { tag: '@p1' }, () => {
  test('liste /consultations + filtre + ouvrir CR', async ({ page }) => {
    await loginAsVet(page)
    const visitId = await createWalkInWithReport(page)

    const listRes = page.waitForResponse(
      (r) => r.url().includes('/api/vet/consultations') && r.request().method() === 'GET',
      { timeout: 20000 },
    )
    await page.goto('/consultations', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
    const listed = await listRes
    expect(listed.status()).toBe(200)

    await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('consultations-search')).toBeVisible()
    await expect(page.getByTestId('consultations-status-filter')).toBeVisible()

    await page.getByTestId('consultations-search').fill('Sophie')
    await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 10000 })

    await page.getByTestId(`consultation-open-cr-${visitId}`).click()
    await expect(page.getByTestId('consultation-history-report-modal')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('visit-report-panel')).toBeVisible()
    await expect(page.getByTestId('visit-report-body')).toContainText(/E2E history CR/)
  })
})
