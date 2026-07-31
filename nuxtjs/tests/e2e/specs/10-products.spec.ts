import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page produits affiche Pro Web, Pro Light et les plans clients TTC', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/produits')
  await expect(page.getByTestId('products-page')).toBeVisible()
  const pro = page.getByTestId('products-solution-proComplete')
  await expect(pro).toBeVisible()
  // Offre Pro = Plateforme Web (cabinet) — assertion FR locale
  await expect(pro.getByRole('heading', { level: 2 })).toContainText(/Web/i)
  await expect(page.getByTestId('products-solution-proLight')).toBeVisible()
  // SaaS Pro HT
  await expect(page.getByText(/69/).first()).toBeVisible()
  // Plans clients TTC — FR: "3,50 €" · EN: "€3.50"
  await expect(page.getByText(/€?\s*3[,.]50\s*€?/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*35\b/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*95\b/).first()).toBeVisible()
})
