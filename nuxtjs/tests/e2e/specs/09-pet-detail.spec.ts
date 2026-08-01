import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

const API = process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291'

/** Invitation / overlay ProModal can intercept the CTA click on Cloud Run. */
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

async function apiLogin(email: string, password: string): Promise<string> {
  const res = await fetch(`${API}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) throw new Error(`login ${email} ${res.status}`)
  const body = await res.json()
  return body.data.accessToken as string
}

async function seedHeartRateComment(petId: string): Promise<string> {
  const token = await apiLogin('client.demo@petsfollow.test', 'ClientDemo123!')
  const headers = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  }
  const start = await fetch(`${API}/api/v1/pets/${petId}/heartrate/sessions`, {
    method: 'POST',
    headers,
  })
  if (!start.ok) throw new Error(`start hr ${start.status}`)
  const sessId = (await start.json()).data.id as string
  const complete = await fetch(`${API}/api/v1/heartrate/sessions/${sessId}`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ tapCount: 55 }),
  })
  if (!complete.ok) throw new Error(`complete hr ${complete.status}`)
  const comment = `e2e comment ${Date.now()}`
  const validate = await fetch(`${API}/api/v1/heartrate/sessions/${sessId}/validate`, {
    method: 'POST',
    headers,
    body: JSON.stringify({ comment }),
  })
  if (!validate.ok) throw new Error(`validate hr ${validate.status}`)
  return comment
}

async function demoClientAndPet(): Promise<{ clientId: string; petId: string }> {
  const vetTok = await apiLogin('vet.demo@petsfollow.test', 'VetDemo123!')
  const clientsRes = await fetch(`${API}/api/v1/clients`, {
    headers: { Authorization: `Bearer ${vetTok}` },
  })
  if (!clientsRes.ok) throw new Error(`clients ${clientsRes.status}`)
  const clients = (await clientsRes.json()).data as Array<{ userId: string; email?: string; fullName?: string }>
  const client = clients.find((c) => c.email === 'client.demo@petsfollow.test')
    ?? clients.find((c) => /Sophie/i.test(c.fullName || ''))
  if (!client?.userId) throw new Error('demo client not found')

  const petsRes = await fetch(`${API}/api/v1/clients/${client.userId}/pets`, {
    headers: { Authorization: `Bearer ${vetTok}` },
  })
  if (!petsRes.ok) throw new Error(`client pets ${petsRes.status}`)
  const pets = (await petsRes.json()).data as Array<{ id: string; name?: string; paymentStatus?: string }>
  // Prefer Rex (seed BP + labs) when active, else any active pet.
  const pet =
    pets.find((p) => p.paymentStatus === 'active' && /Rex/i.test(p.name || ''))
    ?? pets.find((p) => p.paymentStatus === 'active')
    ?? pets[0]
  if (!pet?.id) throw new Error('demo pet not found')
  return { clientId: client.userId, petId: pet.id }
}

test('pet detail — CR IA ouvert depuis Soins & RDV', { tag: '@p0' }, async ({ page }) => {
  test.setTimeout(90000)
  const { clientId, petId } = await demoClientAndPet()
  const vetTok = await apiLogin('vet.demo@petsfollow.test', 'VetDemo123!')
  // Unique slot: +3h + jitter — avoids slot_taken 409 on Playwright retry / parallel runs.
  const when = new Date(
    Date.now() + 3 * 60 * 60 * 1000 + Math.floor(Math.random() * 50) * 60_000,
  ).toISOString()
  const create = await fetch(`${API}/api/v1/pets/${petId}/visits`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${vetTok}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      notes: `e2e cr ${Date.now()}`,
      confirmDirect: true,
      scheduledAt: when,
    }),
  })
  // 409 slot_taken: a prior attempt (or seed) already holds a nearby slot — still open CR.
  if (!create.ok && create.status !== 409) {
    throw new Error(`create visit ${create.status}`)
  }

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=care`)
  await expect(page.getByTestId('pet-tab-care')).toBeVisible({ timeout: 15000 })
  const openReport = page.locator('[data-testid^="pet-visit-report-open"]').first()
  await expect(openReport).toBeVisible({ timeout: 30000 })
  await openReport.click()
  await expect(page.getByTestId('visit-report-panel')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('visit-report-body')).toBeVisible()
  await expect(page.getByTestId('visit-report-improve')).toBeVisible()
  await expect(page.getByTestId('visit-report-audio')).toBeAttached()
})

test('pet detail — CR visualisable depuis l’historique', { tag: '@p0' }, async ({ page }) => {
  test.setTimeout(90000)
  const { clientId, petId } = await demoClientAndPet()
  const vetTok = await apiLogin('vet.demo@petsfollow.test', 'VetDemo123!')
  const probe = `e2e history cr ${Date.now()}`
  const headers = {
    Authorization: `Bearer ${vetTok}`,
    'Content-Type': 'application/json',
  }

  let visitId = ''
  for (let attempt = 0; attempt < 4 && !visitId; attempt++) {
    const when = new Date(
      Date.now() + (5 + attempt) * 60 * 60 * 1000 + Math.floor(Math.random() * 50) * 60_000,
    ).toISOString()
    const create = await fetch(`${API}/api/v1/pets/${petId}/visits`, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        notes: '',
        confirmDirect: true,
        scheduledAt: when,
      }),
    })
    if (create.ok) {
      visitId = (await create.json()).data.id as string
      break
    }
    // 409 slot_taken: retry with another slot (parallel / Playwright retry).
    if (create.status !== 409) {
      throw new Error(`create visit ${create.status}`)
    }
  }
  if (!visitId) {
    throw new Error('create visit: all slot retries failed (409)')
  }

  const putReport = await fetch(`${API}/api/v1/visits/${visitId}/report`, {
    method: 'PUT',
    headers,
    body: JSON.stringify({ bodyText: probe }),
  })
  if (!putReport.ok) {
    throw new Error(`put report ${putReport.status}`)
  }
  const done = await fetch(`${API}/api/v1/visits/${visitId}`, {
    method: 'PATCH',
    headers,
    body: JSON.stringify({ status: 'done' }),
  })
  if (!done.ok) {
    throw new Error(`mark done ${done.status}`)
  }

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=overview`)
  await expect(page.getByTestId('pet-timeline-card')).toBeVisible({ timeout: 15000 })
  const historyTile = page.getByTestId('pet-history-visit-report').first()
  await expect(historyTile).toBeVisible({ timeout: 30000 })
  await expect(historyTile).toContainText(probe)
  await historyTile.click()
  await expect(page.getByTestId('visit-report-panel')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('visit-report-body')).toBeVisible()
})

test('pet detail — CTA nouvelle consultation', { tag: '@p1' }, async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()
  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}`, { waitUntil: 'networkidle' })
  await expect(page.getByTestId('pet-detail-page')).toBeVisible({ timeout: 15000 })
  await dismissProModals(page)
  const cta = page.getByTestId('pet-new-consultation')
  await expect(cta).toBeVisible({ timeout: 15000 })
  await expect(cta).toBeEnabled()
  try {
    await cta.click({ timeout: 5000 })
  }
  catch {
    await dismissProModals(page)
    await cta.click()
  }
  await expect(page.getByTestId('consultation-modal')).toBeVisible({ timeout: 20000 })
  const petSelect = page.getByTestId('consultation-pet-select')
  await expect(petSelect).toBeEnabled({ timeout: 15000 })
  await expect(petSelect).toHaveValue(petId)
})

test('pet detail — panneau dispenses DAF (C7.17)', { tag: '@p1' }, async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()

  await page.route(`**/api/pets/${petId}/daf-dispenses**`, async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          items: [{
            dafId: 'daf-dispense-mock',
            displayNumber: 'DAF-E2E-1',
            finalizedAt: new Date().toISOString(),
            items: [{ medicationName: 'Vaccin Rage', qty: 1, lotNumber: 'LOT-1', ammNumber: 'BE-DEMO-RAGE-1' }],
          }],
        },
      }),
    })
  })

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}`, { waitUntil: 'networkidle' })
  await expect(page.getByTestId('pet-detail-page')).toBeVisible({ timeout: 15000 })
  await dismissProModals(page)

  const panel = page.getByTestId('pet-daf-dispenses')
  if ((await panel.count()) === 0) {
    test.skip(true, 'pharmacy off ou pharmacy.read absent')
  }
  await expect(panel).toBeVisible({ timeout: 10000 })
  await expect(panel.getByText('DAF-E2E-1')).toBeVisible({ timeout: 10000 })
  await expect(panel.getByText(/Vaccin Rage/)).toBeVisible()
  await page.unroute(`**/api/pets/${petId}/daf-dispenses**`)
})

test('pet detail — overview graphes + historique par jour', { tag: '@p0' }, async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()
  await seedHeartRateComment(petId)

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=overview`)
  await expect(page.getByTestId('pet-detail-page')).toBeVisible()
  await expect(page.getByTestId('pet-tab-overview')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('pet-overview-charts')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('pet-timeline-card')).toBeVisible()
  await expect(page.getByTestId('pet-history-day').first()).toBeVisible()
})

test('pet detail — chart filtres, shares, commentaire HR', { tag: '@p0' }, async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()
  const comment = await seedHeartRateComment(petId)

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=vitals`)
  await expect(page.getByTestId('pet-detail-page')).toBeVisible()
  const vitals = page.getByTestId('pet-tab-vitals')
  await expect(vitals).toBeVisible({ timeout: 15000 })
  await expect(vitals.getByTestId('pet-chart-range-3m')).toBeVisible()
  await vitals.getByTestId('pet-chart-range-6m').click()
  await vitals.getByTestId('pet-filter-all').click()
  await expect(vitals.getByTestId('pet-reading-comment').filter({ hasText: comment })).toBeVisible({
    timeout: 15000,
  })

  await page.goto(`/clients/${clientId}/pets/${petId}?tab=sharing`)
  await expect(page.getByTestId('pet-shares-card')).toBeVisible({ timeout: 15000 })
})

test('pet detail — suivi poids chart + tableau', async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()
  const clientTok = await apiLogin('client.demo@petsfollow.test', 'ClientDemo123!')
  const kg = 31.4
  const comment = `e2e weight ${Date.now()}`
  const create = await fetch(`${API}/api/v1/pets/${petId}/weights`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${clientTok}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ weightKg: kg, comment }),
  })
  if (!create.ok) throw new Error(`create weight ${create.status}`)

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=vitals`)
  await expect(page.getByTestId('pet-detail-page')).toBeVisible()
  await expect(page.getByTestId('pet-weight-table-card')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('pet-weight-comment').filter({ hasText: comment })).toBeVisible({
    timeout: 15000,
  })
  await expect(page.getByText(`${kg} kg`).first()).toBeVisible()
})

test('pet detail — tension + panel labo @p1', { tag: '@p1' }, async ({ page }) => {
  test.setTimeout(90000)
  const { clientId, petId } = await demoClientAndPet()
  const clientTok = await apiLogin('client.demo@petsfollow.test', 'ClientDemo123!')
  const vetTok = await apiLogin('vet.demo@petsfollow.test', 'VetDemo123!')
  const stamp = Date.now()

  const bpRes = await fetch(`${API}/api/v1/pets/${petId}/blood-pressure`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${clientTok}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      systolicMmHg: 142,
      diastolicMmHg: 91,
      method: 'doppler',
      comment: `e2e bp ${stamp}`,
    }),
  })
  if (!bpRes.ok) throw new Error(`create bp ${bpRes.status} ${await bpRes.text()}`)

  const labRes = await fetch(`${API}/api/v1/pets/${petId}/lab-panels`, {
    method: 'POST',
    headers: {
      Authorization: `Bearer ${vetTok}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({
      labName: `E2E Lab ${stamp}`,
      notes: 'e2e panel',
      results: [
        { analyteCode: 'crea', valueNum: 2.2, unit: 'mg/dL', refLow: 0.5, refHigh: 1.5 },
        { analyteCode: 'alat', valueNum: 40, unit: 'U/L', refLow: 10, refHigh: 100 },
      ],
    }),
  })
  if (!labRes.ok) throw new Error(`create lab ${labRes.status} ${await labRes.text()}`)

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=vitals`, { waitUntil: 'networkidle' })
  await expect(page.getByTestId('pet-detail-page')).toBeVisible({ timeout: 15000 })
  // Deep-link ?tab= can race with hydration — force vitals tab like a user click.
  await page.getByTestId('section-tab-vitals').click()
  const vitals = page.getByTestId('pet-tab-vitals')
  await expect(vitals).toBeVisible({ timeout: 15000 })
  await expect(vitals.getByTestId('pet-bp-table-card')).toBeVisible({ timeout: 15000 })
  await expect(vitals.getByTestId('pet-bp-table-card').getByText('142/91')).toBeVisible()
  await expect(vitals.getByTestId('pet-labs-card')).toBeVisible()
  await expect(vitals.getByTestId('pet-labs-card').getByText(`E2E Lab ${stamp}`)).toBeVisible()
  await vitals.getByTestId('pet-lab-panel-row').filter({ hasText: `E2E Lab ${stamp}` }).click()
  await expect(page.getByTestId('pet-lab-detail')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('pet-lab-detail').getByText(/crea|CREA|Créatinine/i)).toBeVisible()

  await page.getByTestId('pet-lab-edit').click()
  await expect(page.getByTestId('pet-lab-edit-form')).toBeVisible()
  const creaValue = page.getByTestId('pet-lab-edit-value').first()
  await creaValue.fill('2.8')
  await page.getByTestId('pet-lab-save-results').click()
  await expect(page.getByTestId('pet-lab-edit-form')).toBeHidden({ timeout: 10000 })
  await expect(page.getByTestId('pet-lab-detail').getByText('2.8')).toBeVisible()
})

test('heartrate — durées cabinet exposées au client + BPM sur 15s', async () => {
  test.setTimeout(90000)
  const password = 'TestPass123!'
  const stamp = Date.now()
  const vetEmail = `hr-dur-vet-${stamp}@petsfollow.test`
  const clientEmail = `hr-dur-client-${stamp}@petsfollow.test`

  const reg = await fetch(`${API}/api/v1/auth/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({
      email: vetEmail,
      password,
      fullName: 'Dr HR Dur',
      practiceName: 'Cabinet HR Dur',
      consent: true,
    }),
  })
  if (!reg.ok) throw new Error(`register vet ${reg.status}`)
  const confirmPath = (await reg.json()).data?.confirmPath as string | undefined
  // confirmPath only when DEV_SEED_ENABLED=true (local/CI) — staging production-like omits it.
  test.skip(!confirmPath, 'confirmPath masqué hors DEV_SEED_ENABLED')
  const confirmToken = confirmPath!.replace('/confirm-email?token=', '')
  const confirm = await fetch(`${API}/api/v1/auth/confirm-email`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ token: confirmToken }),
  })
  if (!confirm.ok) throw new Error(`confirm vet ${confirm.status}`)
  const vetTok = (await confirm.json()).data.accessToken as string
  const vetHeaders = {
    Authorization: `Bearer ${vetTok}`,
    'Content-Type': 'application/json',
  }

  const putRes = await fetch(`${API}/api/v1/vet/profile`, {
    method: 'PUT',
    headers: vetHeaders,
    body: JSON.stringify({
      vetFullName: 'Dr HR Dur',
      practiceName: 'Cabinet HR Dur',
      contactEmail: vetEmail,
      phone: '+32123456789',
      addressLine1: 'Rue Test 1',
      city: 'Bruxelles',
      postalCode: '1000',
      heartrateDurationsSec: [15, 30],
    }),
  })
  if (!putRes.ok) throw new Error(`put profile ${putRes.status}`)

  const createClient = await fetch(`${API}/api/v1/vet/clients`, {
    method: 'POST',
    headers: vetHeaders,
    body: JSON.stringify({
      email: clientEmail,
      password,
      fullName: 'Client HR Dur',
    }),
  })
  if (!createClient.ok) throw new Error(`create client ${createClient.status}`)

  const clientTok = await apiLogin(clientEmail, password)
  const clientHeaders = {
    Authorization: `Bearer ${clientTok}`,
    'Content-Type': 'application/json',
  }

  const createPet = await fetch(`${API}/api/v1/pets`, {
    method: 'POST',
    headers: clientHeaders,
    body: JSON.stringify({
      name: 'HR Dur Pet',
      species: 'dog',
      breed: 'test',
      plan: 'triennial',
      billingMode: 'subscription',
    }),
  })
  if (!createPet.ok) throw new Error(`create pet ${createPet.status}`)
  const petPayload = (await createPet.json()).data as {
    id?: string
    pet?: { id: string; ownerUserId?: string }
    ownerUserId?: string
  }
  const petId = petPayload.pet?.id ?? petPayload.id
  let ownerId = petPayload.pet?.ownerUserId ?? petPayload.ownerUserId
  if (!petId) throw new Error('missing pet id')
  if (!ownerId) {
    const me = await fetch(`${API}/api/v1/me`, { headers: clientHeaders })
    ownerId = (await me.json()).data.userId as string
  }

  const mockComplete = await fetch(
    `${API}/api/v1/billing/dev/mock-complete?pet_id=${petId}&owner_user_id=${ownerId}&plan_code=triennial&billing_mode=subscription`,
    { headers: clientHeaders },
  )
  if (!mockComplete.ok) throw new Error(`mock-complete ${mockComplete.status}`)

  const petsRes = await fetch(`${API}/api/v1/pets`, { headers: clientHeaders })
  if (!petsRes.ok) throw new Error(`list pets ${petsRes.status}`)
  const pets = (await petsRes.json()).data as Array<{
    id: string
    heartrateDurationsSec?: number[]
  }>
  const pet = pets.find((p) => p.id === petId)
  expect(pet?.heartrateDurationsSec).toEqual([15, 30])

  const start = await fetch(`${API}/api/v1/pets/${petId}/heartrate/sessions`, {
    method: 'POST',
    headers: clientHeaders,
    body: JSON.stringify({ durationSec: 15 }),
  })
  if (!start.ok) throw new Error(`start hr ${start.status}`)
  const sess = (await start.json()).data as { id: string; durationSec: number }
  expect(sess.durationSec).toBe(15)

  const taps = 15
  const complete = await fetch(`${API}/api/v1/heartrate/sessions/${sess.id}`, {
    method: 'PATCH',
    headers: clientHeaders,
    body: JSON.stringify({ tapCount: taps }),
  })
  if (!complete.ok) throw new Error(`complete hr ${complete.status}`)
  const done = (await complete.json()).data as { bpm: number; tapCount: number }
  expect(done.tapCount).toBe(taps)
  expect(done.bpm).toBe(60) // (15 * 60) / 15
})
