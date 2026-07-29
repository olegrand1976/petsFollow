import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/** Escape corpus (subset of Vitest / Go normalize suite). */
const ESCAPE_NOTES = 'Notes "quotes" \'apos\' «guillemets» back\\slash &amp; <script>alert(1)</script> **bold** été 🐶'

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

async function openConsultationReport(page: Page) {
  await page.goto('/clients', { waitUntil: 'networkidle' })
  if (page.url().includes('/login')) {
    await loginAsVet(page)
    await page.goto('/clients', { waitUntil: 'networkidle' })
  }
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
  const createdBody = await created.json().catch(() => null)
  const visitId = String((createdBody as any)?.data?.id ?? (createdBody as any)?.id ?? '')
  expect(visitId).toBeTruthy()
  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('visit-report-pane-left')).toBeVisible()
  await expect(page.getByTestId('visit-report-pane-right')).toBeVisible()
  return visitId
}

test.describe('CR split versions + boutons', { tag: '@p1' }, () => {
  test.describe.configure({ mode: 'serial' })

  test('panes + boutons idle / dirty / cancel / save + escape round-trip', async ({ page }) => {
    test.setTimeout(120000)
    await loginAsVet(page)
    const visitId = await openConsultationReport(page)

    const transcript = page.getByTestId('visit-report-transcript')
    const body = page.getByTestId('visit-report-body')
    await expect(transcript).toBeEnabled({ timeout: 20000 })
    await expect(body).toBeEnabled({ timeout: 20000 })

    // Idle empty: improve + cancel + save off
    await expect(page.getByTestId('visit-report-improve')).toBeDisabled()
    await expect(page.getByTestId('visit-report-cancel')).toBeDisabled()
    await expect(page.getByTestId('visit-report-save')).toBeDisabled()
    await expect(page.getByTestId('visit-report-dictate')).toBeEnabled()
    await expect(page.getByTestId('visit-report-audio-btn')).toBeEnabled()

    // Body-only → improve enabled (source = droite)
    await body.fill('CR body only for improve gate')
    await expect(page.getByTestId('visit-report-improve')).toBeEnabled()
    await page.getByTestId('visit-report-cancel').click()
    await expect(body).toHaveValue('')

    await transcript.fill(ESCAPE_NOTES)
    await expect(page.getByTestId('visit-report-improve')).toBeEnabled()
    await expect(page.getByTestId('visit-report-cancel')).toBeEnabled()
    await expect(page.getByTestId('visit-report-dirty-hint')).toBeVisible()
    await expect(page.getByTestId('visit-report-cancel')).toContainText(/modifications|Discard|Wijzigingen|Descartar|Tühista|Annulla/i)

    await body.fill('CR édité distinct pour historique v2')
    await expect(page.getByTestId('visit-report-finalize')).toBeEnabled()
    await expect(page.getByTestId('visit-report-save')).toBeEnabled()

    // Cancel restores hydrate (empty)
    await page.getByTestId('visit-report-cancel').click()
    await expect(transcript).toHaveValue('')
    await expect(body).toHaveValue('')
    await expect(page.getByTestId('visit-report-cancel')).toBeDisabled()

    // Save escape corpus
    await transcript.fill(ESCAPE_NOTES)
    await body.fill(`E2E escape CR ${Date.now()}\n**Anamnèse**`)
    const putRes = page.waitForResponse(
      (r) => r.url().includes(`/api/visits/${visitId}/report`) && r.request().method() === 'PUT',
      { timeout: 20000 },
    )
    await page.getByTestId('visit-report-save').click()
    const put = await putRes
    expect(put.status()).toBe(200)
    const putReq = put.request().postDataJSON() as { bodyText?: string, transcriptText?: string } | null
    expect(String(putReq?.transcriptText || '')).toContain('"quotes"')
    expect(String(putReq?.transcriptText || '')).toContain('back\\slash')
    expect(String(putReq?.bodyText || '')).toContain('Anamnèse')
    const putBody = await put.json()
    const saved = (putBody as any)?.data ?? putBody
    // Server must echo transcriptText (PUT persistence) — fail if API stale.
    expect(
      String(saved.transcriptText || ''),
      'API must persist and return transcriptText on PUT /report',
    ).toContain('"quotes"')
    const persistedTranscript = String(saved.transcriptText)
    await expect(page.getByTestId('consultation-cta-done')).toBeVisible({ timeout: 15000 })

    // History restore targets — require server-persisted transcript (PUT transcriptText).
    await page.getByTestId('visit-report-history').locator('summary').click()
    await expect(page.getByTestId('visit-report-v0')).toBeVisible()
    await expect(
      page.getByTestId('visit-report-restore-v0'),
      'restore v0 requires API PUT transcriptText persistence — rebuild/restart API if missing',
    ).toBeVisible({ timeout: 5000 })
    await transcript.fill('dirty left before restore')
    await page.getByTestId('visit-report-restore-v0').click()
    await expect(transcript).toHaveValue(persistedTranscript)

    // Preview must not execute script from corpus if shown
    if (await page.getByTestId('visit-report-edit-tab').isVisible()) {
      await page.getByTestId('visit-report-preview-tab').click()
      const preview = page.getByTestId('visit-report-preview')
      await expect(preview).toBeVisible()
      const html = await preview.innerHTML()
      expect(html.toLowerCase()).not.toContain('<script')
    }

    await page.getByTestId('consultation-cta-done').click()
    await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
  })

  test('improve sourceText + restore v1/v0 + finalize lock', async ({ page }) => {
    test.setTimeout(120000)
    await loginAsVet(page)
    const visitId = await openConsultationReport(page)

    const transcript = page.getByTestId('visit-report-transcript')
    const body = page.getByTestId('visit-report-body')
    await expect(transcript).toBeEnabled({ timeout: 20000 })

    const notes = `E2E notes IA ${Date.now()}`
    const bodyBefore = `E2E body before IA ${Date.now()}`
    const improved = `**Anamnèse / motif :**\n\nIA improved ${Date.now()}`

    await transcript.fill(notes)
    await body.fill(bodyBefore)

    let improveSource = ''
    await page.route('**/api/visits/*/report-improve', async (route) => {
      const payload = route.request().postDataJSON() as { sourceText?: string } | null
      improveSource = String(payload?.sourceText || '')
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: 'e2e-mock-report',
            visitId,
            status: 'draft',
            bodyText: improved,
            transcriptText: notes,
            improvedText: improved,
          },
        }),
      })
    })

    await page.getByTestId('visit-report-improve').click()
    await expect.poll(() => improveSource, { timeout: 20000 }).toContain(notes)
    await expect(page.getByTestId('visit-report-msg')).toBeVisible({ timeout: 15000 })

    // History + restore v1 → droite
    const history = page.getByTestId('visit-report-history')
    await history.locator('summary').click()
    await expect(page.getByTestId('visit-report-restore-v1')).toBeVisible()
    await page.getByTestId('visit-report-edit-tab').click()
    await body.fill('dirty right before restore v1')
    await page.getByTestId('visit-report-restore-v1').click()
    await page.getByTestId('visit-report-edit-tab').click()
    await expect(body).toHaveValue(improved)

    // Restore v0 → gauche
    await transcript.fill('dirty left before restore v0')
    await page.getByTestId('visit-report-restore-v0').click()
    await expect(transcript).toHaveValue(notes)

    await page.route('**/api/visits/*/report-finalize', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: 'e2e-mock-report',
            visitId,
            status: 'final',
            bodyText: improved,
            transcriptText: notes,
            improvedText: improved,
          },
        }),
      })
    })

    await page.getByTestId('visit-report-finalize').click()
    await expect(page.getByTestId('visit-report-status-final')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('visit-report-status-final')).toContainText(/Finaliz|Afgerond|Lõplik|Finalis/i)
    await expect(page.getByTestId('visit-report-footer')).toHaveCount(0)
    await expect(page.getByTestId('visit-report-improve')).toHaveCount(0)
    await expect(page.getByTestId('visit-report-dictate')).toHaveCount(0)

    await page.unroute('**/api/visits/*/report-improve')
    await page.unroute('**/api/visits/*/report-finalize')
  })
})
