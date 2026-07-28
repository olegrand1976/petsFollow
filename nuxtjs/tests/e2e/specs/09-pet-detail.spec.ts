import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

const API = process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291'

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
  const pets = (await petsRes.json()).data as Array<{ id: string; paymentStatus?: string }>
  const pet = pets.find((p) => p.paymentStatus === 'active') ?? pets[0]
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
  const openReport = page.getByTestId('pet-visit-report-open').first()
  await expect(openReport).toBeVisible({ timeout: 30000 })
  await openReport.click()
  await expect(page.getByTestId('visit-report-panel')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('visit-report-body')).toBeVisible()
  await expect(page.getByTestId('visit-report-improve')).toBeVisible()
  await expect(page.getByTestId('visit-report-audio')).toBeAttached()
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
