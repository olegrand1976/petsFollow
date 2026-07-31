import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * Pet sheet Imagerie tab smoke — tag @p0.
 * Preconditions for a real assertion: NUXT_PUBLIC_PACS_ENABLED=true (make nuxtjs-dev)
 * and Orthanc reachable (make up-pacs). Soft-skips when the tab is absent so CI
 * without PACS stays green — that skip does NOT prove Imagerie works.
 */
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

test.describe('PACS pet imaging tab', { tag: '@p0' }, () => {
  test('fiche animal — onglet Imagerie (flag on)', async ({ page }) => {
    test.setTimeout(90000)
    const pacsOn = process.env.NUXT_PUBLIC_PACS_ENABLED
    test.skip(
      pacsOn === 'false' || pacsOn === '0',
      'NUXT_PUBLIC_PACS_ENABLED off',
    )

    const { clientId, petId } = await demoClientAndPet()
    await loginAsVet(page)
    await page.goto(`/clients/${clientId}/pets/${petId}?tab=imaging`, { waitUntil: 'networkidle' })

    const imagingTabBtn = page.getByTestId('section-tab-imaging')
    // Tab absent if public flag off (bake/runtime).
    if ((await imagingTabBtn.count()) === 0) {
      test.skip(true, 'PACS imaging tab not rendered (public flag off)')
    }

    await imagingTabBtn.click()
    await expect(page.getByTestId('pet-tab-imaging')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-dev-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-status-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-refresh-btn')).toBeVisible()
    // Wake only when not already ready (staging Orthanc warm → bouton absent).
    const wake = page.getByTestId('pacs-wake-btn')
    if (await wake.isVisible().catch(() => false)) {
      await expect(wake).toBeEnabled()
    } else {
      await expect(page.getByTestId('pacs-status-badge')).toContainText(/ready|prêt|klaar|listo|valmis|pronto/i)
    }
  })
})
