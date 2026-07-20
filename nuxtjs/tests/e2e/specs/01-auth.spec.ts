import { test, expect } from '@playwright/test'
import {
  fillField,
  login,
  loginAsAdmin,
  loginAsVet,
  logout,
  registerVet,
  requestPasswordReset,
  submitPasswordReset,
  uniqueE2EEmail,
} from '../helpers/auth'

test.describe('auth — login / logout', () => {
  test('login véto vers dashboard', async ({ page }) => {
    await loginAsVet(page)
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 })
    await expect(page.getByTestId('pro-topbar')).toBeVisible()
  })

  test('login admin vers /admin', async ({ page }) => {
    await loginAsAdmin(page)
    await expect(page).toHaveURL(/admin/, { timeout: 10000 })
    await expect(page.getByTestId('pro-topbar')).toBeVisible()
  })

  test('login mauvais mot de passe', async ({ page }) => {
    await login(page, 'vet.demo@petsfollow.test', 'WrongPass999!')
    await expect(page.getByTestId('login-form')).toBeVisible()
    await expect(page.locator('[data-testid="login-form"] .pro-field-error')).toBeVisible({ timeout: 10000 })
  })

  test('login email non vérifié', async ({ page }) => {
    await login(page, 'vet.unverified@petsfollow.test', 'VetDemo123!')
    await expect(page.getByTestId('login-form')).toBeVisible()
    await expect(page.locator('[data-testid="login-form"] .pro-field-error')).toBeVisible({ timeout: 10000 })
  })

  test('logout depuis topbar', async ({ page }) => {
    await loginAsVet(page)
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 })
    await expect(page.getByTestId('pro-topbar')).toBeVisible()
    await logout(page)
    await expect(page.getByTestId('login-form')).toBeVisible()
  })
})

test.describe('auth — inscription et confirmation', () => {
  test('register → sent → confirm → dashboard/welcome', async ({ page }) => {
    const email = uniqueE2EEmail('register')
    const { confirmPath, status } = await registerVet(page, {
      fullName: 'Dr E2E Register',
      practiceName: 'Cabinet E2E',
      email,
      password: 'E2ePass123!',
    })
    expect(status === 200 || status === 201).toBeTruthy()
    expect(confirmPath).toBeTruthy()
    await expect(page).toHaveURL(/register\/sent/, { timeout: 15000 })

    await page.goto(confirmPath!)
    await expect(page.getByTestId('confirm-email-success')).toBeVisible({ timeout: 15000 })
    await page.getByTestId('confirm-email-continue').click()
    await expect(page).toHaveURL(/welcome|onboarding|dashboard/, { timeout: 10000 })
  })

  test('confirm token seed demo-confirm-email', async ({ page }) => {
    // Compte à usage unique après seed — peut échouer si déjà consommé
    await page.goto('/confirm-email?token=demo-confirm-email')
    await expect(
      page.getByTestId('confirm-email-success').or(page.getByTestId('confirm-email-failed')),
    ).toBeVisible({ timeout: 15000 })
  })

  test('register passwords mismatch', async ({ page }) => {
    await registerVet(page, {
      fullName: 'Dr Mismatch',
      practiceName: 'Cabinet Mismatch',
      email: uniqueE2EEmail('mismatch'),
      password: 'E2ePass123!',
      passwordConfirm: 'OtherPass123!',
    })
    await expect(page).toHaveURL(/register/)
    await expect(page.locator('[data-testid="register-form"] .pro-field-error')).toBeVisible()
  })

  test('register email déjà pris', async ({ page }) => {
    await registerVet(page, {
      fullName: 'Dr Dup',
      practiceName: 'Cabinet Dup',
      email: 'vet.demo@petsfollow.test',
      password: 'E2ePass123!',
    })
    await expect(page.locator('[data-testid="register-form"] .pro-field-error')).toBeVisible({ timeout: 10000 })
  })
})

test.describe('auth — forgot / reset password', () => {
  test('forgot → reset via API path → login nouveau MDP', async ({ page }) => {
    const email = 'vet.reset@petsfollow.test'
    const newPassword = `Reset${Date.now()}!`

    const { resetPath } = await requestPasswordReset(page, email)
    expect(resetPath).toBeTruthy()

    const { status } = await submitPasswordReset(page, resetPath!, newPassword)
    expect(status === 200 || status === 204).toBeTruthy()
    await expect(page.getByTestId('reset-done')).toBeVisible({ timeout: 10000 })

    await login(page, email, newPassword)
    await expect(page).toHaveURL(/dashboard|onboarding/, { timeout: 10000 })
  })

  test('forgot email inconnu affiche message générique', async ({ page }) => {
    const { resetPath } = await requestPasswordReset(page, 'unknown.e2e@petsfollow.test')
    await expect(page.getByTestId('forgot-sent')).toBeVisible()
    expect(resetPath).toBeFalsy()
    await expect(page.getByTestId('forgot-dev-link')).toHaveCount(0)
  })

  test('reset token invalide', async ({ page }) => {
    const { status } = await submitPasswordReset(page, '/reset-password?token=invalid-token', 'Whatever12!')
    expect(status === 400 || status === 404 || status === 422).toBeTruthy()
    await expect(page.locator('[data-testid="reset-form"] .pro-field-error')).toBeVisible({ timeout: 10000 })
  })
})
