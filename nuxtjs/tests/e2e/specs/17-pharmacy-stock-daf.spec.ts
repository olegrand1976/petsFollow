import { test, expect } from '@playwright/test'
import { login } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'

type Envelope = {
  data?: {
    items?: Array<Record<string, unknown>>
    batch?: Record<string, unknown>
    id?: string
    lines?: Array<Record<string, unknown>>
    vamregStatus?: string
    hasAntibiotic?: boolean
    softWarnings?: number
    status?: string
  }
}

async function jsonBody(res: { json: () => Promise<unknown> }): Promise<Envelope> {
  const body = await res.json()
  return (body && typeof body === 'object' ? body : {}) as Envelope
}

/** Date calendaire Europe/Brussels (évite le décalage UTC de toISOString). */
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

async function ensurePharmacy(page: import('@playwright/test').Page) {
  const search = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
  if (search.status() === 404) {
    test.skip(true, 'PHARMACY_ENABLED off')
  }
  expect(search.status()).toBe(200)
  return search
}

test.describe('pharmacy stock + DAF trace', { tag: ['@p0', '@pharmacy'] }, () => {
  test('receipt → finalize → movements?dafId= expose lot + daf links', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    // Non-antibiotique : finalize n'exige pas vamregPayload (seed Vaccin Rage Demo).
    const search = await ensurePharmacy(page)
    const searchEnv = await jsonBody(search)
    const items = searchEnv.data?.items ?? []
    const medId = typeof items[0]?.id === 'string' ? items[0].id : ''
    expect(medId, 'seed CNK Vaccin Rage Demo').toBeTruthy()

    const expiresOn = brusselsYMD(120)
    const lot = `E2E-LOT-${Date.now()}`

    const receipt = await page.request.post('/api/vet/pharmacy/batches', {
      data: { medicationId: medId, lotNumber: lot, expiresOn, qty: 5, unit: 'box' },
    })
    expect(receipt.status(), await receipt.text()).toBe(201)
    const receiptEnv = await jsonBody(receipt)
    expect(receiptEnv.data?.batch?.lotNumber).toBe(lot)

    const blank = await page.request.post('/api/vet/pharmacy/batches', {
      data: { medicationId: medId, lotNumber: '  ', expiresOn, qty: 1 },
    })
    expect(blank.status()).toBe(400)

    const draft = await page.request.post('/api/vet/pharmacy/daf', {
      data: {
        notes: 'e2e',
        items: [{ medicationId: medId, qty: 2, ammNumber: 'BE-E2E-1', unit: 'box' }],
      },
    })
    expect(draft.status(), await draft.text()).toBe(201)
    const dafId = String((await jsonBody(draft)).data?.id ?? '')
    expect(dafId).toBeTruthy()

    // FEFO peut servir le reliquat d'un run précédent (même péremption, plus
    // ancien) : prédire l'allocation réelle plutôt que supposer une base vierge.
    const preview = await page.request.post('/api/vet/pharmacy/daf/preview-fefo', {
      data: { items: [{ medicationId: medId, qty: 2, unit: 'box' }] },
    })
    expect(preview.status(), await preview.text()).toBe(200)
    const previewBody = await preview.json() as { data?: { lines?: Array<{ lotNumber?: string }> } }
    const expectedLots = (previewBody.data?.lines ?? []).map((l) => String(l.lotNumber ?? ''))
    expect(expectedLots.length).toBeGreaterThanOrEqual(1)
    expect(expectedLots.every(Boolean)).toBe(true)

    const finalize = await page.request.post(`/api/vet/pharmacy/daf/${dafId}/finalize`)
    expect(finalize.status(), await finalize.text()).toBe(200)

    const mov = await page.request.get(`/api/vet/pharmacy/movements?dafId=${encodeURIComponent(dafId)}&limit=20`)
    expect(mov.status()).toBe(200)
    const movItems = (await jsonBody(mov)).data?.items ?? []
    const linked = movItems.filter((m) => m.reason === 'daf' && m.dafId === dafId)
    expect(linked.length).toBeGreaterThanOrEqual(1)
    for (const m of linked) {
      expect(m.dafItemId).toBeTruthy()
      expect(expectedLots).toContain(String(m.lotNumber))
    }
    const totalDelta = linked.reduce((sum, m) => sum + Number(m.delta), 0)
    expect(totalDelta).toBe(-2)

    const ret = await page.request.get('/api/vet/pharmacy/movements/retention-stats')
    expect(ret.status()).toBe(200)
    const retData = (await ret.json() as { data?: { retentionYears?: number; total?: number; immutableAppRole?: boolean } }).data
    expect(retData?.retentionYears).toBe(5)
    expect(retData?.immutableAppRole).toBe(true)
    expect(Number(retData?.total ?? 0)).toBeGreaterThanOrEqual(1)

    const refs = await page.request.get('/api/vet/pharmacy/vamreg-refs?kind=target_species')
    expect(refs.status(), await refs.text()).toBe(200)
  })
})

test.describe('pharmacy ops (reorder + inventory + VAMReg)', { tag: ['@p1', '@pharmacy'] }, () => {
  test('reorder threshold → alert → delivery-note receipt', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })
    await ensurePharmacy(page)

    const search = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
    const medId = String((await jsonBody(search)).data?.items?.[0]?.id ?? '')
    expect(medId).toBeTruthy()

    const thr = await page.request.put('/api/vet/pharmacy/reorder-thresholds', {
      data: { medicationId: medId, minQty: 10_000 },
    })
    expect(thr.status(), await thr.text()).toBe(200)

    const alerts = await page.request.get('/api/vet/pharmacy/reorder-alerts')
    expect(alerts.status()).toBe(200)
    const alertsBody = await alerts.json() as { data?: Array<{ medicationId?: string }> }
    const list = Array.isArray(alertsBody.data) ? alertsBody.data : []
    expect(list.some((a) => a.medicationId === medId)).toBeTruthy()

    const bl = await page.request.post('/api/vet/pharmacy/delivery-notes', {
      data: {
        noteNumber: `E2E-BL-${Date.now()}`,
        notify: false,
        items: [{
          medicationId: medId,
          lotNumber: `E2E-BL-LOT-${Date.now()}`,
          expiresOn: brusselsYMD(180),
          qty: 3,
        }],
      },
    })
    expect(bl.status(), await bl.text()).toBe(201)

    // Remise seuil (idempotence seed — pas d’alerte permanente).
    await page.request.put('/api/vet/pharmacy/reorder-thresholds', {
      data: { medicationId: medId, minQty: 0 },
    })
  })

  test('inventory session open → count → close', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })
    await ensurePharmacy(page)

    const search = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
    const medId = String((await jsonBody(search)).data?.items?.[0]?.id ?? '')
    expect(medId).toBeTruthy()

    const lot = `E2E-INV-${Date.now()}`
    const receipt = await page.request.post('/api/vet/pharmacy/batches', {
      data: { medicationId: medId, lotNumber: lot, expiresOn: brusselsYMD(200), qty: 8 },
    })
    expect(receipt.status(), await receipt.text()).toBe(201)
    const batchId = String((await jsonBody(receipt)).data?.batch?.id ?? '')
    expect(batchId).toBeTruthy()

    // Cancel any leftover open session from a previous run.
    const existing = await page.request.get('/api/vet/pharmacy/inventory/sessions')
    if (existing.status() === 200) {
      const existingBody = await existing.json() as { data?: Array<{ id?: string, status?: string }> }
      const list = Array.isArray(existingBody.data) ? existingBody.data : []
      for (const s of list) {
        if (s.status === 'open' && typeof s.id === 'string') {
          await page.request.post(`/api/vet/pharmacy/inventory/sessions/${s.id}/cancel`, { data: {} })
        }
      }
    }

    const start = await page.request.post('/api/vet/pharmacy/inventory/sessions', { data: {} })
    expect(start.status(), await start.text()).toBe(201)
    const sessBody = await start.json() as {
      data?: { id?: string, lines?: Array<{ id?: string, batchId?: string, systemQty?: number }> }
    }
    const sessionId = String(sessBody.data?.id ?? '')
    const lines = sessBody.data?.lines ?? []
    expect(sessionId).toBeTruthy()
    expect(lines.length).toBeGreaterThan(0)

    const target = lines.find((ln) => ln.batchId === batchId) ?? lines[0]
    const lineId = String(target.id ?? '')
    const systemQty = Number(target.systemQty ?? 0)
    expect(lineId).toBeTruthy()

    const patch = await page.request.patch(
      `/api/vet/pharmacy/inventory/sessions/${sessionId}/lines/${lineId}`,
      { data: { countedQty: systemQty - 1 } },
    )
    expect(patch.status(), await patch.text()).toBe(200)

    const counts = lines.map((ln) => {
      const id = String(ln.id ?? '')
      const sys = Number(ln.systemQty ?? 0)
      return { lineId: id, countedQty: id === lineId ? systemQty - 1 : sys }
    })
    const close = await page.request.post(`/api/vet/pharmacy/inventory/sessions/${sessionId}/close`, {
      data: { counts },
    })
    expect(close.status(), await close.text()).toBe(200)
    const closed = await close.json() as { data?: { status?: string } }
    expect(closed.data?.status).toBe('closed')
  })

  test('antibiotic DAF finalize → vamregStatus sent (dry-run) + audits', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })
    await ensurePharmacy(page)

    const search = await page.request.get('/api/vet/pharmacy/medications/search?q=Amoxicilline&limit=5')
    expect(search.status()).toBe(200)
    const searchBody = await search.json() as { data?: { items?: Array<{ id?: string, isAntibiotic?: boolean }> } }
    const items = searchBody.data?.items ?? []
    const ab = items.find((m) => m.isAntibiotic === true) ?? items[0]
    const medId = String(ab?.id ?? '')
    expect(medId, 'seed Amoxicilline Vet Demo').toBeTruthy()

    const lot = `E2E-AB-${Date.now()}`
    const receipt = await page.request.post('/api/vet/pharmacy/batches', {
      data: { medicationId: medId, lotNumber: lot, expiresOn: brusselsYMD(150), qty: 4, unit: 'box' },
    })
    expect(receipt.status(), await receipt.text()).toBe(201)

    const draft = await page.request.post('/api/vet/pharmacy/daf', {
      data: {
        notes: 'e2e-vamreg',
        items: [{
          medicationId: medId,
          qty: 1,
          ammNumber: 'BE-E2E-AB',
          unit: 'box',
          vamregPayload: { species: 'dog', indication: 'infection', durationDays: 5, posology: '1x/day' },
        }],
      },
    })
    expect(draft.status(), await draft.text()).toBe(201)
    const dafId = String((await jsonBody(draft)).data?.id ?? '')
    expect(dafId).toBeTruthy()

    const finalize = await page.request.post(`/api/vet/pharmacy/daf/${dafId}/finalize`)
    expect(finalize.status(), await finalize.text()).toBe(200)
    const finBody = await finalize.json() as {
      data?: {
        hasAntibiotic?: boolean
        vamregStatus?: string
        daf?: { hasAntibiotic?: boolean, vamregStatus?: string }
      }
    }
    const statusDoc = finBody.data?.daf ?? finBody.data
    expect(statusDoc?.hasAntibiotic).toBe(true)
    expect(statusDoc?.vamregStatus).toBe('sent')

    const audits = await page.request.get(`/api/vet/pharmacy/daf/${dafId}/vamreg/audits`)
    expect(audits.status()).toBe(200)
    const auditsBody = await audits.json() as { data?: unknown[] }
    expect(Array.isArray(auditsBody.data) && auditsBody.data.length > 0).toBeTruthy()
  })

  test('prices PUT/GET + MANUAL delivery-note + inventory CSV', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })
    await ensurePharmacy(page)

    const search = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
    const medId = String((await jsonBody(search)).data?.items?.[0]?.id ?? '')
    expect(medId).toBeTruthy()

    const putPrice = await page.request.put(`/api/vet/pharmacy/prices/${medId}`, {
      data: { purchasePriceCents: 800, sellPriceCents: 1900, vatPercent: 21 },
    })
    expect(putPrice.status(), await putPrice.text()).toBe(200)
    const getPrice = await page.request.get(`/api/vet/pharmacy/prices/${medId}`)
    expect(getPrice.status()).toBe(200)
    const priceBody = await getPrice.json() as { data?: { sellPriceCents?: number } }
    expect(priceBody.data?.sellPriceCents).toBe(1900)

    const bl = await page.request.post('/api/vet/pharmacy/delivery-notes', {
      data: {
        noteNumber: '',
        notify: false,
        items: [{
          medicationId: medId,
          lotNumber: `E2E-MAN-${Date.now()}`,
          expiresOn: brusselsYMD(200),
          qty: 2,
        }],
      },
    })
    expect(bl.status(), await bl.text()).toBe(201)
    const blBody = await bl.json() as { data?: { noteNumber?: string } }
    expect(String(blBody.data?.noteNumber ?? '')).toMatch(/^MANUAL-/)

    const existing = await page.request.get('/api/vet/pharmacy/inventory/sessions')
    if (existing.status() === 200) {
      const list = ((await existing.json()) as { data?: Array<{ id?: string, status?: string }> }).data ?? []
      for (const s of list) {
        if (s.status === 'open' && typeof s.id === 'string') {
          await page.request.post(`/api/vet/pharmacy/inventory/sessions/${s.id}/cancel`, { data: {} })
        }
      }
    }
    const start = await page.request.post('/api/vet/pharmacy/inventory/sessions', { data: {} })
    expect(start.status(), await start.text()).toBe(201)
    const sess = await start.json() as {
      data?: { id?: string, lines?: Array<{ id?: string, systemQty?: number }> }
    }
    const sessionId = String(sess.data?.id ?? '')
    const lines = sess.data?.lines ?? []
    const counts = lines.map((ln) => ({
      lineId: String(ln.id ?? ''),
      countedQty: Number(ln.systemQty ?? 0),
    }))
    const close = await page.request.post(`/api/vet/pharmacy/inventory/sessions/${sessionId}/close`, {
      data: { counts },
    })
    expect(close.status(), await close.text()).toBe(200)

    const csv = await page.request.get(`/api/vet/pharmacy/inventory/sessions/${sessionId}/export.csv`)
    expect(csv.status(), await csv.text()).toBe(200)
    const text = await csv.text()
    expect(text).toContain('cnk;name;lot')
  })
})

test.describe('pharmacy stock UI shell', { tag: ['@p1', '@pharmacy'] }, () => {
  test('/stock shows bands + receipt + inventory + dev badge', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    const probe = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=1')
    if (probe.status() === 404) {
      test.skip(true, 'PHARMACY_ENABLED off')
    }

    await page.goto('/stock')
    await expect(page.getByTestId('stock-page')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('stock-page-dev-badge')).toBeVisible()
    await expect(page.getByTestId('stock-bands')).toBeVisible()
    await expect(page.getByTestId('stock-settings')).toBeVisible()
    await expect(page.getByTestId('stock-pricing')).toBeVisible()
    await expect(page.getByTestId('stock-receipt')).toBeVisible()
    await expect(page.getByTestId('stock-inventory')).toBeVisible()
    await expect(page.getByTestId('stock-batches')).toBeVisible()
    await expect(page.getByTestId('stock-movements')).toBeVisible()
    await expect(page.getByTestId('stock-deposits')).toBeVisible()
    await expect(page.getByTestId('stock-receipt-deposit-hint')).toBeVisible()
    await expect(page.getByTestId('stock-receipt-deposit')).toHaveCount(0)

    const depCode = `E2E${Date.now()}${Math.floor(Math.random() * 1e6)}`
    const createDep = await page.request.post('/api/vet/pharmacy/deposits', {
      data: { name: `E2E Depot ${depCode}`, code: depCode, isDefault: false },
    })
    expect(createDep.ok()).toBeTruthy()
    await page.reload()
    await expect(page.getByTestId('stock-receipt-deposit')).toBeVisible()
    await expect(page.getByTestId('stock-deposit-filter')).toBeVisible()
  })
})
