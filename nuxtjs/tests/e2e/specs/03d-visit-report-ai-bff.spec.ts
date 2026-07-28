import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * Anti-régression BFF CR IA.
 *
 * Historique : nested `…/report/improve` était enregistré dans Nitro mais
 * non matché au runtime (leaf `…/report` GET/PUT) → 404 « Page not found ».
 * Les actions POST sont aplaties : `/report-improve`, `/report-finalize`, `/report-transcribe`.
 *
 * Assert route vivante (pas succès Gemini) : 200 / 402 / 503.
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
  return o.error === true && o.statusCode === 404
}

function unwrapId(body: unknown): string {
  const o = body as { data?: { id?: string }, id?: string } | null
  return String(o?.data?.id || o?.id || '')
}

test.describe('BFF visit report AI routes', { tag: '@p1' }, () => {
  test('POST report-improve + report-finalize ne renvoient pas 404 Nitro', async ({ page }) => {
    test.setTimeout(90000)
    await loginAsVet(page)

    // Pet hors client.demo (évite de polluer l’historique de 09-pet-detail).
    const petsRes = await page.request.get('/api/vet/pets')
    expect(petsRes.status(), await petsRes.text()).toBe(200)
    const petsBody = await petsRes.json()
    const pets = (petsBody?.data ?? petsBody) as Array<{
      id?: string
      name?: string
      ownerName?: string
    }>
    const pet = pets.find(p =>
      /max/i.test(String(p.name || '')) && /paul/i.test(String(p.ownerName || '')),
    ) || pets.find(p => /max/i.test(String(p.name || '')))
    const petId = String(pet?.id || '')
    expect(petId, 'pet Max (Paul) seed').toBeTruthy()

    let visitId = ''
    for (let attempt = 0; attempt < 4 && !visitId; attempt++) {
      const when = new Date(
        Date.now() + (4 + attempt) * 60 * 60 * 1000 + Math.floor(Math.random() * 40) * 60_000,
      ).toISOString()
      const createRes = await page.request.post(`/api/pets/${petId}/visits`, {
        data: {
          notes: `e2e bff improve ${Date.now()}`,
          confirmDirect: true,
          scheduledAt: when,
          durationMinutes: 30,
        },
      })
      if (createRes.status() === 409) continue
      expect([200, 201], await createRes.text()).toContain(createRes.status())
      visitId = unwrapId(await createRes.json())
    }
    expect(visitId, 'visite créée').toBeTruthy()

    try {
      const putRes = await page.request.put(`/api/visits/${visitId}/report`, {
        data: { bodyText: `E2E improve BFF ${Date.now()}\nAnamnese courte pour IA.` },
      })
      expect(putRes.status(), await putRes.text()).toBe(200)

      const improveRes = await page.request.post(`/api/visits/${visitId}/report-improve`)
      const improveBody = await improveRes.json().catch(() => null)
      expect(
        isNuxtPageNotFound(improveRes.status(), improveBody),
        `report-improve ne doit pas être 404 Nitro, got ${improveRes.status()} ${JSON.stringify(improveBody)}`,
      ).toBe(false)
      expect([200, 402, 503]).toContain(improveRes.status())

      const finalizeRes = await page.request.post(`/api/visits/${visitId}/report-finalize`)
      const finalizeBody = await finalizeRes.json().catch(() => null)
      expect(
        isNuxtPageNotFound(finalizeRes.status(), finalizeBody),
        `report-finalize ne doit pas être 404 Nitro, got ${finalizeRes.status()} ${JSON.stringify(finalizeBody)}`,
      ).toBe(false)
      expect([200, 400, 402, 409, 503]).toContain(finalizeRes.status())

      const roiRes = await page.request.get('/api/me/ai-module/roi')
      const roiBody = await roiRes.json().catch(() => null)
      expect(
        isNuxtPageNotFound(roiRes.status(), roiBody),
        `ai-module/roi ne doit pas être 404 Nitro, got ${roiRes.status()} ${JSON.stringify(roiBody)}`,
      ).toBe(false)
    }
    finally {
      await page.request.patch(`/api/visits/${visitId}`, { data: { status: 'cancelled' } }).catch(() => undefined)
    }
  })
})
