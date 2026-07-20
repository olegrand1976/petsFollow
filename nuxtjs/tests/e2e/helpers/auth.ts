import type { Page, Response } from '@playwright/test'
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

/** Attend que le formulaire Vue soit hydraté (évite un submit HTML GET natif). */
export async function waitForAuthForm(page: Page, testId: string) {
  await expect(page.getByTestId(testId)).toBeVisible()
  await page.waitForFunction((id) => {
    const form = document.querySelector(`[data-testid="${id}"]`)
    return !!form && !!document.querySelector('#__nuxt')
  }, testId)
}

function unwrapData(body: unknown): Record<string, unknown> {
  if (!body || typeof body !== 'object') return {}
  const obj = body as { data?: Record<string, unknown> }
  return (obj.data && typeof obj.data === 'object' ? obj.data : obj) as Record<string, unknown>
}

async function jsonFromResponse(res: Response | null): Promise<Record<string, unknown>> {
  if (!res) return {}
  try {
    return unwrapData(await res.json())
  } catch {
    return {}
  }
}

export async function login(page: Page, email: string, password: string) {
  await page.context().clearCookies()
  await page.goto('/login', { waitUntil: 'networkidle' })
  await waitForAuthForm(page, 'login-form')
  await fillField(page, 'login-email', email)
  await fillField(page, 'login-password', password)
  await expect(page.getByTestId('login-email')).toHaveValue(email)
  await expect(page.getByTestId('login-password')).toHaveValue(password)
  await page.getByTestId('login-submit').click()
}

export async function loginAsVet(page: Page, email = 'vet.demo@petsfollow.test', password = 'VetDemo123!') {
  await login(page, email, password)
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
}

export async function loginAsAdmin(page: Page, email = 'admin.demo@petsfollow.test', password = 'AdminDemo123!') {
  await login(page, email, password)
  await page.waitForURL(/\/admin/, { timeout: 20000 })
}

export async function loginAsCommercial(
  page: Page,
  email = 'commercial.demo@petsfollow.test',
  password = 'CommercialDemo123!',
) {
  await login(page, email, password)
  await page.waitForURL((url) => /^\/commercial(?:\/|$)/.test(url.pathname), { timeout: 20000 })
}

export async function loginAsCommercialManager(
  page: Page,
  email = 'commercial.manager@petsfollow.test',
  password = 'CommercialDemo123!',
) {
  await login(page, email, password)
  await page.waitForURL(/\/commercial-manager/, { timeout: 20000 })
}

export async function logout(page: Page) {
  await page.getByTestId('pro-profile-btn').click()
  await page.getByTestId('pro-logout-btn').click()
  await expect(page).toHaveURL(/login/)
}

export async function registerVet(
  page: Page,
  input: { fullName: string; practiceName: string; email: string; password: string; passwordConfirm?: string },
): Promise<{ confirmPath?: string; status?: number }> {
  await page.goto('/register', { waitUntil: 'networkidle' })
  await waitForAuthForm(page, 'register-form')
  await fillField(page, 'register-fullname', input.fullName)
  await fillField(page, 'register-practice', input.practiceName)
  await fillField(page, 'register-email', input.email)
  const confirm = input.passwordConfirm ?? input.password
  await fillField(page, 'register-password', input.password)
  await fillField(page, 'register-password-confirm', confirm)

  const expectApi = confirm === input.password
  const responsePromise = expectApi
    ? page.waitForResponse(
      (r) => r.url().includes('/api/auth/register') && r.request().method() === 'POST',
      { timeout: 15000 },
    )
    : Promise.resolve(null)

  await page.getByTestId('register-submit').click()
  const res = await responsePromise
  const data = await jsonFromResponse(res)
  const confirmPath = typeof data.confirmPath === 'string' ? data.confirmPath : undefined
  return { confirmPath, status: res?.status() }
}

export async function confirmEmail(page: Page, confirmPath: string) {
  await page.goto(confirmPath)
  await expect(page.getByTestId('confirm-email-success').or(page.getByTestId('confirm-email-failed'))).toBeVisible({
    timeout: 15000,
  })
}

export async function requestPasswordReset(
  page: Page,
  email: string,
): Promise<{ resetPath?: string }> {
  await page.goto('/forgot-password', { waitUntil: 'networkidle' })
  await waitForAuthForm(page, 'forgot-form')
  await fillField(page, 'forgot-email', email)

  const responsePromise = page.waitForResponse(
    (r) => r.url().includes('/api/auth/forgot-password') && r.request().method() === 'POST',
    { timeout: 15000 },
  )

  await page.getByTestId('forgot-submit').click()
  const res = await responsePromise
  await expect(page.getByTestId('forgot-sent')).toBeVisible({ timeout: 10000 })
  const data = await jsonFromResponse(res)
  const resetPath = typeof data.resetPath === 'string' ? data.resetPath : undefined
  return { resetPath }
}

export function uniqueE2EEmail(prefix = 'e2e') {
  return `${prefix}+${Date.now()}@petsfollow.test`
}
