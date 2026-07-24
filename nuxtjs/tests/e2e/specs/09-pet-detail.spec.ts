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

test('pet detail — chart filtres, shares, commentaire HR', async ({ page }) => {
  test.setTimeout(60000)
  const { clientId, petId } = await demoClientAndPet()
  const comment = await seedHeartRateComment(petId)

  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}`)
  await expect(page.getByTestId('pet-detail-page')).toBeVisible()
  await expect(page.getByTestId('pet-chart-range-3m')).toBeVisible()
  await page.getByTestId('pet-chart-range-6m').click()
  await page.getByTestId('pet-filter-all').click()
  await expect(page.getByTestId('pet-shares-card')).toBeVisible()

  await expect(page.getByTestId('pet-reading-comment').filter({ hasText: comment })).toBeVisible({
    timeout: 15000,
  })
})
