import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsVet } from '../helpers/auth'

test('admin voit le catalogue Stripe', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/stripe-catalog')
  await expect(page.getByTestId('admin-stripe-catalog-page')).toBeVisible()
  await expect(page.getByTestId('admin-stripe-import-products')).toBeVisible()
  await expect(page.getByTestId('admin-stripe-import-prices')).toBeVisible()
  await expect(page.getByTestId('admin-stripe-products-file')).toBeVisible()
  await expect(page.getByTestId('admin-stripe-prices-file')).toBeVisible()
})

test('véto bloqué sur catalogue Stripe admin', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/admin/stripe-catalog')
  await expect(page).not.toHaveURL(/\/admin\/stripe-catalog/)
})
