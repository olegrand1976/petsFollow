import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../helpers/auth'

/**
 * AFMPS CSV admin — smoke navigation (middleware admin-only).
 * Tag @p1 : page list + lien sidebar ; pas d’import CSV live.
 */
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

test.describe('AFMPS admin imports', { tag: ['@p1', '@pharmacy'] }, () => {
  test.beforeAll(async ({ request }) => {
    const apiBase = (process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291')
      .replace(/\/$/, '')
    const res = await request.post(`${apiBase}/api/v1/auth/login`, {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    test.skip(res.status() !== 200, `admin.demo seed absent (HTTP ${res.status()})`)
  })

  test('D11c nav sidebar → /admin/afmps-imports', async ({ page }) => {
    test.setTimeout(60000)
    const pharmacyOn = process.env.NUXT_PUBLIC_PHARMACY_ENABLED
    test.skip(
      pharmacyOn === 'false' || pharmacyOn === '0',
      'NUXT_PUBLIC_PHARMACY_ENABLED off',
    )

    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin', { waitUntil: 'networkidle' })

    const nav = page.getByTestId('nav-admin-afmps-imports')
    await expect(nav).toBeVisible({ timeout: 15000 })
    await nav.click()

    await expect(page).toHaveURL(/\/admin\/afmps-imports\/?$/, { timeout: 15000 })
    await expect(page.getByTestId('admin-afmps-imports-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('afmps-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-new')).toBeVisible()

    await page.getByTestId('admin-afmps-new').click()
    await expect(page.getByTestId('admin-afmps-upload-card')).toBeVisible()
    await expect(page.getByTestId('admin-afmps-file')).toBeVisible()
  })
})
