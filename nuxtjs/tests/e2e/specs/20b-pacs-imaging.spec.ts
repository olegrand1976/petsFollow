import path from 'node:path'
import { fileURLToPath } from 'node:url'
import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * Pet sheet Imagerie — tag @p0.
 * Requires NUXT_PUBLIC_PACS_ENABLED + Orthanc (local make up-pacs / staging warm).
 * Soft-skips only when the public flag is off so CI without PACS stays green.
 * When the tab is present, upload + canvas assertions are hard (no soft-skip on Orthanc).
 */
const API = process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291'
const FIXTURE = path.resolve(
  path.dirname(fileURLToPath(import.meta.url)),
  '../../../../testdata/pacs/demo-rx.dcm',
)

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

async function canvasHasNonBackgroundPixels(page: import('@playwright/test').Page): Promise<boolean> {
  return page.getByTestId('dicom-canvas-left').evaluate((canvas) => {
    const c = canvas as HTMLCanvasElement
    const ctx = c.getContext('2d')
    if (!ctx || c.width < 8 || c.height < 8) return false
    const { data } = ctx.getImageData(0, 0, c.width, c.height)
    // Background fill is #0b1220 — count pixels that clearly diverge.
    let lit = 0
    for (let i = 0; i < data.length; i += 16) {
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      if (Math.abs(r - 0x0b) > 12 || Math.abs(g - 0x12) > 12 || Math.abs(b - 0x20) > 12) {
        lit++
        if (lit > 40) return true
      }
    }
    return lit > 40
  })
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
    test.skip(
      (await imagingTabBtn.count()) === 0,
      'PACS imaging tab not rendered (public flag off)',
    )

    await imagingTabBtn.click()
    await expect(page.getByTestId('pet-tab-imaging')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-dev-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-status-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-refresh-btn')).toBeVisible()
    const wake = page.getByTestId('pacs-wake-btn')
    if (await wake.isVisible().catch(() => false)) {
      await expect(wake).toBeEnabled()
    } else {
      await expect(page.getByTestId('pacs-status-badge')).toContainText(/ready|prêt|klaar|listo|valmis|pronto/i)
    }
  })

  test('upload fixture + preview canvas non vide', async ({ page }) => {
    test.setTimeout(180000)
    const pacsOn = process.env.NUXT_PUBLIC_PACS_ENABLED
    test.skip(
      pacsOn === 'false' || pacsOn === '0',
      'NUXT_PUBLIC_PACS_ENABLED off',
    )

    const { clientId, petId } = await demoClientAndPet()
    await loginAsVet(page)
    await page.goto(`/clients/${clientId}/pets/${petId}?tab=imaging`, { waitUntil: 'networkidle' })

    const imagingTabBtn = page.getByTestId('section-tab-imaging')
    test.skip(
      (await imagingTabBtn.count()) === 0,
      'PACS imaging tab not rendered (public flag off)',
    )
    await imagingTabBtn.click()
    await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 15000 })

    const wake = page.getByTestId('pacs-wake-btn')
    if (await wake.isVisible().catch(() => false)) {
      await wake.click()
    }
    await expect(page.getByTestId('pacs-status-badge')).toContainText(
      /ready|prêt|klaar|listo|valmis|pronto/i,
      { timeout: 120000 },
    )

    await page.getByTestId('pacs-upload-input').setInputFiles(FIXTURE)

    // Study list grows / existing study clickable after upload (newest first).
    await expect(page.getByTestId('pacs-study-item').first()).toBeVisible({ timeout: 90000 })
    const previewOk = page.waitForResponse(
      (r) => r.url().includes('/frames/0/preview') && r.request().method() === 'GET',
      { timeout: 90000 },
    )
    await page.getByTestId('pacs-study-item').first().click()

    const previewRes = await previewOk
    expect(previewRes.status(), 'preview must be 200 (ParentSeries authz)').toBe(200)

    await expect(page.getByTestId('dicom-viewer')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('dicom-canvas-left')).toBeVisible()
    await expect(page.getByTestId('dicom-preview-error')).toHaveCount(0)

    await expect.poll(async () => canvasHasNonBackgroundPixels(page), {
      timeout: 20000,
      message: 'canvas left should show non-background pixels after preview load',
    }).toBe(true)

    await expect(page.getByTestId('dicom-tool-pan')).toBeVisible()
    await expect(page.getByTestId('dicom-tool-zoom')).toBeVisible()
    await expect(page.getByTestId('dicom-tool-wl')).toBeVisible()
    await page.getByTestId('pacs-compare-toggle').check()
    // Compare toggle must not blank the left pane.
    await expect.poll(async () => canvasHasNonBackgroundPixels(page), { timeout: 10000 }).toBe(true)
  })
})
