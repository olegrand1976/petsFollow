import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../helpers/auth'

/**
 * PACS admin dashboard smoke — tag @p0, skip if NUXT_PUBLIC_PACS_ENABLED off
 * or admin seed missing.
 */
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

test.describe('PACS admin dashboard', { tag: '@p0' }, () => {
  test.beforeAll(async ({ request }) => {
    const apiBase = (process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291')
      .replace(/\/$/, '')
    const res = await request.post(`${apiBase}/api/v1/auth/login`, {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    test.skip(res.status() !== 200, `admin.demo seed absent (HTTP ${res.status()})`)
  })

  test('admin /admin/pacs metrics + logs (flag on)', async ({ page }) => {
    test.setTimeout(90000)
    const pacsOn = process.env.NUXT_PUBLIC_PACS_ENABLED
    test.skip(
      pacsOn === 'false' || pacsOn === '0',
      'NUXT_PUBLIC_PACS_ENABLED off',
    )

    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin/pacs', { waitUntil: 'networkidle' })

    const disabled = page.getByTestId('admin-pacs-disabled')
    if (await disabled.isVisible().catch(() => false)) {
      test.skip(true, 'PACS public flag off in this build')
    }

    await expect(page.getByTestId('admin-pacs-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-pacs-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-wake')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-metrics')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-logs')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-playground')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-debug')).toBeVisible()
    await expect(page.getByTestId('admin-pacs-pet-select')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 20000 })
  })
})
