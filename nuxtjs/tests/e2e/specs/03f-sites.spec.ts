import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import { ensureMultiSite } from '../helpers/sites'

/**
 * C3.16 — multi-sites Pro (seed VetPlus: primary + Antenne Liège).
 * Requires API + NUXT_PUBLIC_SITES_UI_ENABLED on. Creates a 2nd site if staging
 * was deployed sans seed.
 */
test.describe('multi-sites settings + switcher', { tag: '@p1' }, () => {
  test('vet.demo — switcher, settings rename, consultations site column', async ({ page }) => {
    await loginAsVet(page)
    await ensureMultiSite(page)

    await expect(page.getByTestId('pro-site-switcher')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pro-site-select')).toBeVisible()

    await page.goto('/settings#calendar')
    await expect(page.getByTestId('settings-tab-calendar')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('settings-sites')).toBeVisible()
    const rows = page.getByTestId('settings-sites-list').locator('.settings-sites-row')
    await expect.poll(async () => rows.count()).toBeGreaterThanOrEqual(2)
    await expect(page.getByTestId('settings-site-create')).toBeVisible()
    await expect(page.getByTestId('settings-site-name')).toBeVisible()

    // Rename a non-primary site (has deactivate), then restore to avoid seed pollution.
    const nonPrimaryRow = page.locator('.settings-sites-row').filter({
      has: page.locator('[data-testid^="settings-site-deactivate-"]'),
    }).first()
    await expect(nonPrimaryRow).toBeVisible()
    const renameBtn = nonPrimaryRow.locator('[data-testid^="settings-site-rename-"]')
    await expect(renameBtn).toBeVisible()
    const renameTestId = await renameBtn.getAttribute('data-testid')
    const siteId = renameTestId?.replace('settings-site-rename-', '') || ''
    expect(siteId).toBeTruthy()

    await renameBtn.click()
    const renameInput = page.getByTestId(`settings-site-rename-input-${siteId}`)
    await expect(renameInput).toBeVisible()
    const originalName = await renameInput.inputValue()
    const renamed = `${originalName.replace(/\s*\(e2e\)\s*$/, '')} (e2e)`
    await renameInput.fill(renamed)
    await page.getByTestId(`settings-site-rename-save-${siteId}`).click()
    await expect(nonPrimaryRow.getByText(renamed, { exact: false })).toBeVisible({ timeout: 10000 })

    // Restore original name.
    await nonPrimaryRow.locator(`[data-testid="settings-site-rename-${siteId}"]`).click()
    await expect(renameInput).toBeVisible()
    await renameInput.fill(originalName)
    await page.getByTestId(`settings-site-rename-save-${siteId}`).click()
    await expect(nonPrimaryRow.getByText(originalName, { exact: true })).toBeVisible({ timeout: 10000 })

    // Consultations: multi-site column visible when flag + ≥2 sites.
    await page.goto('/consultations')
    await expect(page.getByTestId('consultations-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pro-site-select')).toBeVisible()
    await expect(page.getByTestId('consultations-column-site')).toBeVisible({ timeout: 10000 })
  })
})
