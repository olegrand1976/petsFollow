import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page produits affiche Pro Web, Pro Light et les plans clients TTC', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/produits')
  const products = page.getByTestId('products-page')
  await expect(products).toBeVisible()
  const pro = page.getByTestId('products-solution-proComplete')
  await expect(pro).toBeVisible()
  // Offre Pro = Plateforme Web (cabinet) — assertion FR locale
  await expect(pro.getByRole('heading', { level: 2 })).toContainText(/Web/i)
  await expect(page.getByTestId('products-solution-proLight')).toBeVisible()
  // Scope all price asserts to products-page (avoid hidden site <option> timestamps).
  // SaaS Pro HTVA — 834,71 € / an (legacy /69/ matched site options).
  await expect(products.getByText(/834[,.]71/)).toBeVisible()
  // Plans clients TTC — FR: "3,50 €" · "35 €" · "95 €"
  await expect(products.getByText(/€?\s*3[,.]50\s*€?/)).toBeVisible()
  await expect(products.getByText(/35\s*€|€\s*35\b/)).toBeVisible()
  await expect(products.getByText(/95\s*€|€\s*95\b/)).toBeVisible()
})
