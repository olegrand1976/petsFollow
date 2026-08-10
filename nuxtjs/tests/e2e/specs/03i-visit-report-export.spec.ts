import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * C2.27 — Exports CR (copy MD / download MD / PDF BFF), mocks PDF blob.
 * Tag @p1.
 */

async function dismissProModals(page: Page) {
  await page.getByTestId('pro-modal').first()
    .waitFor({ state: 'visible', timeout: 1500 })
    .catch(() => undefined)
  for (let i = 0; i < 3; i++) {
    const inviteOnly = page.getByTestId('pro-modal')
    if ((await inviteOnly.count()) === 0) return
    const close = page.getByTestId('pro-modal-close')
    if ((await close.count()) > 0) {
      await close.first().click({ force: true })
    }
    else {
      await page.keyboard.press('Escape')
    }
    await expect(inviteOnly).toHaveCount(0, { timeout: 5000 }).catch(() => undefined)
  }
}

/** Renvoie l'id de la consultation créée, pour la supprimer en fin de test. */
async function openConsultationReport(page: Page): Promise<string> {
  await page.goto('/clients', { waitUntil: 'networkidle' })
  if (page.url().includes('/login')) {
    await loginAsVet(page)
    await page.goto('/clients', { waitUntil: 'networkidle' })
  }
  await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
  await dismissProModals(page)

  const search = page.locator('#client-search')
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
  await expect(petSelect).toBeEnabled({ timeout: 15000 })
  const options = petSelect.locator('option:not([disabled])')
  await options.first().waitFor({ state: 'attached', timeout: 15000 })
  const value = await options.first().getAttribute('value')
  expect(value).toBeTruthy()
  await petSelect.selectOption(value!)

  const createRes = page.waitForResponse(
    (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('consultation-start').click()
  const created = await createRes
  expect([200, 201]).toContain(created.status())
  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })

  const payload = await created.json().catch(() => null)
  return String(payload?.data?.id ?? payload?.id ?? '')
}

test.describe('CR exports PDF/MD (C2.27)', { tag: '@p1' }, () => {
  test('toolbar export : copy MD + PDF mock BFF', async ({ page, context }) => {
    test.setTimeout(120000)
    await loginAsVet(page)
    const visitId = await openConsultationReport(page)

    try {
      const notes = `E2E export notes ${Date.now()}`
      const body = page.getByTestId('visit-report-body')
      await expect(body).toBeAttached({ timeout: 10000 })
      await body.fill(notes)

      const exportBar = page.getByTestId('visit-report-export')
      await expect(exportBar).toBeVisible({ timeout: 10000 })
      await expect(page.getByTestId('visit-report-copy-md')).toBeVisible()
      await expect(page.getByTestId('visit-report-download-md')).toBeVisible()
      await expect(page.getByTestId('visit-report-download-pdf')).toBeVisible()

      await context.grantPermissions(['clipboard-read', 'clipboard-write'])
      await page.getByTestId('visit-report-copy-md').click()
      await expect(page.getByTestId('visit-report-msg')).toContainText(/copié|copied|gekopieerd|copiado|kopeeritud|copiato/i, {
        timeout: 10000,
      })

      // Minimal PDF magic bytes — enough for openVisitReportPdfBlob to treat as PDF.
      const pdfBytes = Buffer.from('%PDF-1.4\n1 0 obj<<>>endobj\ntrailer<<>>\n%%EOF\n')
      await page.route(/\/api\/visits\/[^/]+\/report-pdf(\?.*)?$/, async (route) => {
        await route.fulfill({
          status: 200,
          contentType: 'application/pdf',
          body: pdfBytes,
        })
      })

      // Persist PUT may fire before PDF; let it hit the real API.
      await page.getByTestId('visit-report-download-pdf').click()
      await expect(page.getByTestId('visit-report-msg')).toContainText(/PDF/i, { timeout: 15000 })
    }
    finally {
      // Le CR est persisté : le job de rétention n'annule que les walk-ins sans
      // CR, donc sans ce nettoyage chaque exécution laisse une consultation sur
      // le client de démo.
      if (visitId) await page.request.delete(`/api/visits/${visitId}`).catch(() => undefined)
    }
  })
})
