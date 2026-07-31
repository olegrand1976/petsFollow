import path from 'node:path'
import { fileURLToPath } from 'node:url'
import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'
import { loginAsVet } from './auth'

export const PACS_API = process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291'

export const PACS_FIXTURE = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../../../../testdata/pacs/demo-rx.dcm',
)

export function pacsPublicFlagOff(): boolean {
  const v = process.env.NUXT_PUBLIC_PACS_ENABLED
  return v === 'false' || v === '0'
}

export function pacsViewerEngine(): string {
  return (process.env.NUXT_PUBLIC_PACS_VIEWER_ENGINE || 'canvas').toLowerCase()
}

async function apiLogin(email: string, password: string): Promise<string> {
  const res = await fetch(`${PACS_API}/api/v1/auth/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password }),
  })
  if (!res.ok) throw new Error(`login ${email} ${res.status}`)
  const body = await res.json()
  return body.data.accessToken as string
}

export async function demoClientAndPet(): Promise<{ clientId: string; petId: string }> {
  const vetTok = await apiLogin('vet.demo@petsfollow.test', 'VetDemo123!')
  const clientsRes = await fetch(`${PACS_API}/api/v1/clients`, {
    headers: { Authorization: `Bearer ${vetTok}` },
  })
  if (!clientsRes.ok) throw new Error(`clients ${clientsRes.status}`)
  const clients = (await clientsRes.json()).data as Array<{ userId: string; email?: string; fullName?: string }>
  const client = clients.find((c) => c.email === 'client.demo@petsfollow.test')
    ?? clients.find((c) => /Sophie/i.test(c.fullName || ''))
  if (!client?.userId) throw new Error('demo client not found')

  const petsRes = await fetch(`${PACS_API}/api/v1/clients/${client.userId}/pets`, {
    headers: { Authorization: `Bearer ${vetTok}` },
  })
  if (!petsRes.ok) throw new Error(`client pets ${petsRes.status}`)
  const pets = (await petsRes.json()).data as Array<{ id: string; paymentStatus?: string }>
  const pet = pets.find((p) => p.paymentStatus === 'active') ?? pets[0]
  if (!pet?.id) throw new Error('demo pet not found')
  return { clientId: client.userId, petId: pet.id }
}

/** Wake PACS if needed, upload fixture, open first study — returns Orthanc instance id from media URL. */
export async function openFirstStudyWithFixture(
  page: Page,
): Promise<{ clientId: string; petId: string; instanceId: string }> {
  const engine = pacsViewerEngine()
  const { clientId, petId } = await demoClientAndPet()
  await loginAsVet(page)
  await page.goto(`/clients/${clientId}/pets/${petId}?tab=imaging`, { waitUntil: 'networkidle' })

  const imagingTabBtn = page.getByTestId('section-tab-imaging')
  expect(await imagingTabBtn.count(), 'imaging tab missing').toBeGreaterThan(0)
  await imagingTabBtn.click()
  await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 15000 })

  const wake = page.getByTestId('pacs-wake-btn')
  if (await wake.isVisible().catch(() => false)) {
    await wake.click()
    await expect(page.getByTestId('pacs-launch-pad')).toBeVisible({ timeout: 10000 })
  }
  await expect(page.getByTestId('pacs-status-badge')).toContainText(
    /ready|prêt|klaar|listo|valmis|pronto/i,
    { timeout: 120000 },
  )
  await expect(page.getByTestId('pacs-launch-pad')).toHaveCount(0, { timeout: 5000 })
  await expect(page.getByTestId('pacs-upload-input')).toBeEnabled()

  await page.getByTestId('pacs-upload-input').setInputFiles(PACS_FIXTURE)
  await expect(page.getByTestId('pacs-study-item').first()).toBeVisible({ timeout: 90000 })

  // Prefer RX/DX / demo fixture over orphan CT indexes (study meta → 502).
  const preferred = page
    .getByTestId('pacs-study-item')
    .filter({ hasText: /rx|thorax|demo|petsfollow|\b(DX|CR|DR|PX)\b/i })
  const studyBtn = (await preferred.count()) > 0 ? preferred.first() : page.getByTestId('pacs-study-item').first()

  const mediaOk = page.waitForResponse(
    (r) => {
      const u = r.url()
      const preview = u.includes('/frames/0/preview') && r.request().method() === 'GET'
      const file = u.includes('/instances/') && u.includes('/file') && r.request().method() === 'GET'
      return engine === 'cornerstone' ? file : preview
    },
    { timeout: 90000 },
  )
  await studyBtn.click()
  const mediaRes = await mediaOk
  expect(mediaRes.status(), 'DICOM media must be 200').toBe(200)

  const m = mediaRes.url().match(/\/instances\/([^/]+)\//)
  const instanceId = m?.[1]
  expect(instanceId, 'instance id from media URL').toBeTruthy()

  if (engine === 'cornerstone') {
    await expect(page.getByTestId('dicom-viewer-cornerstone')).toBeVisible({ timeout: 30000 })
  } else {
    await expect(page.getByTestId('dicom-viewer')).toBeVisible({ timeout: 15000 })
  }

  return { clientId, petId, instanceId: instanceId! }
}
