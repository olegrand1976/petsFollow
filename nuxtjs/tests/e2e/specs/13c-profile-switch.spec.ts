import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../helpers/auth'

test.describe('multi-profil topbar switcher', { tag: '@p1' }, () => {
  test('admin.demo — liste verticale + profil secrétaire', async ({ page }) => {
    const { status } = await loginAsAdmin(page)
    test.skip(status !== 200, `admin.demo seed absent (HTTP ${status})`)

    await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 15000 })

    await page.getByTestId('pro-profile-btn').click()
    const switcher = page.getByTestId('pro-profile-switcher')
    await expect(switcher).toBeVisible({ timeout: 10000 })

    const secretary = page.getByTestId('pro-profile-switch-secretary')
    await expect(secretary).toBeVisible()

    const buttons = switcher.locator('button.pro-topbar__profile-switch')
    const count = await buttons.count()
    expect(count).toBeGreaterThanOrEqual(3)

    const box0 = await buttons.nth(0).boundingBox()
    const box1 = await buttons.nth(1).boundingBox()
    expect(box0 && box1).toBeTruthy()
    if (box0 && box1) {
      expect(box1.y).toBeGreaterThan(box0.y + box0.height / 2)
    }
  })
})
