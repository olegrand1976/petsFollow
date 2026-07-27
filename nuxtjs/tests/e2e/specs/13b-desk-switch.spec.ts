import { test, expect, type Page } from '@playwright/test'
import { fillField, login } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'

/** Remplit le MDP (ProInput contrôlé) + attend login BFF puis disparition de l'overlay. */
async function unlockWithPassword(page: Page, password: string) {
  await fillField(page, 'pro-desk-lock-password', password)
  const loginRes = page.waitForResponse(
    (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('pro-desk-lock-submit').click()
  const res = await loginRes
  expect(res.status()).toBe(200)
  // completeUnlock → location.replace : attendre le shell Pro rechargé.
  await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0, { timeout: 25000 })
  await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 25000 })
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
    await unlockWithPassword(page, STAFF_PASSWORD)
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

    await unlockWithPassword(page, STAFF_PASSWORD)

    // Switch back to vet.demo — should restore /calendar saved as lastPath.
    await page.getByTestId('pro-desk-user-vet.demo@petsfollow.test').click()
    await expect(page.getByTestId('pro-desk-lock')).toBeVisible({ timeout: 10000 })
    await fillField(page, 'pro-desk-lock-password', STAFF_PASSWORD)
    const loginRes = page.waitForResponse(
      (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await Promise.all([
      page.waitForURL(/calendar/, { timeout: 25000 }),
      page.getByTestId('pro-desk-lock-submit').click(),
    ])
    expect((await loginRes).status()).toBe(200)
    await expect(page).toHaveURL(/calendar/)
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })
  })

  test('C: switch purges session — cancel falls back to lock (no silent restore)', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('pro-desk-switcher')).toBeVisible({ timeout: 15000 })

    await page.getByTestId('pro-desk-user-secretary.demo@petsfollow.test').click()
    const lock = page.getByTestId('pro-desk-lock')
    await expect(lock).toBeVisible({ timeout: 10000 })

    // JWT httpOnly purged via BFF logout — pf_session marker must be gone.
    await expect.poll(async () => {
      return page.evaluate(() => document.cookie.split(';').some((c) => c.trim().startsWith('pf_session=')))
    }, { timeout: 10000 }).toBe(false)

    const me = await page.request.get('/api/me')
    expect(me.status()).toBe(401)

    await page.getByTestId('pro-desk-lock-cancel').click()
    // Cancel after purge ≠ restore précédent : on reste en veille.
    await expect(lock).toBeVisible()
    await expect(page.getByTestId('pro-desk-lock-login')).toBeVisible({ timeout: 5000 })
    await expect(page.getByTestId('pro-topbar')).toHaveCount(0)
  })
})
