import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/**
 * C3.16 — multi-sites Pro (seed VetPlus: primary + Antenne Liège).
 * Requires API seeded + NUXT_PUBLIC_SITES_UI_ENABLED on (nuxtjs-dev default).
 * Deep CRUD / overlap / booking: Go + Flutter widgets (see 15-PLAN-TESTS C3.16).
 */
test.describe('multi-sites settings + switcher', { tag: '@p1' }, () => {
  test('vet.demo — switcher, settings rename, filter consultations', async ({ page }) => {
    await loginAsVet(page)

    await expect(page.getByTestId('pro-site-switcher')).toBeVisible({ timeout: 15000 })
    const siteSelect = page.getByTestId('pro-site-select')
    await expect(siteSelect).toBeVisible()

    await page.goto('/settings#calendar')
    await expect(page.getByTestId('settings-tab-calendar')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('settings-sites')).toBeVisible()
    const rows = page.getByTestId('settings-sites-list').locator('.settings-sites-row')
    await expect.poll(async () => rows.count()).toBeGreaterThanOrEqual(2)
    await expect(page.getByTestId('settings-site-create')).toBeVisible()
    await expect(page.getByTestId('settings-site-name')).toBeVisible()

    // Rename: first active site rename control.
    const renameBtn = page.locator('[data-testid^="settings-site-rename-"]').first()
    await expect(renameBtn).toBeVisible()
    const renameTestId = await renameBtn.getAttribute('data-testid')
    const siteId = renameTestId?.replace('settings-site-rename-', '') || ''
    expect(siteId).toBeTruthy()
    await renameBtn.click()
    const renameInput = page.getByTestId(`settings-site-rename-input-${siteId}`)
    await expect(renameInput).toBeVisible()
    const currentName = await renameInput.inputValue()
    const renamed = currentName.includes('(e2e)')
      ? currentName.replace(/\s*\(e2e\)\s*$/, '')
      : `${currentName} (e2e)`
    await renameInput.fill(renamed)
    await page.getByTestId(`settings-site-rename-save-${siteId}`).click()
    await expect(page.getByText(renamed, { exact: false }).first()).toBeVisible({ timeout: 10000 })

    // Consultations: multi-site column visible (any selected site or all).
    await page.goto('/consultations')
    await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pro-site-select')).toBeVisible()
    await expect(page.getByTestId('consultations-column-site')).toBeVisible({ timeout: 10000 })
  })
})
