import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * C2.26 — Améliorer IA avancé (SSE Agent Loader), mocks BFF (pas de CrewAI live).
 * Tag @p1. Flag NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED on en CI / staging.
 */

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

async function openConsultationReport(page: Page) {
  await page.goto('/clients', { waitUntil: 'networkidle' })
  if (page.url().includes('/login')) {
    await loginAsVet(page)
    await page.goto('/clients', { waitUntil: 'networkidle' })
  }
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
  await expect(petSelect).toBeEnabled({ timeout: 15000 })
  const options = petSelect.locator('option:not([disabled])')
  await options.first().waitFor({ state: 'attached', timeout: 15000 })
  const value = await options.first().getAttribute('value')
  expect(value).toBeTruthy()
  await petSelect.selectOption(value!)

  const createRes = page.waitForResponse(
    (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('consultation-start').click()
  const created = await createRes
  expect([200, 201]).toContain(created.status())
  const createdBody = await created.json().catch(() => null)
  const visitId = String((createdBody as any)?.data?.id ?? (createdBody as any)?.id ?? '')
  expect(visitId).toBeTruthy()
  await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
  return visitId
}

const advancedFlagOn = (() => {
  const v = process.env.NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED
  return v !== 'false' && v !== '0'
})()

test.describe('CR improve-advanced SSE (C2.26)', { tag: '@p1' }, () => {
  test.skip(!advancedFlagOn, 'NUXT_PUBLIC_AI_CR_ADVANCED_ENABLED off')

  test('bouton avancé → loader steps → final appliqué au body', async ({ page }) => {
    test.setTimeout(120000)
    await loginAsVet(page)
    const visitId = await openConsultationReport(page)

    const advancedBtn = page.getByTestId('visit-report-improve-advanced')
    await expect(advancedBtn).toBeVisible({ timeout: 15000 })

    const notes = `E2E advanced notes ${Date.now()}`
    const improvedMarker = `E2E advanced improved ${Date.now()}`
    // TipTap round-trip may drop ATX ## headings — assert on unique body text.
    const improved = `## Anamnèse / motif\n\n${improvedMarker}`
    await page.getByTestId('visit-report-transcript').fill(notes)
    await expect(advancedBtn).toBeEnabled()

    const runId = 'e2e-adv-run-1'
    let resolveEvents: (() => void) | null = null
    const eventsGate = new Promise<void>((resolve) => {
      resolveEvents = resolve
    })
    const releaseEvents = () => {
      resolveEvents?.()
      resolveEvents = null
    }

    // Anchor on end-of-path so …/events is not swallowed by this route + continue().
    await page.route(/\/api\/visits\/[^/]+\/report-improve-advanced$/, async (route) => {
      if (route.request().method() !== 'POST') {
        await route.fallback()
        return
      }
      await route.fulfill({
        status: 202,
        contentType: 'application/json',
        body: JSON.stringify({ data: { runId, status: 'queued' } }),
      })
    })

    await page.route(/\/api\/visits\/[^/]+\/report-improve-advanced\/[^/]+\/events$/, async (route) => {
      await eventsGate
      const sse = [
        'id: 1\nevent: step\ndata: {"agent":"tri","label":"Extraction"}\n\n',
        'id: 2\nevent: step\ndata: {"agent":"clinicien","label":"RAG"}\n\n',
        'id: 3\nevent: step\ndata: {"agent":"redacteur","label":"Sections"}\n\n',
        `event: final\ndata: ${JSON.stringify({ report: improved })}\n\n`,
        'event: ping\ndata: {"done":true}\n\n',
      ].join('')
      await route.fulfill({
        status: 200,
        headers: {
          'content-type': 'text/event-stream',
          'cache-control': 'no-cache',
        },
        body: sse,
      })
    })

    // After SSE final, UI re-fetches GET report — return improved payload.
    await page.route(new RegExp(`/api/visits/${visitId}/report$`), async (route) => {
      if (route.request().method() !== 'GET') {
        await route.fallback()
        return
      }
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: {
            id: 'e2e-adv-report',
            visitId,
            status: 'draft',
            bodyText: improved,
            transcriptText: notes,
            improvedText: improved,
            isReference: false,
          },
        }),
      })
    })

    const sawThreeSteps = page.waitForFunction(() => {
      const ol = document.querySelector('[data-testid="visit-report-advanced-steps"]')
      return Boolean(ol && ol.querySelectorAll('li').length >= 3)
    }, { timeout: 20000 })

    try {
      await advancedBtn.click()
      await expect(page.getByTestId('visit-report-advanced-improving')).toBeVisible({ timeout: 10000 })
      await expect(page.getByTestId('visit-report-advanced-cancel')).toBeVisible()
    }
    finally {
      releaseEvents()
    }

    await sawThreeSteps
    await expect(page.getByTestId('visit-report-advanced-improving')).toHaveCount(0, { timeout: 15000 })
    await expect(page.getByTestId('visit-report-body')).toHaveValue(improved, { timeout: 15000 })
    await expect(page.getByTestId('visit-report-quality')).toBeVisible({ timeout: 10000 })
  })
})
