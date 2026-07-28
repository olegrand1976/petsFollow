import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * Anti-régression Nitro : conflit fichier/dossier
 * (`report.put.ts` + `report/`) masquait POST …/report/improve|finalize|transcribe
 * → 404 « Page not found » côté BFF (HAR prod).
 *
 * On assert que la route BFF existe (pas le succès Gemini).
 * Statuts métier OK : 200, 402 (module IA), 503 (Gemini off).
 */

/** Payload Nitro (pas l’enveloppe Go `{ error: { code } }`). */
function isNuxtPageNotFound(status: number, body: unknown): boolean {
  if (status !== 404 || !body || typeof body !== 'object') return false
  const o = body as {
    error?: unknown
    statusMessage?: string
    message?: string
    statusCode?: number
  }
  const msg = `${o.statusMessage || ''} ${o.message || ''}`
  if (/page not found/i.test(msg)) return true
  // Nuxt : `"error": true` (bool) ; Go : `"error": { code, ... }`
  return o.error === true && o.statusCode === 404
}

function unwrapId(body: unknown): string {
  const o = body as { data?: { id?: string }, id?: string } | null
  return String(o?.data?.id || o?.id || '')
}

test.describe('BFF visit report AI routes', { tag: '@p1' }, () => {
  test('POST report/improve + finalize ne renvoient pas 404 Nitro', async ({ page }) => {
    test.setTimeout(90000)
    await loginAsVet(page)

    const petsRes = await page.request.get('/api/vet/pets')
    expect(petsRes.status(), await petsRes.text()).toBe(200)
    const petsBody = await petsRes.json()
    const pets = (petsBody?.data ?? petsBody) as Array<{ id?: string }>
    const petId = String(pets?.[0]?.id || '')
    expect(petId, 'seed pet attendu').toBeTruthy()

    const when = new Date(
      Date.now() + 4 * 60 * 60 * 1000 + Math.floor(Math.random() * 40) * 60_000,
    ).toISOString()
    const createRes = await page.request.post(`/api/pets/${petId}/visits`, {
      data: {
        notes: `e2e bff improve ${Date.now()}`,
        confirmDirect: true,
        consultationSession: true,
        scheduledAt: when,
        durationMinutes: 30,
      },
    })
    // 409 slot_taken : retry géré ci-dessous inutile — on exige une visite.
    expect([200, 201], await createRes.text()).toContain(createRes.status())
    const visitId = unwrapId(await createRes.json())
    expect(visitId).toBeTruthy()

    const putRes = await page.request.put(`/api/visits/${visitId}/report`, {
      data: { bodyText: `E2E improve BFF ${Date.now()}\nAnamnese courte pour IA.` },
    })
    expect(putRes.status(), await putRes.text()).toBe(200)

    const improveRes = await page.request.post(`/api/visits/${visitId}/report/improve`)
    const improveBody = await improveRes.json().catch(() => null)
    expect(
      isNuxtPageNotFound(improveRes.status(), improveBody),
      `improve ne doit pas être 404 Nitro, got ${improveRes.status()} ${JSON.stringify(improveBody)}`,
    ).toBe(false)
    expect([200, 402, 503]).toContain(improveRes.status())

    const finalizeRes = await page.request.post(`/api/visits/${visitId}/report/finalize`)
    const finalizeBody = await finalizeRes.json().catch(() => null)
    expect(
      isNuxtPageNotFound(finalizeRes.status(), finalizeBody),
      `finalize ne doit pas être 404 Nitro, got ${finalizeRes.status()} ${JSON.stringify(finalizeBody)}`,
    ).toBe(false)
    // 200 finalisé · 402 module · 503 Gemini · 409 déjà final (si improve a finalisé) · 400 body vide
    expect([200, 400, 402, 409, 503]).toContain(finalizeRes.status())

    // Nested me/ai-module/* : même classe de conflit fichier/dossier.
    const roiRes = await page.request.get('/api/me/ai-module/roi')
    const roiBody = await roiRes.json().catch(() => null)
    expect(
      isNuxtPageNotFound(roiRes.status(), roiBody),
      `ai-module/roi ne doit pas être 404 Nitro, got ${roiRes.status()} ${JSON.stringify(roiBody)}`,
    ).toBe(false)

    // Cleanup walk-in
    await page.request.patch(`/api/visits/${visitId}`, { data: { status: 'cancelled' } }).catch(() => undefined)
  })
})
