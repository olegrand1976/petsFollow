import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsCommercial, uniqueE2EEmail, fillField, nativeClick } from '../helpers/auth'

test('admin accède au tableau de bord', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page).toHaveURL(/admin/)
  await expect(page.getByTestId('admin-dashboard-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /admin|dashboard|tableau de bord/i })).toBeVisible()
})

test('admin crée un commercial', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/users', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('admin-create-commercial')).toBeVisible()
  const email = uniqueE2EEmail('pw-commercial')
  // Remplir le nom en dernier : le fill password peut vider le champ name (autofill navigateur).
  await fillField(page, 'admin-commercial-email', email)
  await fillField(page, 'admin-commercial-password', 'CommercialDemo123!')
  await fillField(page, 'admin-commercial-name', 'E2E Commercial')
  await expect(page.getByTestId('admin-commercial-name')).toHaveValue('E2E Commercial')
  await Promise.all([
    page.waitForResponse(
      (r) => r.request().method() === 'POST' && /\/api\/admin\/commercials\/?$/.test(new URL(r.url()).pathname),
      { timeout: 20000 },
    ),
    nativeClick(page, 'admin-commercial-submit'),
  ])
  await expect(page.getByTestId('admin-commercial-msg')).toBeVisible({ timeout: 15000 })
})

test('admin voit commercials et prospects', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/commercials')
  await expect(page.getByTestId('admin-commercials-page')).toBeVisible()
  await expect(page.getByTestId('admin-assign-vet')).toBeVisible()
  await expect(page.getByTestId('admin-assign-vet-select')).toBeVisible()
  await page.goto('/admin/prospects')
  await expect(page.getByTestId('admin-prospects-page')).toBeVisible()
})

test('admin voit page bonus SPIFF', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/commercial-bonuses')
  await expect(page.getByTestId('admin-commercial-bonuses-page')).toBeVisible()
  await expect(page.getByTestId('bonus-filter-status')).toBeVisible()
})

test('commercial bloqué sur admin', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/admin')
  await expect(page).not.toHaveURL(/\/admin$/)
})
