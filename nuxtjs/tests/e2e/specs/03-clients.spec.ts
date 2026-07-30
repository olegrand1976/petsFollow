import { test, expect, type Page } from '@playwright/test'
import { fillField, loginAsVet } from '../helpers/auth'

/** /clients peut ouvrir un ProModal (invitations) qui bloque les clics. */
async function dismissProModals(page: Page) {
  for (let i = 0; i < 3; i++) {
    const close = page.getByTestId('pro-modal-close')
    if (await close.isVisible().catch(() => false)) {
      await close.click({ force: true }).catch(() => {})
      await page.waitForTimeout(200)
    } else {
      break
    }
  }
}

test('liste clients avec recherche', { tag: '@p0' }, async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)

  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await dismissProModals(page)

  const search = page.locator('#client-search')
  await search.fill('Sophie')
  // Timeout large : premier chargement de la page en dev (compilation Vite) sous suite complète.
  // Ne pas assert le téléphone seed : staging peut ne pas avoir re-seed les contact_phone.
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })
})

test('create + edit client contactPhone', { tag: '@p1' }, async ({ page }) => {
  await loginAsVet(page)
  await page.addInitScript(() => {
    try {
      localStorage.setItem('pf-clients-view', 'table')
    } catch {
      /* ignore */
    }
  })
  await page.goto('/clients', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await dismissProModals(page)

  const email = `e2e.phone.${Date.now()}@petsfollow.test`
  const phone = '0470 55 66 77'
  const updated = '0470 55 66 88'

  await page.getByTestId('create-client-open').click()
  await expect(page.getByTestId('vet-create-client-form')).toBeVisible({ timeout: 10000 })
  await fillField(page, 'create-client-name', 'E2E Phone')
  await fillField(page, 'create-client-email', email)
  await fillField(page, 'create-client-phone', phone)
  await fillField(page, 'create-client-password', 'TempPass12!')
  await page.getByTestId('create-client-submit').click()
  await expect(page.getByTestId('create-client-msg')).toBeVisible({ timeout: 20000 })

  // Ferme la modale create pour retrouver la liste.
  const modalClose = page.getByTestId('pro-modal-close')
  if (await modalClose.isVisible().catch(() => false)) {
    await modalClose.click()
  }
  await dismissProModals(page)

  await page.locator('#client-search').fill(email)
  await expect(page.getByText(phone).first()).toBeVisible({ timeout: 15000 })

  const profileLink = page.locator('[data-testid^="client-profile-"]').first()
  await expect(profileLink).toBeVisible({ timeout: 10000 })
  await profileLink.click()

  const phoneInput = page.getByTestId('client-phone-input')
  await phoneInput.scrollIntoViewIfNeeded()
  // Staging CI a parfois reporté l'input "hidden" malgré le DOM — force + valeur.
  await expect(phoneInput).toBeAttached({ timeout: 15000 })
  await expect(phoneInput).toHaveValue(phone, { timeout: 10000 })

  await fillField(page, 'client-phone-input', updated)
  const patch = page.waitForResponse(
    (r) => /\/api\/clients\/[^/]+$/.test(r.url()) && r.request().method() === 'PATCH' && r.ok(),
    { timeout: 15000 },
  )
  await page.getByTestId('client-phone-save').click({ force: true })
  await patch
  await expect(phoneInput).toHaveValue(updated)
  await expect(page.getByTestId('client-phone-msg')).toBeVisible({ timeout: 5000 })
})
