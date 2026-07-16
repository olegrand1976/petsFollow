import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsCommercial, uniqueE2EEmail, fillField } from '../helpers/auth'

test('admin accède au tableau de bord', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page).toHaveURL(/admin/)
  await expect(page.getByTestId('admin-dashboard-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /admin|dashboard|tableau de bord/i })).toBeVisible()
})

test('admin crée un commercial', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/users')
  await expect(page.getByTestId('admin-create-commercial')).toBeVisible()
  const email = uniqueE2EEmail('pw-commercial')
  await fillField(page, 'admin-commercial-name', 'E2E Commercial')
  await fillField(page, 'admin-commercial-email', email)
  await fillField(page, 'admin-commercial-password', 'CommercialDemo123!')
  await page.getByTestId('admin-commercial-submit').click()
  await expect(page.getByTestId('admin-commercial-msg')).toBeVisible({ timeout: 15000 })
})

test('admin voit commercials et prospects', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/commercials')
  await expect(page.getByTestId('admin-commercials-page')).toBeVisible()
  await page.goto('/admin/prospects')
  await expect(page.getByTestId('admin-prospects-page')).toBeVisible()
})

test('commercial bloqué sur admin', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/admin')
  await expect(page).not.toHaveURL(/\/admin$/)
})
