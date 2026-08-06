import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * C3.16 — multi-sites Pro (seed VetPlus: primary + Antenne Liège).
 * Requires API seeded + NUXT_PUBLIC_SITES_UI_ENABLED on (nuxtjs-dev default).
 */
test.describe('multi-sites settings + switcher', { tag: '@p1' }, () => {
  test('vet.demo — switcher + settings sites list', async ({ page }) => {
    await loginAsVet(page)

    await expect(page.getByTestId('pro-site-switcher')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pro-site-select')).toBeVisible()

    await page.goto('/settings#calendar')
    await expect(page.getByTestId('settings-tab-calendar')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('settings-sites')).toBeVisible()
    const rows = page.getByTestId('settings-sites-list').locator('.settings-sites-row')
    await expect.poll(async () => rows.count()).toBeGreaterThanOrEqual(2)
    await expect(page.getByTestId('settings-site-create')).toBeVisible()
    await expect(page.getByTestId('settings-site-name')).toBeVisible()
  })
})
