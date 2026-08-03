import { test, expect, type Page } from '@playwright/test'
import { login, fillField, waitForAuthForm } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

const STAFF_PASSWORD = 'VetDemo123!'
/** Mot de passe post-invite (mustChangePassword) — stable pour rejouer le smoke. */
const STAFF_PASSWORD_AFTER = 'VetDemo123!Changed'

async function completeForcedPasswordChange(page: Page, nextPassword = STAFF_PASSWORD_AFTER) {
  await waitForAuthForm(page, 'force-change-password-form')
  await fillField(page, 'force-change-password', nextPassword)
  await fillField(page, 'force-change-password-confirm', nextPassword)
  await expect(page.getByTestId('force-change-password')).toHaveValue(nextPassword)
  await expect(page.getByTestId('force-change-password-confirm')).toHaveValue(nextPassword)
  const patch = page.waitForResponse(
    (r) => r.url().includes('/api/me/password') && r.request().method() === 'PATCH',
    { timeout: 25000 },
  )
  await page.getByTestId('force-change-password-form').evaluate((node) => {
    const form = node as HTMLFormElement
    if (typeof form.requestSubmit === 'function') form.requestSubmit()
    else form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
  })
  const res = await patch
  if (!res.ok()) {
    const errText = await page.locator('.pro-field-error').innerText().catch(() => '')
    throw new Error(`force-change password PATCH ${res.status()} ${errText}`)
  }
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

test.describe('team staff smoke — assist / secretary / reference', { tag: '@p0' }, () => {
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

  test('B2: secretary — nav sans commissions, sections présentes', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await expect(page.getByTestId('nav-commissions')).toHaveCount(0)
    await expect(page.locator('a[href="/commissions"]')).toHaveCount(0)
    await expect(page.getByText(/^Journée$|^Daily$/i).first()).toBeVisible({ timeout: 10000 })
    await expect(page.getByText(/^Cabinet$|^Practice$/i).first()).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('nav-clients')).toBeVisible({ timeout: 10000 })
  })

  test('B2b: secretary — agenda OK, pas d’historique consultations', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await expect(page.getByTestId('nav-calendar')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('nav-consultations')).toHaveCount(0)
    await expect(page.locator('a[href="/consultations"]')).toHaveCount(0)
    await page.goto('/consultations', { waitUntil: 'domcontentloaded' })
    await page.waitForURL((url) => !url.pathname.includes('/consultations'), { timeout: 15000 })
    await expect(page.getByTestId('consultations-page')).toHaveCount(0)
  })

  test('B2c: secretary — détail RDV desk (pas de CR clinique)', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/calendar', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('calendar-page')).toBeVisible({ timeout: 15000 })
    const chip = page.locator('[data-testid^="calendar-chip-"]:not(.cal-chip--walkin)').first()
    if ((await chip.count()) === 0) {
      test.skip(true, 'aucun RDV non walk-in visible sur l’agenda seed pour le smoke desk')
      return
    }
    await chip.click()
    await expect(page.getByTestId('calendar-desk-note')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('calendar-desk-note-input')).toBeVisible()
    await expect(page.getByTestId('visit-address')).toHaveCount(0)
    await expect(page.getByTestId('calendar-delete-visit')).toBeVisible()
    await expect(
      page.getByTestId('calendar-waiting-room-on').or(page.getByTestId('calendar-waiting-room-off')),
    ).toBeVisible()
    // Pas de CR clinique ni de CTA consultation pour la secrétaire.
    await expect(page.getByTestId('visit-report-panel')).toHaveCount(0)
    await expect(page.getByTestId('calendar-open-consultation')).toHaveCount(0)
    await expect(page.getByTestId('calendar-view-consultation')).toHaveCount(0)
  })

  test('B2d: vet — détail RDV desk (mêmes options que secrétaire + CTA consultation)', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await page.goto('/calendar', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('calendar-page')).toBeVisible({ timeout: 15000 })
    const chip = page
      .locator('[data-testid^="calendar-chip-"].cal-chip--success:not(.cal-chip--walkin)')
      .first()
    if ((await chip.count()) === 0) {
      test.skip(true, 'aucun RDV confirmé non walk-in visible sur l’agenda seed pour le smoke desk')
      return
    }
    await chip.click()
    // Parité desk avec la secrétaire : note, modifier l'heure, supprimer, salle d'attente.
    await expect(page.getByTestId('calendar-desk-note')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('calendar-change-time')).toBeVisible()
    await expect(page.getByTestId('calendar-delete-visit')).toBeVisible()
    await expect(
      page.getByTestId('calendar-waiting-room-on').or(page.getByTestId('calendar-waiting-room-off')),
    ).toBeVisible()
    // Les données de consultation ne sont plus dans le détail RDV — CTA vers l'écran consultation.
    await expect(page.getByTestId('visit-report-panel')).toHaveCount(0)
    await expect(
      page.getByTestId('calendar-open-consultation').or(page.getByTestId('calendar-view-consultation')),
    ).toBeVisible()
  })

  test('B2e: assist — détail RDV desk (mêmes options que secrétaire + CTA consultation)', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.assist@petsfollow.test')
    await page.goto('/calendar', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('calendar-page')).toBeVisible({ timeout: 15000 })
    const chip = page
      .locator('[data-testid^="calendar-chip-"].cal-chip--success:not(.cal-chip--walkin)')
      .first()
    if ((await chip.count()) === 0) {
      test.skip(true, 'aucun RDV confirmé non walk-in visible sur l’agenda seed pour le smoke desk')
      return
    }
    await chip.click()
    // Parité desk avec la secrétaire : note, modifier l'heure, supprimer, salle d'attente.
    await expect(page.getByTestId('calendar-desk-note')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('calendar-change-time')).toBeVisible()
    await expect(page.getByTestId('calendar-delete-visit')).toBeVisible()
    await expect(
      page.getByTestId('calendar-waiting-room-on').or(page.getByTestId('calendar-waiting-room-off')),
    ).toBeVisible()
    // Les données de consultation ne sont plus dans le détail RDV — CTA vers l'écran consultation.
    await expect(page.getByTestId('visit-report-panel')).toHaveCount(0)
    await expect(
      page.getByTestId('calendar-open-consultation').or(page.getByTestId('calendar-view-consultation')),
    ).toBeVisible()
  })

  test('B3: secretary — /prescriptions/nouveau redirect (no write_clinical)', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/prescriptions/nouveau', { waitUntil: 'domcontentloaded' })
    await page.waitForURL((url) => !url.pathname.includes('/prescriptions/nouveau'), { timeout: 15000 })
    await expect(page).toHaveURL(/dashboard/, { timeout: 10000 })
  })

  test('B4: secretary — pas de CTA nouveau prescriptions', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/prescriptions', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('prescriptions-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('prescriptions-new')).toHaveCount(0)
  })

  test('B5: secretary — pas de CTA consultation (no write_clinical)', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('create-client-open')).toBeVisible({ timeout: 10000 })
    // Seed VetPlus : au moins un client — sinon le count CTA serait vrai à vide.
    await expect(page.locator('table tbody tr, .pro-kanban-card').first()).toBeVisible({ timeout: 15000 })
    await expect(page.locator('[data-testid^="new-consultation-"]')).toHaveCount(0)
  })

  test('B6: secretary — settings sans fiche cabinet, sans invitations', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/settings', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('settings-practice-profile-denied')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('settings-practice-profile')).toHaveCount(0)
    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('clients-invitations-open')).toHaveCount(0)
  })

  test('B7: secretary — factu sans activate Billit (pas practice.settings)', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')

    await page.goto('/invoicing', { waitUntil: 'networkidle' })
    if (!page.url().includes('/invoicing')) {
      test.skip(true, 'NUXT_PUBLIC_BILLIT_ENABLED off (redirect)')
    }
    await expect(page.getByTestId('invoicing-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('invoicing-connect-start')).toHaveCount(0)
    if (!INVOICING_UI_ENABLED) {
      await expect(page.getByTestId('invoicing-under-development')).toBeVisible({ timeout: 10000 })
      return
    }
    await expect(page.getByTestId('invoicing-under-development')).toHaveCount(0)
    await expect(page.getByTestId('invoicing-connection-readonly')).toBeVisible({ timeout: 10000 })
  })

  test('B8: secretary — voit onglet partages (read) sans formulaire manage', async ({ page }) => {
    await loginExpectDashboard(page, 'secretary.demo@petsfollow.test')
    await page.goto('/clients', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('clients-page')).toBeVisible({ timeout: 15000 })
    const profile = page.locator('[data-testid^="client-profile-"]').first()
    await expect(profile).toBeVisible({ timeout: 15000 })
    await profile.click()
    await expect(page.getByTestId('section-tab-sharing')).toBeVisible({ timeout: 15000 })
    await page.getByTestId('section-tab-sharing').click()
    await expect(page.getByTestId('client-shares-card')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('client-share-permission')).toHaveCount(0)
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

  test('D2: vet.demo — ACL labels i18n (not raw keys)', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await page.goto('/team', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('vet-team-page')).toBeVisible({ timeout: 15000 })
    const perms = page.getByTestId('team-perms').first()
    await expect(perms).toBeVisible({ timeout: 10000 })
    await expect(perms.getByText(/voir les clients|view clients/i)).toBeVisible({ timeout: 5000 })
    await expect(perms.getByText('clients.read')).toHaveCount(0)
    await expect(perms.getByText('pets.write_clinical')).toHaveCount(0)
  })

  test('E: vet.demo — /commissions allowed for reference', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await page.goto('/commissions', { waitUntil: 'networkidle' })
    await expect(page).toHaveURL(/commissions/)
    await expect(page.getByTestId('vet-commissions-page')).toBeVisible({ timeout: 15000 })
  })

  test('E2: vet.demo — nav commissions + sections', async ({ page }) => {
    await loginExpectDashboard(page, 'vet.demo@petsfollow.test')
    await expect(page.getByTestId('nav-commissions')).toBeVisible({ timeout: 10000 })
    await expect(page.getByText(/^Journée$|^Daily$/i).first()).toBeVisible({ timeout: 10000 })
    await expect(page.getByText(/^Finance$/i).first()).toBeVisible({ timeout: 10000 })
  })
})
