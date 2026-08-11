/**
 * Flux croisé P1 : consultation (UI) → CTA consignes → Rx (API) → dispense/DAF/stock (UI) → facture.
 * La création Rx reste via API (combobox catalogue flaky) ; le CTA hub + dispense + finalize sont UI.
 * Prérequis : api-dev + nuxtjs-dev + seed (PHARMACY + PRESCRIPTIONS + BILLIT).
 * Patient : Spirit (cheval) — DAF espèce applicable.
 */
import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

type Envelope = {
  data?: {
    items?: Array<Record<string, unknown>>
    batch?: Record<string, unknown>
    id?: string
    visitId?: string
    petId?: string
    dafId?: string
    status?: string
    lines?: Array<Record<string, unknown>>
    quantity?: number
    unitPriceExclCents?: number
    description?: string
  }
}

async function jsonBody(res: { json: () => Promise<unknown> }): Promise<Envelope> {
  const body = await res.json()
  return (body && typeof body === 'object' ? body : {}) as Envelope
}

function brusselsYMD(daysFromToday: number): string {
  const parts = new Intl.DateTimeFormat('en-CA', {
    timeZone: 'Europe/Brussels',
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
  }).formatToParts(new Date())
  const y = Number(parts.find((p) => p.type === 'year')?.value)
  const m = Number(parts.find((p) => p.type === 'month')?.value)
  const d = Number(parts.find((p) => p.type === 'day')?.value)
  const utcMidnight = new Date(Date.UTC(y, m - 1, d + daysFromToday))
  const yy = utcMidnight.getUTCFullYear()
  const mm = String(utcMidnight.getUTCMonth() + 1).padStart(2, '0')
  const dd = String(utcMidnight.getUTCDate()).padStart(2, '0')
  return `${yy}-${mm}-${dd}`
}

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

async function fillVisitReportBody(page: Page, text: string) {
  await expect(page.getByTestId('visit-report-editor')).toBeVisible({ timeout: 20000 })
  const body = page.getByTestId('visit-report-body')
  await expect(body).toBeVisible({ timeout: 10000 })
  const editable = body.locator('[contenteditable="true"]').or(body)
  await editable.first().click()
  await editable.first().fill(text)
}

async function confirmGoToNextSteps(page: Page) {
  await expect(page.getByTestId('consultation-next-prompt')).toBeVisible({ timeout: 15000 })
  await page.getByTestId('consultation-next-continue').click()
  await expect(page.getByTestId('consultation-next-steps')).toBeVisible({ timeout: 10000 })
}

test.describe('flux consult → consignes → stock → facture', {
  tag: ['@p1', '@pharmacy', '@invoicing'],
}, () => {
  test('hub CR → CTA consignes → dispense DAF → finalize → prefill facture', async ({ page }) => {
    test.setTimeout(180_000)

    await loginAsVet(page)

    // Flags en tête — avant toute mutation stock / visite.
    const rxProbe = await page.request.get('/api/vet/prescriptions?limit=1')
    if (rxProbe.status() === 404) {
      test.skip(true, 'PRESCRIPTIONS_ENABLED off')
    }
    expect(rxProbe.status()).toBe(200)

    const phProbe = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
    if (phProbe.status() === 404) {
      test.skip(true, 'PHARMACY_ENABLED off')
    }
    expect(phProbe.status()).toBe(200)

    if (!INVOICING_UI_ENABLED) {
      test.skip(true, 'INVOICING_UI_ENABLED off')
    }
    const billitProbe = await page.request.get('/api/invoicing/connection')
    if (billitProbe.status() === 404) {
      test.skip(true, 'BILLIT_ENABLED off')
    }

    const medItems = (await jsonBody(phProbe)).data?.items ?? []
    const med = medItems.find((m) => String(m.name || '').toLowerCase().includes('rage'))
      ?? medItems[0]
    const medId = typeof med?.id === 'string' ? med.id : ''
    expect(medId, 'seed Vaccin Rage Demo').toBeTruthy()

    const clientsRes = await page.request.get('/api/clients')
    const clients = ((await clientsRes.json()) as {
      data?: Array<{ userId: string, email?: string }>
    }).data ?? []
    const demoClient = clients.find((c) => c.email === 'client.demo@petsfollow.test')
    expect(demoClient, 'client.demo seed').toBeTruthy()

    const petsRes = await page.request.get(`/api/clients/${demoClient!.userId}/pets`)
    const pets = ((await petsRes.json()) as {
      data?: Array<{ id: string, name?: string, species?: string }>
    }).data ?? []
    const spirit = pets.find((p) => /spirit/i.test(String(p.name || '')) || p.species === 'horse')
    expect(spirit?.id, 'Spirit (cheval) seed pour DAF').toBeTruthy()
    const petId = spirit!.id

    const lot = `E2E-FLOW-${Date.now()}`
    const receipt = await page.request.post('/api/vet/pharmacy/batches', {
      data: {
        medicationId: medId,
        lotNumber: lot,
        expiresOn: brusselsYMD(120),
        qty: 5,
        unit: 'box',
      },
    })
    expect(receipt.status(), await receipt.text()).toBe(201)

    // 1) Consultation walk-in UI sur Spirit
    await page.goto('/clients', { waitUntil: 'domcontentloaded' })
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
    await petSelect.selectOption(petId)

    const createRes = page.waitForResponse(
      (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('consultation-start').click()
    const created = await createRes
    expect([200, 201]).toContain(created.status())
    const createdBody = await created.json().catch(() => null) as { data?: { id?: string }, id?: string } | null
    const visitId = String(createdBody?.data?.id || createdBody?.id || '')
    expect(visitId).toBeTruthy()

    await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
    await fillVisitReportBody(page, `E2E flux consignes+stock+facture ${Date.now()}`)
    await expect(page.getByTestId('visit-report-save')).toBeEnabled()
    await page.getByTestId('visit-report-save').click()
    await confirmGoToNextSteps(page)

    const rxCta = page.getByTestId('consultation-cta-prescription')
    if ((await rxCta.count()) === 0) {
      test.skip(true, 'CTA consignes absent (flag / ACL)')
    }
    await Promise.all([
      page.waitForURL((url) => url.pathname.includes('/prescriptions/nouveau'), { timeout: 20000 }),
      rxCta.click(),
    ])
    const rxUrl = new URL(page.url())
    expect(rxUrl.searchParams.get('visitId')).toBe(visitId)
    expect(rxUrl.searchParams.get('petId')).toBe(petId)
    await expect(page.getByTestId('prescriptions-wizard-page')).toBeVisible({ timeout: 15000 })

    // 2) Consigne API (médicament catalogue) — CTA+query déjà validés ci-dessus
    const rxCreate = await page.request.post('/api/vet/prescriptions', {
      data: {
        petId,
        visitId,
        careAdvice: 'Repos 48h — E2E flux',
        medications: [{
          name: String(med?.name || 'Vaccin Rage Demo'),
          dosage: '1',
          form: 'box',
          quantity: '2',
          posology: 'SID',
          cnk: med?.cnk,
          ref_medication_id: medId,
        }],
      },
    })
    expect(rxCreate.status(), await rxCreate.text()).toBe(201)
    const rxId = String((await jsonBody(rxCreate)).data?.id || '')
    expect(rxId).toBeTruthy()

    // 3) Dispense UI → DAF lié visite (attendre fin load visites pour éviter clear visitId)
    await page.goto(`/prescriptions/${rxId}`, { waitUntil: 'domcontentloaded' })
    await expect(page.getByTestId('prescriptions-detail-page')).toBeVisible({ timeout: 15000 })
    const dispense = page.getByTestId('prescriptions-dispense-daf')
    await expect(dispense).toBeVisible({ timeout: 10000 })
    await expect(dispense).toBeEnabled({ timeout: 15000 })
    await Promise.all([
      page.waitForURL((url) => url.pathname.includes('/daf/nouveau'), { timeout: 20000 }),
      dispense.click(),
    ])
    const dafUrl = new URL(page.url())
    expect(dafUrl.searchParams.get('visitId')).toBe(visitId)
    expect(dafUrl.searchParams.get('dafId')).toBeTruthy()
    const dafId = String(dafUrl.searchParams.get('dafId'))
    await expect(page.getByTestId('daf-wizard-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('daf-from-consultation-banner')).toBeVisible()
    await expect(page.getByTestId('daf-line-0')).toBeVisible()

    // 4) Finalize FEFO (stock)
    await page.getByTestId('daf-finalize-btn').click()
    await expect(page.getByTestId('daf-finalize-confirm')).toBeVisible({ timeout: 10000 })
    const finRes = page.waitForResponse(
      (r) => r.url().includes(`/api/vet/pharmacy/daf/${dafId}/finalize`) && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('daf-finalize-ok').click()
    const finalized = await finRes
    expect(finalized.status(), await finalized.text()).toBe(200)

    const mov = await page.request.get(`/api/vet/pharmacy/movements?dafId=${encodeURIComponent(dafId)}&limit=20`)
    expect(mov.status()).toBe(200)
    const movItems = (await jsonBody(mov)).data?.items ?? []
    const linked = movItems.filter((m) => m.reason === 'daf' && m.dafId === dafId)
    expect(linked.length).toBeGreaterThanOrEqual(1)
    const totalDelta = linked.reduce((sum, m) => sum + Number(m.delta), 0)
    expect(totalDelta).toBe(-2)

    // 5) Facturation préremplie (sans dafId dans l'URL = résolution via visitId)
    await page.goto(`/invoicing?visitId=${visitId}&mode=fromDaf`, {
      waitUntil: 'domcontentloaded',
    })
    await expect(page.getByTestId('invoicing-page')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('invoicing-consultation-context')).toBeVisible({ timeout: 15000 })

    const prefill = await page.request.get(
      `/api/invoicing/prefill?visitId=${encodeURIComponent(visitId)}`,
    )
    expect(prefill.status()).toBe(200)
    const prefillEnv = await prefill.json() as {
      data?: { lines?: Array<{ quantity?: number, description?: string }>, dafId?: string }
    }
    const prefillData = prefillEnv.data ?? {}
    const prefillLines = prefillData.lines ?? []
    expect(prefillData.dafId, 'DAF résolu via visitId').toBe(dafId)
    expect(prefillLines.length, 'au moins la ligne médicament DAF').toBeGreaterThanOrEqual(1)
    const medLine = prefillLines.find((l) => Number(l.quantity) === 2)
      ?? prefillLines[prefillLines.length - 1]
    expect(Number(medLine?.quantity)).toBe(2)

    await expect(page.getByTestId('invoicing-lines')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('invoicing-line-0')).toBeVisible()

    await page.request.delete(`/api/visits/${visitId}`).catch(() => undefined)
  })
})
