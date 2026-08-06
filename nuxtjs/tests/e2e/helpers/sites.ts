import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'
import { fillField } from './auth'

/**
 * Staging Cloud Run e2e runs sans re-seed — ensure VetPlus has ≥2 sites so the
 * topbar switcher / SITE_ALL calendar path exist (soft-GA multi-site UI).
 */
export async function ensureMultiSite(page: Page) {
  await page.goto('/settings#calendar', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('settings-tab-calendar')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('settings-sites')).toBeVisible({ timeout: 15000 })

  const rows = page.getByTestId('settings-sites-list').locator('[data-testid^="settings-site-row-"]')
  await expect.poll(async () => rows.count(), { timeout: 15000 }).toBeGreaterThanOrEqual(1)

  if ((await rows.count()) < 2) {
    const name = `E2E Antenne ${Date.now()}`
    await fillField(page, 'settings-site-name', name)
    await Promise.all([
      page.waitForResponse(
        (r) => r.request().method() === 'POST' && r.url().includes('/vet/sites'),
        { timeout: 20000 },
      ),
      page.getByTestId('settings-site-create').click(),
    ])
    await expect.poll(async () => rows.count(), { timeout: 15000 }).toBeGreaterThanOrEqual(2)
  }

  // Reload shell so ProTopbar picks up multiSite from /me.
  await page.goto('/', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('pro-site-switcher')).toBeVisible({ timeout: 15000 })
}
