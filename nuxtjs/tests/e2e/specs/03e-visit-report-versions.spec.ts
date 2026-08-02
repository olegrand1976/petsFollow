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

/** Toggle history <details> via property — click is flaky under treatments overlay on Cloud Run. */
async function setVisitReportHistoryOpen(page: Page, open: boolean) {
  const history = page.getByTestId('visit-report-history')
  await history.scrollIntoViewIfNeeded().catch(() => undefined)
  await history.evaluate((el, wantOpen) => {
    ;(el as HTMLDetailsElement).open = wantOpen
  }, open)
  await expect
    .poll(async () => history.evaluate((el) => (el as HTMLDetailsElement).open), { timeout: 5000 })
    .toBe(open)
}

/** History CTAs sit under sticky head/footer — scroll into view then native click (force can miss Vue handlers). */
async function clickVisitReportRestore(page: Page, testId: 'visit-report-restore-v0' | 'visit-report-restore-v1') {
  await setVisitReportHistoryOpen(page, true)
  const btn = page.getByTestId(testId)
  await expect(btn).toBeVisible({ timeout: 5000 })
  await btn.evaluate((el) => {
    const scroll = el.closest('.pro-visit-report__scroll') as HTMLElement | null
    if (scroll) {
      const er = el.getBoundingClientRect()
      const sr = scroll.getBoundingClientRect()
      scroll.scrollTop += er.top - sr.top - sr.height / 2 + er.height / 2
    }
    else {
      el.scrollIntoView({ block: 'center', inline: 'nearest' })
    }
    ;(el as HTMLButtonElement).click()
  })
}

/** Rich editor keeps a hidden markdown mirror (visit-report-body) for fill/toHaveValue. */
async function ensureVisitReportEditMode(page: Page) {
  await setVisitReportHistoryOpen(page, false)
  await expect(page.getByTestId('visit-report-editor')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('visit-report-body')).toBeAttached({ timeout: 10000 })
}

async function openConsultationReport(page: Page) {
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
    await setVisitReportHistoryOpen(page, true)
    await expect(page.getByTestId('visit-report-v0')).toBeVisible()
    await expect(
      page.getByTestId('visit-report-restore-v0'),
      'restore v0 requires API PUT transcriptText persistence — rebuild/restart API if missing',
    ).toBeVisible({ timeout: 5000 })
    await transcript.fill('dirty left before restore')
    await clickVisitReportRestore(page, 'visit-report-restore-v0')
    await expect(transcript).toHaveValue(persistedTranscript)

    // Rich prose must not keep raw <script> from hostile corpus
    await setVisitReportHistoryOpen(page, false)
    const prose = page.getByTestId('visit-report-prose')
    await expect(prose).toBeVisible()
    const html = await prose.innerHTML()
    expect(html.toLowerCase()).not.toContain('<script')

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
    let improveLocale = ''
    let resolveImprove: (() => void) | null = null
    const improveGate = new Promise<void>((resolve) => {
      resolveImprove = resolve
    })
    const releaseImproveGate = () => {
      resolveImprove?.()
      resolveImprove = null
    }
    await page.route('**/api/visits/*/report-improve', async (route) => {
      const payload = route.request().postDataJSON() as {
        sourceText?: string
        targetLocale?: string
      } | null
      improveSource = String(payload?.sourceText || '')
      improveLocale = String(payload?.targetLocale || '')
      await improveGate
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
            isReference: false,
          },
        }),
      })
    })

    await expect(page.getByTestId('visit-report-target-locale')).toHaveValue('auto')
    await page.getByTestId('visit-report-howto').locator('summary').click()
    await expect(page.getByTestId('visit-report-howto-image')).toBeVisible()

    const pharmacyEnv = process.env.NUXT_PUBLIC_PHARMACY_ENABLED
    const pharmacyFlagOn = pharmacyEnv !== 'false' && pharmacyEnv !== '0'

    try {
      await page.getByTestId('visit-report-improve').click()
      await expect.poll(() => improveSource, { timeout: 20000 }).toContain(notes)
      await expect.poll(() => improveLocale, { timeout: 5000 }).toBe('auto')
      // Portable: loader visible while IA is gated.
      await expect(page.getByTestId('visit-report-improving-banner')).toBeVisible({ timeout: 10000 })
      // Meaningful only when pharmacy UI can mount — otherwise always 0 (false green).
      if (pharmacyFlagOn) {
        await expect(page.getByTestId('consultation-treatments')).toHaveCount(0)
        await expect(page.getByTestId('consultation-daf-steps')).toHaveCount(0)
      }
    }
    finally {
      releaseImproveGate()
    }

    await expect(page.getByTestId('visit-report-improving-banner')).toHaveCount(0, { timeout: 10000 })
    await expect(page.getByTestId('visit-report-quality')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('visit-report-msg')).toBeVisible({ timeout: 15000 })
    if (pharmacyFlagOn) {
      const treatments = page.getByTestId('consultation-treatments')
      if ((await treatments.count()) === 0) {
        // Flag on but no write ACL — DAF during-flight assert above was still valid (stayed 0).
        test.info().annotations.push({
          type: 'note',
          description: 'pharmacy flag on but treatments panel absent after improve (ACL pharmacy.write?)',
        })
      }
      else {
        await expect(treatments).toBeVisible({ timeout: 10000 })
        await expect(page.getByTestId('consultation-daf-steps')).toBeVisible()
      }
    }

    // Improve switches markdown editor to preview — reopen history only for restore CTAs.
    await setVisitReportHistoryOpen(page, true)
    await expect(page.getByTestId('visit-report-restore-v1')).toBeVisible()
    await ensureVisitReportEditMode(page)
    await body.fill('dirty right before restore v1')
    await clickVisitReportRestore(page, 'visit-report-restore-v1')
    await expect(body).toHaveValue(improved, { timeout: 10000 })
    await ensureVisitReportEditMode(page)
    await expect(body).toHaveValue(improved)

    // Restore v0 → gauche
    await transcript.fill('dirty left before restore v0')
    await clickVisitReportRestore(page, 'visit-report-restore-v0')
    await expect(transcript).toHaveValue(notes, { timeout: 10000 })

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
            isReference: false,
          },
        }),
      })
    })

    let referencePatched: boolean | null = null
    await page.route('**/api/visits/*/report-reference', async (route) => {
      const payload = route.request().postDataJSON() as { isReference?: boolean } | null
      referencePatched = payload?.isReference === true
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
            isReference: true,
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

    await expect(page.getByTestId('visit-report-reference')).toBeVisible({ timeout: 10000 })
    await page.getByTestId('visit-report-reference-check').check()
    await expect.poll(() => referencePatched, { timeout: 10000 }).toBe(true)
    await expect(page.getByTestId('visit-report-reference-check')).toBeChecked()

    await page.unroute('**/api/visits/*/report-improve')
    await page.unroute('**/api/visits/*/report-finalize')
    await page.unroute('**/api/visits/*/report-reference')
  })
})
