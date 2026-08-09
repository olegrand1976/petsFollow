import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

/** /clients peut ouvrir un ProModal (invitations) qui bloque les clics. */
async function dismissProModals(page: Page) {
  for (let i = 0; i < 3; i++) {
    const close = page.getByTestId('pro-modal-close')
    if (await close.isVisible().catch(() => false)) {
      await close.first().click({ force: true }).catch(() => {})
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
  // Téléphone seed / edit fiche : couverture Go `TestClientContactPhone*` (staging sans re-seed fiable).
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })
})

// Modale seule, aucune création : la bascule Particulier doit retirer les champs
// fiscaux, sinon un n° de TVA saisi ici disparaîtrait sans le dire à la facturation.
test('création client : le type particulier masque TVA et n° d\'entreprise', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  // La modale « invitations en attente » s'ouvre d'elle-même dès que le seed en
  // contient, après le chargement : attendre puis fermer, sinon elle avale le clic.
  await page.waitForLoadState('networkidle')
  await dismissProModals(page)

  await page.getByTestId('create-client-open').click()
  await expect(page.getByTestId('create-client-billing-section')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('create-client-billing-vat')).toBeVisible()

  await page.getByTestId('create-client-billing-customer-kind').selectOption('individual')
  await expect(page.getByTestId('create-client-billing-vat')).toHaveCount(0)
  await expect(page.getByTestId('create-client-billing-company')).toHaveCount(0)

  await page.getByTestId('create-client-billing-customer-kind').selectOption('business')
  await expect(page.getByTestId('create-client-billing-vat')).toBeVisible()
})
