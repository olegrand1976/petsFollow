import { test, expect, type Page } from '@playwright/test'
import { fillField, login } from '../helpers/auth'

const STAFF_PASSWORD = 'VetDemo123!'
/** Après 13-team-staff (InviteTeamMember + mustChangePassword), le seed peut déjà être changé. */
const STAFF_PASSWORD_AFTER = 'VetDemo123!Changed'

/** Remplit le MDP (ProInput contrôlé) + attend login BFF puis disparition de l'overlay. */
async function unlockWithPassword(page: Page, password = STAFF_PASSWORD) {
  await fillField(page, 'pro-desk-lock-password', password)
  const loginRes = page.waitForResponse(
    (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
    { timeout: 20000 },
  )
  await page.getByTestId('pro-desk-lock-submit').click()
  let status = (await loginRes).status()
  if (status === 401 && password === STAFF_PASSWORD) {
    await fillField(page, 'pro-desk-lock-password', STAFF_PASSWORD_AFTER)
    const retryRes = page.waitForResponse(
      (r) => r.url().includes('/api/auth/login') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('pro-desk-lock-submit').click()
    status = (await retryRes).status()
  }
  expect(status).toBe(200)
  // completeUnlock → location.replace : attendre le shell Pro rechargé.
  await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0, { timeout: 25000 })
  await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 25000 })
}

/** /clients ouvre ProModal si link-requests seedés — bloque les clics topbar. */
async function dismissProModals(page: Page) {
  const modal = page.getByTestId('pro-modal')
  for (let i = 0; i < 3; i++) {
    if ((await modal.count()) === 0) return
    const close = page.getByTestId('pro-modal-close')
    if ((await close.count()) > 0) {
      await close.first().click({ force: true })
    } else {
      await page.keyboard.press('Escape')
    }
    await expect(modal).toHaveCount(0, { timeout: 5000 }).catch(() => undefined)
  }
  await expect(modal).toHaveCount(0, { timeout: 5000 })
}

test.describe('desk switch — shared workstation', { tag: '@p0' }, () => {
  test('A: roster in header + force lock → unlock as secretary restores path', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/clients/, { timeout: 15000 })
    await dismissProModals(page)

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
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('')

    await page.getByTestId('pro-desk-lock-user-secretary.demo@petsfollow.test').click()
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('')
    await unlockWithPassword(page)
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
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('')
    await expect(page.getByTestId('pro-desk-lock-email')).toContainText('secretary.demo@petsfollow.test')

    await unlockWithPassword(page)

    // Switch back to vet.demo — should restore /calendar saved as lastPath.
    await page.getByTestId('pro-desk-user-vet.demo@petsfollow.test').click()
    await expect(page.getByTestId('pro-desk-lock')).toBeVisible({ timeout: 10000 })
    await unlockWithPassword(page)
    await expect(page).toHaveURL(/calendar/, { timeout: 20000 })
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })
  })

  test('C: switch purges session — cancel falls back to lock (no silent restore)', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    // Éviter /clients (modal invitations auto si link-requests seedés).
    await page.goto('/dashboard', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('pro-desk-switcher')).toBeVisible({ timeout: 15000 })
    await dismissProModals(page)

    await page.getByTestId('pro-desk-user-secretary.demo@petsfollow.test').click()
    const lock = page.getByTestId('pro-desk-lock')
    await expect(lock).toBeVisible({ timeout: 10000 })

    // Saisie partielle puis Annuler → veille : zone MDP toujours vidée.
    await fillField(page, 'pro-desk-lock-password', 'partial-secret')
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('partial-secret')

    // JWT httpOnly purged via BFF logout — pf_session marker must be gone.
    await expect.poll(async () => {
      return page.evaluate(() => document.cookie.split(';').some((c) => c.trim().startsWith('pf_session=')))
    }, { timeout: 10000 }).toBe(false)

    // Wait until BFF logout has cleared httpOnly JWT (pf_session alone is not enough).
    await expect.poll(async () => {
      const me = await page.request.get('/api/me')
      return me.status()
    }, { timeout: 15000 }).toBe(401)

    await page.getByTestId('pro-desk-lock-cancel').click()
    // Cancel after purge ≠ restore précédent : on reste en veille.
    await expect(lock).toBeVisible()
    await expect(page.getByTestId('pro-desk-lock-login')).toBeVisible({ timeout: 5000 })
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('')
    await expect(page.getByTestId('pro-topbar')).toHaveCount(0)
  })

  test('D: cancel switch immediately — no silent restore (race purge)', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/dashboard', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('pro-desk-switcher')).toBeVisible({ timeout: 15000 })
    await dismissProModals(page)

    await page.getByTestId('pro-desk-user-secretary.demo@petsfollow.test').click()
    const lock = page.getByTestId('pro-desk-lock')
    await expect(lock).toBeVisible({ timeout: 10000 })

    // Annuler immédiatement (purge logout peut encore tourner) → veille, pas restore.
    await page.getByTestId('pro-desk-lock-cancel').click()
    await expect(lock).toBeVisible()
    await expect(page.getByTestId('pro-desk-lock-login')).toBeVisible({ timeout: 5000 })
    await expect(page.getByTestId('pro-desk-lock-password')).toHaveValue('')
    await expect(page.getByTestId('pro-topbar')).toHaveCount(0)
  })

  test('E: solo roster — idle does not lock + switcher hidden', async ({ page }) => {
    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/dashboard', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('pro-desk-switcher')).toBeVisible({ timeout: 15000 })
    await dismissProModals(page)

    await page.evaluate(() => {
      const w = window as Window & {
        __PF_DESK_IDLE_MS?: number
        __PF_DESK_SET_ROSTER?: (m: Array<{ email: string; fullName: string; teamRole: string }>) => void
      }
      w.__PF_DESK_IDLE_MS = 400
      w.__PF_DESK_SET_ROSTER?.([
        { email: 'vet.demo@petsfollow.test', fullName: 'Vet Solo', teamRole: 'vet' },
      ])
    })

    await expect(page.getByTestId('pro-desk-switcher')).toHaveCount(0)
    await page.waitForTimeout(900)
    await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0)
    await expect(page.getByTestId('pro-topbar')).toBeVisible()
  })

  test('F: login via /login with stale pf_desk_locked — no immediate veille', async ({ page }) => {
    // Simule une veille précédente (flag sessionStorage) puis login classique :
    // ne doit PAS re-afficher l'overlay MDP juste après la 1re connexion.
    await page.goto('/login', { waitUntil: 'networkidle' })
    await page.evaluate(() => {
      sessionStorage.setItem('pf_desk_locked', '1')
    })

    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await expect(page.getByTestId('pro-desk-lock')).toHaveCount(0)
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 20000 })
    const stillLocked = await page.evaluate(() => sessionStorage.getItem('pf_desk_locked'))
    expect(stillLocked).toBeNull()
  })

  test('G: veille mid-consultation → unlock même user reprend le modal', async ({ page }) => {
    test.setTimeout(90_000)
    await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })
    await page.evaluate(() => {
      ;(window as any).__PF_DESK_IDLE_MS = 60_000
      const w = window as Window & {
        __PF_DESK_SET_ROSTER?: (m: Array<{ email: string, fullName: string, teamRole: string }>) => void
      }
      w.__PF_DESK_SET_ROSTER?.([
        { email: 'vet.demo@petsfollow.test', fullName: 'Vet Demo', teamRole: 'vet' },
        { email: 'vet.colleague@petsfollow.test', fullName: 'Colleague', teamRole: 'vet' },
      ])
    })

    await page.goto('/clients', { waitUntil: 'networkidle' })
    await dismissProModals(page)
    const search = page.getByPlaceholder(/nom ou email|name or email|naam of e-mail/i)
    await search.fill('Sophie')
    await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })
    const cta = page.locator('[data-testid^="new-consultation-"]').first()
    const petsRes = page.waitForResponse(
      (r) => /\/api\/clients\/[^/]+\/pets\b/.test(r.url()) && r.request().method() === 'GET',
      { timeout: 15000 },
    )
    await cta.click()
    await expect(page.getByTestId('consultation-modal')).toBeVisible({ timeout: 10000 })
    expect((await petsRes).status()).toBe(200)
    const petSelect = page.getByTestId('consultation-pet-select')
    await expect(petSelect).toBeEnabled({ timeout: 15000 })
    const options = petSelect.locator('option:not([disabled])')
    await options.first().waitFor({ state: 'attached', timeout: 15000 })
    const value = await options.first().getAttribute('value')
    expect(value).toBeTruthy()
    await petSelect.selectOption(value!)
    await page.getByTestId('consultation-start').click()
    await expect(page.getByTestId('consultation-report')).toBeVisible({ timeout: 15000 })
    await page.getByTestId('visit-report-body').fill(`Desk resume CR ${Date.now()}`)

    await page.evaluate(() => {
      const w = window as Window & { __PF_DESK_FORCE_LOCK?: () => void }
      w.__PF_DESK_FORCE_LOCK?.()
    })
    await expect(page.getByTestId('pro-desk-lock')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('consultation-modal')).toHaveCount(0)

    // Autosave may be in-flight before tokens cleared — wait for lock settle.
    await page.waitForTimeout(500)
    await unlockWithPassword(page)
    await expect(page.getByTestId('consultation-modal')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('consultation-report')).toBeVisible()
    // /clients réouvre la modale invitations (pro-modal) au-dessus du footer consultation.
    const invite = page.getByTestId('pro-modal')
    if (await invite.count()) {
      await invite.getByTestId('pro-modal-close').first().click({ force: true })
      await expect(invite).toHaveCount(0, { timeout: 5000 })
    }
    await page.getByTestId('consultation-cancel').click()
    const leavePrompt = page.getByTestId('consultation-leave-prompt')
    if (await leavePrompt.isVisible().catch(() => false)) {
      await page.getByTestId('consultation-leave-discard').click()
    }
    await expect(page.getByTestId('consultation-modal')).toHaveCount(0, { timeout: 10000 })
  })
})
