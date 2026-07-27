import { test, expect, type Page } from '@playwright/test'
import { login, fillField } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'
const STAFF_PASSWORD_AFTER = 'VetDemo123!Changed'

async function submitDeskUnlock(page: Page, password: string) {
  await fillField(page, 'pro-desk-lock-password', password)
  const loginRes = page.waitForResponse(
    (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('pro-desk-lock-submit').click()
  const res = await loginRes
  return res.status()
}

async function unlockWithStaffPassword(page: Page) {
  let status = await submitDeskUnlock(page, STAFF_PASSWORD)
  if (status === 401) {
    // Staff invited via seed may already have completed force-change in a prior test.
    await fillField(page, 'pro-desk-lock-password', STAFF_PASSWORD_AFTER)
    const loginRes = page.waitForResponse(
      (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('pro-desk-lock-submit').click()
    status = (await loginRes).status()
  }
  expect(status).toBe(200)
  await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0, { timeout: 25000 })
  await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })
}

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
    await unlockWithStaffPassword(page)
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

    await unlockWithStaffPassword(page)

    // Switch back to vet.demo — should restore /calendar saved as lastPath.
    await page.getByTestId('pro-desk-user-vet.demo@petsfollow.test').click()
    await expect(page.getByTestId('pro-desk-lock')).toBeVisible({ timeout: 10000 })
    await unlockWithStaffPassword(page)
    await expect(page).toHaveURL(/calendar/, { timeout: 20000 })
  })
})
