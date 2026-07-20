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
  await expect(page.getByTestId('commercial-vets-cards')).toBeVisible()
  await expect(page.getByTestId('commercial-card-vet')).toBeVisible()
  await expect(page.getByTestId('commercial-card-client')).toBeVisible()
  await expect(page.getByTestId('commercial-card-client-standalone')).toBeVisible()
  await page.getByTestId('commercial-card-vet').click()
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

test('commercial ouvre les formulaires client lié et sans liaison', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/vets')
  await page.getByTestId('commercial-card-client-standalone').click()
  await expect(page.getByTestId('commercial-client-form')).toBeVisible()
  await expect(page.getByTestId('create-client-vet')).toHaveCount(0)
  await page.getByTestId('commercial-back-cards').click()
  await page.getByTestId('commercial-card-client').click()
  await expect(page.getByTestId('commercial-client-form')).toBeVisible()
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
  await expect(page.getByTestId('commercial-bonus-cards')).toBeVisible()
  await expect(page.getByTestId('bonus-card-commercial_ramp')).toBeVisible()
  await expect(page.getByTestId('bonus-card-commercial_mix')).toBeVisible()
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
