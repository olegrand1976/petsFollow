import { test, expect } from '@playwright/test'
import { login } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'

type Envelope = {
  data?: {
    items?: Array<Record<string, unknown>>
    batch?: Record<string, unknown>
    id?: string
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

test.describe('pharmacy stock + DAF trace', { tag: ['@p0', '@pharmacy'] }, () => {
  test('receipt → finalize → movements?dafId= expose lot + daf links', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    // Non-antibiotique : finalize n'exige pas vamregPayload (seed Vaccin Rage Demo).
    const search = await page.request.get('/api/vet/pharmacy/medications/search?q=vaccin&limit=5')
    if (search.status() === 404) {
      test.skip(true, 'PHARMACY_ENABLED off')
    }
    expect(search.status()).toBe(200)
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
  })
})
