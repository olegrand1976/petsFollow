import type { Page } from '@playwright/test'
import { expect } from '@playwright/test'

/** Remplit un input Pro (contrôlé Vue) et vérifie la valeur. */
export async function fillField(page: Page, testId: string, value: string) {
  const el = page.getByTestId(testId)
  await expect(el).toBeVisible()
  await el.click()
  await el.fill(value)
  if ((await el.inputValue()) !== value) {
    await el.evaluate((node, v) => {
      const input = node as HTMLInputElement
      const proto = window.HTMLInputElement.prototype
      const desc = Object.getOwnPropertyDescriptor(proto, 'value')
      desc?.set?.call(input, v)
      input.dispatchEvent(new Event('input', { bubbles: true }))
      input.dispatchEvent(new Event('change', { bubbles: true }))
    }, value)
  }
  await expect(el).toHaveValue(value)
}

export async function login(page: Page, email: string, password: string) {
  await page.context().clearCookies()
  await page.goto('/login')
  await expect(page.getByTestId('login-form')).toBeVisible()
  await fillField(page, 'login-email', email)
  await fillField(page, 'login-password', password)
  await page.getByTestId('login-submit').click()
}

export async function loginAsVet(page: Page, email = 'vet.demo@petsfollow.test', password = 'VetDemo123!') {
  await login(page, email, password)
}

export async function loginAsAdmin(page: Page, email = 'admin.demo@petsfollow.test', password = 'AdminDemo123!') {
  await login(page, email, password)
}

export async function loginAsCommercial(
  page: Page,
  email = 'commercial.demo@petsfollow.test',
  password = 'CommercialDemo123!',
) {
  await login(page, email, password)
}

export async function logout(page: Page) {
  await page.getByTestId('pro-profile-btn').click()
  await page.getByTestId('pro-logout-btn').click()
  await expect(page).toHaveURL(/login/)
}

export async function registerVet(
  page: Page,
  input: { fullName: string; practiceName: string; email: string; password: string; passwordConfirm?: string },
) {
  await page.goto('/register')
  await expect(page.getByTestId('register-form')).toBeVisible()
  await fillField(page, 'register-fullname', input.fullName)
  await fillField(page, 'register-practice', input.practiceName)
  await fillField(page, 'register-email', input.email)
  await fillField(page, 'register-password', input.password)
  await fillField(page, 'register-password-confirm', input.passwordConfirm ?? input.password)
  await page.getByTestId('register-submit').click()
}

export async function confirmEmail(page: Page, confirmPath: string) {
  await page.goto(confirmPath)
  await expect(page.getByTestId('confirm-email-success').or(page.getByTestId('confirm-email-failed'))).toBeVisible({
    timeout: 15000,
  })
}

export async function requestPasswordReset(page: Page, email: string) {
  await page.goto('/forgot-password')
  await expect(page.getByTestId('forgot-form')).toBeVisible()
  await fillField(page, 'forgot-email', email)
  await page.getByTestId('forgot-submit').click()
  await expect(page.getByTestId('forgot-sent')).toBeVisible({ timeout: 10000 })
}

export function uniqueE2EEmail(prefix = 'e2e') {
  return `${prefix}+${Date.now()}@petsfollow.test`
}
