import { test, expect } from '@playwright/test'
import { fillField, loginAsVet } from '../helpers/auth'

const SEED_PHONE = '0470 00 00 01'

test('liste clients avec recherche', { tag: '@p0' }, async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)

  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()

  const search = page.locator('#client-search')
  await search.fill('Sophie')
  // Timeout large : premier chargement de la page en dev (compilation Vite) sous suite complète.
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })
  // Seed contactPhone Sophie (C2.2 / COMPTES) — lecture seule, pas de mutation @p0.
  await expect(page.getByText(SEED_PHONE).first()).toBeVisible({ timeout: 10000 })
})

test('fiche client : edit contactPhone + restore seed', { tag: '@p1' }, async ({ page }) => {
  await loginAsVet(page)
  // Force table pour `client-profile-*` (kanban n’a pas le testid).
  await page.addInitScript(() => {
    try {
      localStorage.setItem('pf-clients-view', 'table')
    } catch {
      /* ignore */
    }
  })
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()

  const search = page.locator('#client-search')
  await search.fill('Sophie')
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })

  const profileLink = page.locator('[data-testid^="client-profile-"]').first()
  await expect(profileLink).toBeVisible({ timeout: 10000 })
  await profileLink.click()

  const phoneInput = page.getByTestId('client-phone-input')
  await expect(phoneInput).toBeVisible({ timeout: 15000 })

  const restorePhone = async () => {
    const current = await phoneInput.inputValue()
    if (current === SEED_PHONE) return
    await fillField(page, 'client-phone-input', SEED_PHONE)
    const patch = page.waitForResponse(
      (r) => /\/api\/clients\/[^/]+$/.test(r.url()) && r.request().method() === 'PATCH' && r.ok(),
      { timeout: 15000 },
    )
    await page.getByTestId('client-phone-save').click()
    await patch
    await expect(phoneInput).toHaveValue(SEED_PHONE)
  }

  try {
    await fillField(page, 'client-phone-input', '0470 00 00 99')
    const patch = page.waitForResponse(
      (r) => /\/api\/clients\/[^/]+$/.test(r.url()) && r.request().method() === 'PATCH' && r.ok(),
      { timeout: 15000 },
    )
    await page.getByTestId('client-phone-save').click()
    await patch
    await expect(phoneInput).toHaveValue('0470 00 00 99')
    await expect(page.getByTestId('client-phone-msg')).toBeVisible({ timeout: 5000 })
  } finally {
    await restorePhone()
  }
})
