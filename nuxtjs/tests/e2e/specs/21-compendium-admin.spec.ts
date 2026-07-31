import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../helpers/auth'

/**
 * Compendium PDF admin — smoke navigation (middleware admin-only).
 * Tag @p1 : page list + lien sidebar ; pas d’extract Gemini live.
 */
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

test.describe('Compendium admin imports', { tag: ['@p1', '@pharmacy'] }, () => {
  test.beforeAll(async ({ request }) => {
    const apiBase = (process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291')
      .replace(/\/$/, '')
    const res = await request.post(`${apiBase}/api/v1/auth/login`, {
      data: { email: ADMIN_EMAIL, password: ADMIN_PASSWORD },
    })
    test.skip(res.status() !== 200, `admin.demo seed absent (HTTP ${res.status()})`)
  })

  test('D11b nav sidebar → /admin/compendium-imports', async ({ page }) => {
    test.setTimeout(60000)
    await loginAsAdmin(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin', { waitUntil: 'networkidle' })

    const nav = page.getByTestId('nav-admin-compendium-imports')
    await expect(nav).toBeVisible({ timeout: 15000 })
    await nav.click()

    await expect(page).toHaveURL(/\/admin\/compendium-imports\/?$/, { timeout: 15000 })
    await expect(page.getByTestId('admin-compendium-imports-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('compendium-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-compendium-new')).toBeVisible()
  })
})
