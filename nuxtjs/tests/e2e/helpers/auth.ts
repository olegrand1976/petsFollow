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
  // Cloud Run cold start: wait for JS hydration before submit/click handlers bind.
  await page.waitForLoadState('networkidle', { timeout: 15000 }).catch(() => {})
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

export async function login(
  page: Page,
  email: string,
  password: string,
  opts?: { expectStatus?: number },
): Promise<{ status?: number }> {
  await page.context().clearCookies()
  await page.goto('/login', { waitUntil: 'networkidle' })
  await waitForAuthForm(page, 'login-form')
  await fillField(page, 'login-email', email)
  await fillField(page, 'login-password', password)
  await expect(page.getByTestId('login-email')).toHaveValue(email)
  await expect(page.getByTestId('login-password')).toHaveValue(password)

  const responsePromise = page.waitForResponse(
    (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('login-submit').click()
  const res = await responsePromise
  if (opts?.expectStatus != null) {
    expect(res.status()).toBe(opts.expectStatus)
  }
  // Multi-profil démo (research attaché) : un smoke peut laisser active_profile=research
  // → clients 403 / pas de /dashboard. Réactive le rôle primaire attendu.
  if (res.status() === 200) {
    const primary = demoPrimaryRole(email)
    if (primary) await ensureActiveRole(page, primary)
  }
  return { status: res.status() }
}

function demoPrimaryRole(email: string): string | null {
  const e = email.trim().toLowerCase()
  if (e === 'vet.demo@petsfollow.test') return 'vet'
  if (e === 'admin.demo@petsfollow.test') return 'admin'
  if (e === 'dev.demo@petsfollow.test') return 'dev'
  return null
}

/** Si un smoke a laissé active_profile=research, bascule vers le rôle attendu (évite clients 403). */
async function ensureActiveRole(page: Page, role: string) {
  const me = await page.request.get('/api/me')
  if (!me.ok()) return
  const meBody = unwrapData(await me.json())
  if (meBody.role === role) return
  const profilesRes = await page.request.get('/api/me/profiles')
  if (!profilesRes.ok()) return
  const body = await profilesRes.json()
  const list = Array.isArray(body) ? body : ((body as { data?: Array<{ id: string; role: string }> }).data ?? [])
  const target = list.find((p: { id: string; role: string }) => p.role === role)
  if (!target?.id) return
  await page.request.post('/api/me/profiles/switch', { data: { profileId: target.id } })
}

export async function loginAsVet(page: Page, email = 'vet.demo@petsfollow.test', password = 'VetDemo123!') {
  await login(page, email, password)
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
}

export async function loginAsAdmin(page: Page, email = 'admin.demo@petsfollow.test', password = 'AdminDemo123!') {
  await login(page, email, password)
  await page.waitForURL(/\/admin/, { timeout: 20000 })
}

/** Ops support IT (même MDP seed que admin — passwordDev = passwordAdmin). */
export async function loginAsDev(page: Page, email = 'dev.demo@petsfollow.test', password = 'AdminDemo123!') {
  const { status } = await login(page, email, password)
  expect(status, `login ${email}`).toBe(200)
  await page.waitForURL(/\/admin/, { timeout: 20000 })
}

/** Gate commercial : si contact_phone vide (staging sans re-seed), complète le numéro démo. */
async function completeContactPhoneIfNeeded(page: Page, phone: string) {
  if (!page.url().includes('/complete-contact-phone')) return
  await waitForAuthForm(page, 'complete-contact-phone-form')
  await fillField(page, 'complete-contact-phone', phone)
  await Promise.all([
    page.waitForURL((url) => !url.pathname.includes('/complete-contact-phone'), { timeout: 20000 }),
    page.getByTestId('complete-contact-phone-submit').click(),
  ])
}

export async function loginAsCommercial(
  page: Page,
  email = 'commercial.demo@petsfollow.test',
  password = 'CommercialDemo123!',
) {
  await login(page, email, password)
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
  await completeContactPhoneIfNeeded(page, '0470 12 34 56')
  await page.waitForURL((url) => /^\/commercial(?:\/|$)/.test(url.pathname), { timeout: 20000 })
}

export async function loginAsCommercialManager(
  page: Page,
  email = 'commercial.manager@petsfollow.test',
  password = 'CommercialDemo123!',
) {
  await login(page, email, password)
  await page.waitForURL((url) => !url.pathname.includes('/login'), { timeout: 20000 })
  await completeContactPhoneIfNeeded(page, '0472 11 22 33')
  await page.waitForURL(/\/commercial-manager/, { timeout: 20000 })
}

/** Clique natif (fiable avec handlers Vue / overlays). */
export async function nativeClick(page: Page, testId: string) {
  const el = page.getByTestId(testId)
  await expect(el).toBeVisible()
  await el.evaluate((node) => (node as HTMLElement).click())
}

export async function logout(page: Page) {
  await expect(page.getByTestId('pro-profile-btn')).toBeVisible({ timeout: 15000 })
  // Native <details>/<summary> — open attribute is the source of truth.
  await page.getByTestId('pro-profile-btn').click()
  await expect(page.getByTestId('pro-profile-details')).toHaveAttribute('open', /.*/, { timeout: 10000 })
  await expect(page.getByTestId('pro-logout-btn')).toBeVisible({ timeout: 10000 })
  // Full-document logout-redirect (clears httpOnly cookies server-side).
  await Promise.all([
    page.waitForURL(/\/login/, { timeout: 20000 }),
    page.getByTestId('pro-logout-btn').click(),
  ])
  await waitForAuthForm(page, 'login-form')
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
  await page.getByTestId('register-consent').check()

  const expectApi = confirm === input.password
  const responsePromise = expectApi
    ? page.waitForResponse(
      (r) => r.url().includes('/api/auth/register') && r.request().method() === 'POST',
      { timeout: 30000 },
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

/** Remplit et soumet le formulaire reset (avec attente d’hydratation). */
export async function submitPasswordReset(
  page: Page,
  resetPath: string,
  password: string,
): Promise<{ status?: number }> {
  await page.goto(resetPath, { waitUntil: 'networkidle' })
  await waitForAuthForm(page, 'reset-form')
  await fillField(page, 'reset-password', password)
  await fillField(page, 'reset-password-confirm', password)

  const responsePromise = page.waitForResponse(
    (r) => r.url().includes('/api/auth/reset-password') && r.request().method() === 'POST',
    { timeout: 15000 },
  ).catch(() => null)

  await page.getByTestId('reset-submit').click()
  const res = await responsePromise
  return { status: res?.status() }
}

export function uniqueE2EEmail(prefix = 'e2e') {
  return `${prefix}+${Date.now()}@petsfollow.test`
}
