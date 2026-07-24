import { test, expect } from '@playwright/test'
import { loginAsCommercial, uniqueE2EEmail, fillField, nativeClick } from '../helpers/auth'

test('commercial accède au dashboard KPI', async ({ page }) => {
  await loginAsCommercial(page)
  await expect(page).toHaveURL(/commercial/)
  await expect(page.getByTestId('commercial-dashboard-page')).toBeVisible()
})

test('commercial encode un véto', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/vets', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-vets-cards')).toBeVisible()
  await nativeClick(page, 'commercial-card-vet')
  await expect(page.getByTestId('commercial-vet-form')).toBeVisible({ timeout: 10000 })
  const email = uniqueE2EEmail('pw-vet')
  await fillField(page, 'encode-vet-email', email)
  await fillField(page, 'encode-vet-password', 'VetDemo123!')
  await fillField(page, 'encode-vet-name', 'Dr E2E')
  await fillField(page, 'encode-vet-practice', 'Cabinet E2E')
  await fillField(page, 'encode-vet-city', 'Lyon')
  await nativeClick(page, 'encode-vet-submit')
  await expect(page.getByTestId('encode-vet-name')).toHaveValue('', { timeout: 15000 })
})

test('commercial ouvre les formulaires client lié et sans liaison', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/vets', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-vets-cards')).toBeVisible()
  await nativeClick(page, 'commercial-card-client-standalone')
  await expect(page.getByTestId('commercial-client-form')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('create-client-vet')).toHaveCount(0)
  await nativeClick(page, 'commercial-back-cards')
  await expect(page.getByTestId('commercial-vets-cards')).toBeVisible()
  await nativeClick(page, 'commercial-card-client')
  await expect(page.getByTestId('commercial-client-form')).toBeVisible({ timeout: 10000 })
  const noVets = page.getByTestId('create-client-no-vets')
  const vetSelect = page.getByTestId('create-client-vet')
  await expect(noVets.or(vetSelect)).toBeVisible()
})

test('commercial voit pitch et commissions', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/pitch')
  await expect(page.getByTestId('commercial-pitch-page')).toBeVisible()
  await page.goto('/commercial/commissions')
  await expect(page.getByTestId('commercial-commissions-page')).toBeVisible()
  await page.getByTestId('commission-details').locator('summary').click()
  await expect(page.getByTestId('commercial-bonus-cards')).toBeVisible()
  await expect(page.getByTestId('bonus-card-commercial_ramp')).toBeVisible()
  await expect(page.getByTestId('bonus-card-commercial_mix')).toBeVisible()
})

test('commercial CRM prospects', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/prospects', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-prospects-page')).toBeVisible()
  await expect(page.getByTestId('commercial-prospect-form')).toBeVisible()
  await nativeClick(page, 'prospect-create-toggle')
  await expect(page.getByTestId('prospect-practice')).toBeVisible({ timeout: 10000 })
  const practice = `Prospect E2E ${Date.now()}`
  await fillField(page, 'prospect-practice', practice)
  await fillField(page, 'prospect-contact', 'Dr Prospect')
  await Promise.all([
    page.waitForResponse(
      (r) => r.request().method() === 'POST' && r.url().includes('/commercial/prospects'),
      { timeout: 20000 },
    ),
    nativeClick(page, 'prospect-submit'),
  ])
  // Succès → formulaire refermé (showCreate = false)
  await expect(page.getByTestId('prospect-practice')).toHaveCount(0, { timeout: 15000 })
  await expect(page.getByTestId('prospect-source-filter')).toHaveValue('commercial')
  await expect(page.getByText(practice)).toBeVisible({ timeout: 15000 })
})
