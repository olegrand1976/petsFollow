import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

/** /clients peut ouvrir un ProModal (invitations) qui bloque les clics. */
async function dismissProModals(page: Page) {
  // La modale invitations s'ouvre après le fetch des link requests (onMounted),
  // parfois après le premier check : laisser une courte fenêtre d'apparition.
  await page.getByTestId('pro-modal').first()
    .waitFor({ state: 'visible', timeout: 1500 })
    .catch(() => undefined)
  for (let i = 0; i < 3; i++) {
    // Only dismiss invitation-style modals before opening consultation.
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

async function ensureVetOnClients(page: Page) {
  await page.goto('/clients', { waitUntil: 'networkidle' })
  if (page.url().includes('/login')) {
    await loginAsVet(page)
    await page.goto('/clients', { waitUntil: 'networkidle' })
  }
  await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
  await dismissProModals(page)
}

async function openConsultationSetup(page: Page) {
  await ensureVetOnClients(page)

  const search = page.getByPlaceholder(/nom ou email|name or email|naam of e-mail/i)
  await search.fill('Sophie')
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })

  const cta = page.locator('[data-testid^="new-consultation-"]').first()
  await expect(cta).toBeVisible({ timeout: 10000 })
  const petsRes = page.waitForResponse(
    (r) => /\/api\/clients\/[^/]+\/pets\b/.test(r.url()) && r.request().method() === 'GET',
    { timeout: 15000 },
  )
  try {
    await cta.click({ timeout: 5000 })
  }
  catch {
    // Modale invitations apparue entre-temps : backdrop intercepte le clic.
    await dismissProModals(page)
    await cta.click()
  }

  await expect(page.getByTestId('consultation-modal')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('consultation-setup')).toBeVisible()

  const petsOk = await petsRes
  expect(petsOk.status()).toBe(200)
  const petSelect = page.getByTestId('consultation-pet-select')
  await expect(petSelect).toBeEnabled({ timeout: 15000 })
  const options = petSelect.locator('option:not([disabled])')
  await options.first().waitFor({ state: 'attached', timeout: 15000 })
  const value = await options.first().getAttribute('value')
  expect(value).toBeTruthy()
  await petSelect.selectOption(value!)
  await expect(page.getByTestId('consultation-start')).toBeEnabled({ timeout: 10000 })
}

async function startConsultationVisit(page: Page): Promise<{ id: string }> {
  const createRes = page.waitForResponse(
    (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('consultation-start').click()
  const created = await createRes
  expect([200, 201]).toContain(created.status())
  const createdBody = await created.json().catch(() => null)
  const visitPayload = (createdBody as any)?.data ?? createdBody
  expect(visitPayload?.id).toBeTruthy()
  expect(visitPayload?.consultationSession).toBe(true)
  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
  return { id: String(visitPayload.id) }
}

async function saveConsultationReport(page: Page) {
  await expect(page.getByTestId('visit-report-panel')).toBeVisible()
  await expect(page.getByTestId('visit-report-pane-left')).toBeVisible()
  await expect(page.getByTestId('visit-report-pane-right')).toBeVisible()
  const reportBody = page.getByTestId('visit-report-body')
  await expect(reportBody).toBeEnabled({ timeout: 20000 })
  await reportBody.click()
  await reportBody.fill(`E2E consultation CR ${Date.now()}`)
  await expect(page.getByTestId('visit-report-save')).toBeEnabled()
  await page.getByTestId('visit-report-save').click()
  await expect(page.getByTestId('consultation-cta-done')).toBeVisible({ timeout: 15000 })
}

test.describe('nouvelle consultation', { tag: '@p0' }, () => {
  test.describe.configure({ mode: 'serial' })

  test('clients → modal → CR save → CTA Terminer', async ({ page }) => {
    await loginAsVet(page)
    await openConsultationSetup(page)
    await startConsultationVisit(page)
    await saveConsultationReport(page)

    await page.getByTestId('consultation-cta-done').click()
    await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
  })

  test('close sans save → visite cancelled', async ({ page }) => {
    await openConsultationSetup(page)
    const { id: visitId } = await startConsultationVisit(page)

    const cancelRes = page.waitForResponse(
      (r) => {
        if (r.request().method() !== 'PATCH') return false
        if (!r.url().includes(`/api/visits/${visitId}`)) return false
        try {
          const body = r.request().postDataJSON() as { status?: string } | null
          return body?.status === 'cancelled'
        }
        catch {
          return false
        }
      },
      { timeout: 20000 },
    )
    await page.getByTestId('consultation-cancel').click()
    await expect(page.getByTestId('consultation-leave-prompt')).toBeVisible({ timeout: 5000 })
    await page.getByTestId('consultation-leave-discard').click()
    const cancelled = await cancelRes
    expect([200, 204]).toContain(cancelled.status())
    await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })

    const body = await cancelled.json().catch(() => null) as any
    const visit = body?.data ?? body
    if (visit?.status) {
      expect(visit.status).toBe('cancelled')
    }
  })

  test('pendant save CR → close désactivé (pas d’orphelin)', async ({ page }) => {
    await openConsultationSetup(page)
    await startConsultationVisit(page)

    let releasePut: (() => void) | undefined
    const putGate = new Promise<void>((resolve) => {
      releasePut = resolve
    })
    let resolvePutDone: (() => void) | undefined
    const putDone = new Promise<void>((resolve) => {
      resolvePutDone = resolve
    })
    const reportRoute = '**/api/visits/*/report'
    await page.route(reportRoute, async (route) => {
      if (route.request().method() !== 'PUT') {
        await route.continue()
        return
      }
      try {
        await putGate
        await route.continue()
      }
      finally {
        resolvePutDone?.()
      }
    })

    try {
      await expect(page.getByTestId('visit-report-panel')).toBeVisible()
      const reportBody = page.getByTestId('visit-report-body')
      await expect(reportBody).toBeEnabled({ timeout: 20000 })
      await reportBody.click()
      await reportBody.fill(`E2E busy gate ${Date.now()}`)
      await page.getByTestId('visit-report-save').click()

      await expect(page.getByTestId('consultation-cancel')).toBeDisabled({ timeout: 5000 })
      await expect(page.getByTestId('pro-modal-close')).toBeDisabled()
      await expect(page.getByTestId('consultation-modal')).toBeVisible()
    }
    finally {
      // Unblock PUT first — do not unroute until continue() has finished.
      releasePut?.()
    }

    try {
      await putDone
      await expect(page.getByTestId('consultation-cta-done')).toBeVisible({ timeout: 20000 })
      await expect(page.getByTestId('consultation-cancel')).toBeEnabled()
      await page.getByTestId('consultation-cta-done').click()
      await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
    }
    finally {
      await page.unroute(reportRoute).catch(() => undefined)
    }
  })

  test('après CR → CTA facture et DAF (visitId query)', async ({ page }) => {
    await openConsultationSetup(page)
    const { id: visitId } = await startConsultationVisit(page)
    await saveConsultationReport(page)

    if (INVOICING_UI_ENABLED) {
      const invoiceCta = page.getByTestId('consultation-cta-invoice')
      if ((await invoiceCta.count()) === 0) {
        // billitEnabled off → page /invoicing redirige dashboard ; CTA masqué.
      }
      else {
        await expect(invoiceCta).toBeVisible({ timeout: 5000 })
        // Direct goto évite la race soft-nav modal + redirect billit.
        await page.goto(`/invoicing?visitId=${visitId}&mode=direct`, { waitUntil: 'networkidle' })
        if (!page.url().includes('/invoicing')) {
          test.skip(true, 'NUXT_PUBLIC_BILLIT_ENABLED off (redirect)')
        }
        await expect(page.getByTestId('invoicing-page')).toBeVisible({ timeout: 15000 })
        const url = new URL(page.url())
        expect(url.searchParams.get('visitId')).toBe(visitId)
        expect(url.searchParams.get('mode')).toBe('direct')
        await expect(page.getByTestId('invoicing-consultation-context')).toBeVisible({ timeout: 10000 })
      }
    }

    // Nouvelle consultation pour le CTA DAF (session déjà authentifiée).
    await openConsultationSetup(page)
    const { id: visitId2 } = await startConsultationVisit(page)
    await saveConsultationReport(page)

    const dafCta = page.getByTestId('consultation-cta-daf')
    if ((await dafCta.count()) === 0) {
      test.skip(true, 'pharmacy off ou pharmacy.write absent')
    }
    await expect(dafCta).toBeVisible()
    await Promise.all([
      page.waitForURL((url) => url.pathname.includes('/daf/nouveau'), { timeout: 20000 }),
      dafCta.click(),
    ])
    const dafUrl = new URL(page.url())
    expect(dafUrl.searchParams.get('visitId')).toBe(visitId2)
    expect(dafUrl.searchParams.get('petId')).toBeTruthy()
    await expect(page.getByTestId('daf-wizard-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('daf-consultation-context')).toBeVisible()
  })
})
