import { test, expect } from '@playwright/test'
import { loginAsCommercial, uniqueE2EEmail, fillField } from '../helpers/auth'

test('commercial accède au dashboard KPI', async ({ page }) => {
  await loginAsCommercial(page)
  await expect(page).toHaveURL(/commercial/)
  await expect(page.getByTestId('commercial-dashboard-page')).toBeVisible()
})

test('commercial encode un véto', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/vets')
  await expect(page.getByTestId('commercial-vets-page')).toBeVisible()
  await expect(page.getByTestId('commercial-vet-form')).toBeVisible()
  const email = uniqueE2EEmail('pw-vet')
  await fillField(page, 'encode-vet-name', 'Dr E2E')
  await fillField(page, 'encode-vet-practice', 'Cabinet E2E')
  await fillField(page, 'encode-vet-email', email)
  await fillField(page, 'encode-vet-password', 'VetDemo123!')
  await fillField(page, 'encode-vet-city', 'Lyon')
  await page.getByTestId('encode-vet-submit').click()
  await expect(page.getByTestId('encode-vet-name')).toHaveValue('', { timeout: 15000 })
})

test('commercial voit pitch et commissions', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/pitch')
  await expect(page.getByTestId('commercial-pitch-page')).toBeVisible()
  await page.goto('/commercial/commissions')
  await expect(page.getByTestId('commercial-commissions-page')).toBeVisible()
})

test('commercial CRM prospects', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/prospects')
  await expect(page.getByTestId('commercial-prospects-page')).toBeVisible()
  await expect(page.getByTestId('commercial-prospect-form')).toBeVisible()
  await fillField(page, 'prospect-practice', `Prospect E2E ${Date.now()}`)
  await fillField(page, 'prospect-contact', 'Dr Prospect')
  await page.getByTestId('prospect-submit').click()
  await expect(page.getByTestId('prospect-practice')).toHaveValue('', { timeout: 15000 })
})
