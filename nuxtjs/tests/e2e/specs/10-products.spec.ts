import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page produits affiche les 2 solutions Pro et les plans clients TTC', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/produits')
  await expect(page.getByTestId('products-page')).toBeVisible()
  await expect(page.getByTestId('products-solution-proComplete')).toBeVisible()
  await expect(page.getByTestId('products-solution-proLight')).toBeVisible()
  // SaaS Pro Complet HT
  await expect(page.getByText(/69/).first()).toBeVisible()
  // Plans clients TTC — FR: "3,50 €" · EN: "€3.50"
  await expect(page.getByText(/€?\s*3[,.]50\s*€?/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*35\b/).first()).toBeVisible()
  await expect(page.getByText(/€?\s*95\b/).first()).toBeVisible()
})
