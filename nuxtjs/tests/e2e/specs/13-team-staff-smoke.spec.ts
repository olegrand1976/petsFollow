import { test, expect, type Page } from '@playwright/test'
import { login, fillField, waitForAuthForm } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'
/** Mot de passe post-invite (mustChangePassword) — stable pour rejouer le smoke. */
const STAFF_PASSWORD_AFTER = 'VetDemo123!Changed'

async function completeForcedPasswordChange(page: Page, nextPassword = STAFF_PASSWORD_AFTER) {
  await waitForAuthForm(page, 'force-change-password-form')
  await fillField(page, 'force-change-password', nextPassword)
  await fillField(page, 'force-change-password-confirm', nextPassword)
  await page.getByTestId('force-change-submit').click()
  await page.waitForURL((url) => !url.pathname.includes('/change-password'), { timeout: 20000 })
}

async function loginExpectDashboard(page: Page, email: string, password = STAFF_PASSWORD) {
  const { status } = await login(page, email, password)
  // Staff invited via seed may already have completed force-change in a prior test.
  if (status === 401) {
    const retry = await login(page, email, STAFF_PASSWORD_AFTER)
    if (retry.status !== 200) {
      const bodyText = await page.locator('body').innerText().catch(() => '')
      if (/proOnly|réservé aux profils Pro|reserved for Pro/i.test(bodyText)) {
        throw new Error(`KO: login ${email} failed with proOnly (status=${retry.status})`)
      }
      throw new Error(`KO: login ${email} failed with status=${retry.status} (tried temp + changed pwd)`)
    }
  } else if (status !== 200) {
    const bodyText = await page.locator('body').innerText().catch(() => '')
    if (/proOnly|réservé aux profils Pro|reserved for Pro/i.test(bodyText)) {
      throw new Error(`KO: login ${email} failed with proOnly (status=${status})`)
    }
    throw new Error(`KO: login ${email} failed with status=${status}`)
  }

  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
  if (page.url().includes('/change-password')) {
    await completeForcedPasswordChange(page)
  }
  await expect(page).toHaveURL(/dashboard/, { timeout: 15000 })
}

test.describe('team staff smoke — assist / secretary / reference', () => {
  test('A: assist login → /dashboard with nav', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.assist@petsfollow.test')
    const nav = page
      .locator('a[href="/clients"], a[href="/dashboard"]')
      .or(page.getByTestId('pro-topbar'))
    await expect(nav.first()).toBeVisible({ timeout: 10000 })
  })

  test('B: secretary login → /dashboard', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
  })

  test('C: assist — /commissions redirected, /team OK', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.assist@petsfollow.test')

    await page.goto('/commissions', { waitUntil: 'domcontentloaded' })
    await page.waitForURL((url) => !url.pathname.includes('/commissions'), { timeout: 15000 })
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 })
    await expect(page.getByTestId('vet-commissions-page')).toHaveCount(0)

    await page.goto('/team', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('vet-team-page')).toBeVisible({ timeout: 15000 })
    await expect(
      page.getByRole('heading').or(page.locator('table, thead, .pro-table')).first(),
    ).toBeVisible({ timeout: 10000 })
  })

  test('D: vet.demo — /team with invite UI', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await page.goto('/team', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('vet-team-page')).toBeVisible({ timeout: 15000 })
    await expect(page.locator('form.pro-form')).toBeVisible({ timeout: 10000 })
    await expect(page.getByRole('button', { name: /invit/i })).toBeVisible({ timeout: 10000 })
  })

  test('E: vet.demo — /commissions allowed for reference', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await page.goto('/commissions', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/commissions/)
    await expect(page.getByTestId('vet-commissions-page')).toBeVisible({ timeout: 15000 })
  })
})
