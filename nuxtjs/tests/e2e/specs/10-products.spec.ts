import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page produits affiche les plans TTC', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/produits')
  await expect(page.getByTestId('products-page')).toBeVisible()
  // FR: "3,50 €" · EN: "€3.50" (plusieurs occurrences page)
  await expect(page.getByText(/€?\s*3[,.]50\s*€?/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*35\b/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*95\b/).first()).toBeVisible()
})
