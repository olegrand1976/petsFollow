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
