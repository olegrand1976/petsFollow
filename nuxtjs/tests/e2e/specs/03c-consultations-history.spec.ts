import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import { softDeleteVisit } from '../helpers/cleanup'

/** C2.17 — historique /consultations (walk-in date DESC + filtres + ouvrir CR). */

async function dismissProModals(page: Page) {
  await page.getByTestId('pro-modal').first()
    .waitFor({ state: 'visible', timeout: 1500 })
    .catch(() => undefined)
  for (let i = 0; i < 3; i++) {
    const modal = page.getByTestId('pro-modal')
    if ((await modal.count()) === 0) return
    const close = page.getByTestId('pro-modal-close')
    if ((await close.count()) > 0) {
      await close.first().click({ force: true })
    }
    else {
      await page.keyboard.press('Escape')
    }
    await expect(modal).toHaveCount(0, { timeout: 5000 }).catch(() => undefined)
  }
}

async function createWalkInWithReport(page: Page): Promise<string> {
  await page.goto('/clients', { waitUntil: 'networkidle' })
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
  await expect(petSelect).toBeEnabled({ timeout: 10000 })
  const options = petSelect.locator('option:not([disabled])')
  await options.first().waitFor({ state: 'attached', timeout: 10000 })
  const value = await options.first().getAttribute('value')
  if (value) await petSelect.selectOption(value)

  const createRes = page.waitForResponse(
    (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('consultation-start').click()
  const created = await createRes
  expect([200, 201]).toContain(created.status())
  const body = await created.json().catch(() => null) as any
  const visitId = String((body?.data ?? body)?.id || '')
  expect(visitId).toBeTruthy()

  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
  const reportBody = page.getByTestId('visit-report-body')
  // Wait until hydrate unlocks the textarea (avoids racing GET /report → empty PUT).
  await expect(reportBody).toBeEnabled({ timeout: 20000 })
  const probe = `E2E history CR ${Date.now()}`
  await reportBody.fill(probe)
  await expect(reportBody).toHaveValue(new RegExp(probe.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')))
  await page.getByTestId('visit-report-save').click()
  await expect(page.getByTestId('consultation-cta-done')).toBeVisible({ timeout: 15000 })
  await page.getByTestId('consultation-cta-done').click()
  await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
  return visitId
}

test.describe('historique consultations', { tag: '@p1' }, () => {
  test('liste /consultations + filtre + ouvrir CR', async ({ page }) => {
    await loginAsVet(page)
    const visitId = await createWalkInWithReport(page)
    try {
      const listRes = page.waitForResponse(
        (r) => r.url().includes('/api/vet/consultations') && r.request().method() === 'GET',
        { timeout: 20000 },
      )
      await page.goto('/consultations', { waitUntil: 'networkidle' })
      await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
      const listed = await listRes
      expect(listed.status()).toBe(200)

      await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 15000 })
      await expect(page.getByTestId('consultations-search')).toBeVisible()
      await expect(page.getByTestId('consultations-status-filter')).toBeVisible()

      await page.getByTestId('consultations-search').fill('Sophie')
      await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 10000 })

      const getReport = page.waitForResponse(
        (r) => r.url().includes(`/api/visits/${visitId}/report`) && r.request().method() === 'GET',
        { timeout: 20000 },
      )
      await page.getByTestId(`consultation-open-cr-${visitId}`).click()
      await expect(page.getByTestId('consultation-history-report-modal')).toBeVisible({ timeout: 10000 })
      await expect(page.getByTestId('visit-report-panel')).toBeVisible()
      await getReport
      await expect(page.getByTestId('visit-report-body')).toHaveValue(/E2E history CR/, { timeout: 15000 })
    }
    finally {
      await softDeleteVisit(page, visitId)
    }
  })

  test('soft-delete retire la consultation de la liste', async ({ page }) => {
    await loginAsVet(page)
    const visitId = await createWalkInWithReport(page)

    await page.goto('/consultations', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 15000 })

    await page.getByTestId(`consultation-delete-${visitId}`).click()
    await expect(page.getByTestId('consultation-delete-modal')).toBeVisible()
    const delRes = page.waitForResponse(
      (r) => r.url().includes(`/api/visits/${visitId}`) && r.request().method() === 'DELETE',
      { timeout: 20000 },
    )
    await page.getByTestId('consultation-delete-confirm').click()
    expect((await delRes).status()).toBe(200)
    await expect(page.getByTestId(`consultation-row-${visitId}`)).toHaveCount(0, { timeout: 10000 })
  })

  test('durée audio affichée + écoute player', async ({ page, request }) => {
    const API = process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291'

    const loginRes = await request.post(`${API}/api/v1/auth/login`, {
      data: { email: 'vet.demo@petsfollow.test', password: 'VetDemo123!' },
    })
    expect(loginRes.ok()).toBeTruthy()
    const vetTok = (await loginRes.json()).data.accessToken as string

    const clientLogin = await request.post(`${API}/api/v1/auth/login`, {
      data: { email: 'client.demo@petsfollow.test', password: 'ClientDemo123!' },
    })
    expect(clientLogin.ok()).toBeTruthy()
    const clientTok = (await clientLogin.json()).data.accessToken as string
    const petsRes = await request.get(`${API}/api/v1/pets`, {
      headers: { Authorization: `Bearer ${clientTok}` },
    })
    expect(petsRes.ok()).toBeTruthy()
    const pets = (await petsRes.json()).data as Array<{ id: string }>
    expect(pets.length).toBeGreaterThan(0)
    const petId = pets[0].id

    const create = await request.post(`${API}/api/v1/pets/${petId}/visits`, {
      headers: { Authorization: `Bearer ${vetTok}` },
      data: {
        scheduledAt: new Date(Date.now() + 5 * 60_000).toISOString(),
        durationMinutes: 30,
        confirmDirect: true,
        silentConfirm: true,
        consultationSession: true,
        notes: `e2e audio duration ${Date.now()}`,
      },
    })
    expect([200, 201]).toContain(create.status())
    const visitId = (await create.json()).data.id as string

    try {
      const put = await request.put(`${API}/api/v1/visits/${visitId}/report`, {
        headers: { Authorization: `Bearer ${vetTok}` },
        data: { bodyText: 'E2E draft CR with pending audio' },
      })
      expect(put.ok()).toBeTruthy()

      const form = new FormData()
      form.append('clientAudioConsent', 'true')
      form.append('audioDurationSec', '154')
      form.append('hint', 'E2E audio duration transcript')
      form.append('audio', new Blob(['fake-webm-e2e'], { type: 'audio/webm' }), 'dictation.webm')
      const tr = await fetch(`${API}/api/v1/visits/${visitId}/report/transcribe`, {
        method: 'POST',
        headers: { Authorization: `Bearer ${vetTok}` },
        body: form,
      })
      const trText = await tr.text()
      expect(tr.status, trText).toBe(200)
      const trBody = JSON.parse(trText) as { data?: { hasAudio?: boolean; audioDurationSec?: number } }
      expect(trBody.data?.hasAudio).toBe(true)
      expect(trBody.data?.audioDurationSec).toBe(154)

      await loginAsVet(page)
      await page.goto('/consultations', { waitUntil: 'networkidle' })
      await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
      await page.getByTestId('consultations-audio-only').check()
      await expect(page.getByTestId(`consultation-row-${visitId}`)).toBeVisible({ timeout: 15000 })
      await expect(page.getByTestId(`consultation-audio-duration-${visitId}`)).toHaveText('02:34')

      await page.getByTestId(`consultation-audio-${visitId}`).click()
      await expect(page.getByTestId('consultation-audio-modal')).toBeVisible({ timeout: 10000 })
      await expect(page.getByTestId('consultation-audio-duration-label')).toContainText('02:34')
      await expect(page.getByTestId('consultation-audio-player')).toBeVisible({ timeout: 15000 })
    }
    finally {
      await fetch(`${API}/api/v1/visits/${visitId}`, {
        method: 'DELETE',
        headers: { Authorization: `Bearer ${vetTok}` },
      }).catch(() => undefined)
    }
  })
})
