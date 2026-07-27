import { test, expect } from '@playwright/test'
import { login } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'

test.describe('desk switch — shared workstation', { tag: '@p0' }, () => {
  test('A: roster in header + force lock → unlock as secretary restores path', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/clients/, { timeout: 15000 })

    const switcher = page.getByTestId('pro-desk-switcher')
    await expect(switcher).toBeVisible({ timeout: 15000 })
    await expect(
      page.getByTestId('pro-desk-user-secretary.demo@petsfollow.test'),
    ).toBeVisible({ timeout: 10000 })

    // Material Icons must not leak ligature text into the theme control.
    const themeBtn = page.getByTestId('pro-theme-toggle')
    await expect(themeBtn.locator('.pro-icon')).toHaveAttribute('translate', 'no')

    await page.evaluate(() => {
      ;(window as Window & { __PF_DESK_FORCE_LOCK?: () => void }).__PF_DESK_FORCE_LOCK?.()
    })

    const lock = page.getByTestId('pro-desk-lock')
    await expect(lock).toBeVisible({ timeout: 10000 })

    await page.getByTestId('pro-desk-lock-user-secretary.demo@petsfollow.test').click()
    await page.getByTestId('pro-desk-lock-password').fill(STAFF_PASSWORD)
    await Promise.all([
      page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 25000 }),
      page.getByTestId('pro-desk-lock-submit').click(),
    ])

    // First unlock for secretary may land on /dashboard (no prior lastPath) or a saved path.
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0)
  })

  test('B: switch from topbar asks password and lands on previous path', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/calendar', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/calendar/, { timeout: 15000 })

    await expect(page.getByTestId('pro-desk-switcher')).toBeVisible({ timeout: 15000 })
    await page.getByTestId('pro-desk-user-secretary.demo@petsfollow.test').click()

    const lock = page.getByTestId('pro-desk-lock')
    await expect(lock).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('pro-desk-lock-email')).toContainText('secretary.demo@petsfollow.test')
    await page.getByTestId('pro-desk-lock-password').fill(STAFF_PASSWORD)

    await Promise.all([
      page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 25000 }),
      page.getByTestId('pro-desk-lock-submit').click(),
    ])

    await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0)
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })

    // Switch back to vet.demo — should restore /calendar saved as lastPath.
    await page.getByTestId('pro-desk-user-vet.demo@petsfollow.test').click()
    await expect(page.getByTestId('pro-desk-lock')).toBeVisible({ timeout: 10000 })
    await page.getByTestId('pro-desk-lock-password').fill(STAFF_PASSWORD)
    await Promise.all([
      page.waitForURL(/calendar/, { timeout: 25000 }),
      page.getByTestId('pro-desk-lock-submit').click(),
    ])
    await expect(page).toHaveURL(/calendar/)
  })
})
