import { test, expect } from '@playwright/test'
import { loginAsDev } from '../helpers/auth'

/**
 * D17 / D17b — rôle DEV support IT (ops léger, sans billing/sales).
 * Compte seed : dev.demo@petsfollow.test / AdminDemo123!
 */
test.describe('DEV support IT shell', { tag: '@p0' }, () => {
  test('D17 login → users / support / flags ; nav sans billing', async ({ page }) => {
    test.setTimeout(90000)
    await loginAsDev(page)

    await expect(page).toHaveURL(/\/admin(?:\/|$)/, { timeout: 20000 })
    await expect(page.getByTestId('admin-dashboard-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-dev-support-kpi')).toBeVisible({ timeout: 15000 })

    await expect(page.locator('a[href="/admin/payments"]')).toHaveCount(0)
    await expect(page.locator('a[href="/admin/commercials"]')).toHaveCount(0)
    await expect(page.locator('a[href="/admin/commissions"]')).toHaveCount(0)
    await expect(page.locator('a[href="/usecases"]')).toHaveCount(0)

    await page.goto('/admin/users', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-users-page')).toBeVisible({ timeout: 15000 })

    await page.goto('/admin/support', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-support-page')).toBeVisible({ timeout: 15000 })

    await page.goto('/admin/runtime-flags', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-runtime-flags-page')).toBeVisible({ timeout: 15000 })
  })

  test('D17b billing pages redirect hors ops DEV', async ({ page }) => {
    test.setTimeout(60000)
    await loginAsDev(page)

    await page.goto('/admin/payments', { waitUntil: 'networkidle' })
    await expect(page).not.toHaveURL(/\/admin\/payments/, { timeout: 15000 })
    await expect(page).toHaveURL(/\/admin(?:\/|$)/, { timeout: 15000 })

    await page.goto('/admin/commercials', { waitUntil: 'networkidle' })
    await expect(page).not.toHaveURL(/\/admin\/commercials/, { timeout: 15000 })
    await expect(page).toHaveURL(/\/admin(?:\/|$)/, { timeout: 15000 })
  })
})
