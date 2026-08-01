import { test, expect } from '@playwright/test'
import { login } from '../helpers/auth'

const RESEARCH_EMAIL = 'research.demo@petsfollow.test'
const RESEARCH_PASSWORD = 'ResearchDemo123!'
const ADMIN_EMAIL = 'admin.demo@petsfollow.test'
const ADMIN_PASSWORD = 'AdminDemo123!'

test.describe('Research observatory', { tag: '@p0' }, () => {
  test.beforeAll(async ({ request }) => {
    const apiBase = (process.env.PETSFOLLOW_API_URL || process.env.NUXT_PUBLIC_API_BASE || 'http://localhost:8291')
      .replace(/\/$/, '')
    const res = await request.post(`${apiBase}/api/v1/auth/login`, {
      data: { email: RESEARCH_EMAIL, password: RESEARCH_PASSWORD },
    })
    test.skip(res.status() !== 200, `research.demo seed absent (HTTP ${res.status()})`)
  })

  test('login research → overview + timeseries + heatmap', async ({ page }) => {
    test.setTimeout(90000)
    const researchOn = process.env.NUXT_PUBLIC_RESEARCH_ENABLED
    test.skip(
      researchOn === 'false' || researchOn === '0',
      'NUXT_PUBLIC_RESEARCH_ENABLED off',
    )

    await login(page, RESEARCH_EMAIL, RESEARCH_PASSWORD)
    await page.waitForURL((url) => url.pathname.startsWith('/research'), { timeout: 20000 })
    await page.goto('/research', { waitUntil: 'networkidle' })

    const disabled = page.getByTestId('research-disabled')
    if (await disabled.isVisible().catch(() => false)) {
      test.skip(true, 'Research public flag off in this build')
    }

    await expect(page.getByTestId('research-overview-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('research-dev-badge')).toBeVisible()
    await expect(page.getByTestId('research-overview-kpi')).toBeVisible({ timeout: 15000 })

    await page.goto('/research/timeseries', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('research-timeseries-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('research-timeseries-table')).toBeVisible({ timeout: 15000 })

    await page.goto('/research/heatmap', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('research-heatmap-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('research-heatmap-table')).toBeVisible({ timeout: 15000 })

    await page.goto('/research/groups', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('research-groups-page')).toBeVisible({ timeout: 15000 })

    await page.goto('/research/dataroom', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('research-dataroom-page')).toBeVisible({ timeout: 15000 })
    // Seed enables dataroom on demo group → table ; otherwise forbidden gate.
    const drTable = page.getByTestId('research-dataroom-table')
    const drGate = page.getByTestId('research-dataroom-group-required')
    await expect(drTable.or(drGate)).toBeVisible({ timeout: 15000 })
  })

  test('admin research opt-in list', async ({ page }) => {
    test.setTimeout(90000)
    const researchOn = process.env.NUXT_PUBLIC_RESEARCH_ENABLED
    test.skip(
      researchOn === 'false' || researchOn === '0',
      'NUXT_PUBLIC_RESEARCH_ENABLED off',
    )

    await login(page, ADMIN_EMAIL, ADMIN_PASSWORD)
    await page.goto('/admin/research', { waitUntil: 'networkidle' })
    if (await page.getByTestId('research-disabled').isVisible().catch(() => false)) {
      test.skip(true, 'Research public flag off in this build')
    }
    await expect(page.getByTestId('admin-research-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-research-dev-badge')).toBeVisible()
    await expect(page.getByTestId('admin-research-optins-table')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('admin-research-groups-table')).toBeVisible({ timeout: 15000 })

    const toggle = page.getByTestId('admin-research-dataroom-toggle').first()
    await expect(toggle).toBeVisible({ timeout: 15000 })
    const before = (await toggle.innerText()).trim()
    await toggle.click()
    await expect(toggle).not.toHaveText(before, { timeout: 15000 })
    // Restore seed-friendly state (toggle back).
    const mid = (await toggle.innerText()).trim()
    await toggle.click()
    await expect(toggle).not.toHaveText(mid, { timeout: 15000 })
  })
})
