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
  test.setTimeout(60000)
  await loginAsCommercial(page)
  await page.goto('/commercial/pitch')
  await expect(page.getByTestId('commercial-pitch-page')).toBeVisible()
  await expect(page.getByTestId('pitch-open-deck')).toBeVisible()
  await expect(page.getByTestId('pitch-open-brochure')).toBeVisible()
  await expect(page.getByTestId('pitch-open-asv-memo')).toBeVisible()
  await page.goto('/commercial/brochure', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('commercial-brochure')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('brochure-print')).toBeVisible()
  await expect(page.getByTestId('nav-commercial-brochure')).toBeVisible()
  await page.goto('/commercial/asv-memo', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('commercial-asv-memo')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('asv-memo-print')).toBeVisible()
  await expect(page.getByTestId('nav-commercial-asv-memo')).toBeVisible()
  await page.goto('/commercial/pitch-deck', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/\/presentation\?step=welcome/, { timeout: 20000 })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 20000 })
  // Client-side step nav after pitch-deck redirect is flaky on Cloud Run; deep-link instead (25 covers TOC).
  await page.goto('/presentation?step=pain', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/[?&]step=pain(?:&|$)/, { timeout: 15000 })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 10000 })
  await page.goto('/commercial/commissions', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-commissions-page')).toBeVisible()
  // Bonus cards are always visible (no details disclosure on this page).
  await expect(page.getByTestId('commercial-bonus-cards')).toBeVisible({ timeout: 15000 })
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

test('commercial mail templates et envoi prospect', async ({ page }) => {
  test.setTimeout(90000)
  await loginAsCommercial(page)
  await page.goto('/commercial/email-templates', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-email-templates-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('email-tpl-row-intro_after_request')).toBeVisible({ timeout: 15000 })
  await nativeClick(page, 'email-tpl-edit-intro_after_request')
  await expect(page.getByTestId('email-tpl-editor')).toBeVisible()
  await expect(page.getByTestId('email-tpl-subject')).not.toHaveValue('')

  await page.goto('/commercial/prospects', { waitUntil: 'networkidle' })
  await nativeClick(page, 'prospect-create-toggle')
  const practice = `Prospect Mail ${Date.now()}`
  await fillField(page, 'prospect-practice', practice)
  await fillField(page, 'prospect-contact', 'Dr Mail')
  await fillField(page, 'prospect-email', `mail.${Date.now()}@petsfollow.test`)
  await Promise.all([
    page.waitForResponse(
      (r) => r.request().method() === 'POST' && r.url().includes('/commercial/prospects'),
      { timeout: 20000 },
    ),
    nativeClick(page, 'prospect-submit'),
  ])
  await expect(page.getByText(practice)).toBeVisible({ timeout: 15000 })
  const row = page.locator('tr').filter({ hasText: practice }).first()
  await row.locator('[data-testid^="prospect-email-"]').click()
  await expect(page.getByTestId('prospect-mail-modal')).toBeVisible({ timeout: 10000 })
  await page.getByTestId('prospect-mail-template').selectOption({ index: 1 })
  await Promise.all([
    page.waitForResponse(
      (r) => r.request().method() === 'POST' && r.url().includes('/emails'),
      { timeout: 20000 },
    ),
    nativeClick(page, 'prospect-mail-send'),
  ])
  await expect(page.getByTestId('prospect-mail-history')).toBeVisible({ timeout: 15000 })
  await page.goto('/commercial/emails', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('commercial-emails-page')).toBeVisible()
  await expect(page.locator('[data-testid^="email-send-row-"]').first()).toBeVisible({ timeout: 15000 })
})

test.describe('commercial filiation', { tag: '@p1' }, () => {
  test('commercial ouvre la page filiation', async ({ page }) => {
    await loginAsCommercial(page)
    await page.goto('/commercial/filiation', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('commercial-filiation-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('filiation-table')).toBeVisible()
    await expect(page.getByTestId('filiation-error')).toHaveCount(0)
    await expect(page.getByTestId('filiation-export-csv')).toBeVisible()
    await expect(page.getByTestId('filiation-history')).toBeVisible()
  })
})
