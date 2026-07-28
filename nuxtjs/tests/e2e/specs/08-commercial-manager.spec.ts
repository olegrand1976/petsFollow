import { test, expect } from '@playwright/test'
import { loginAsCommercialManager } from '../helpers/auth'

test('responsable commercial accède au dashboard équipe', async ({ page }) => {
  await loginAsCommercialManager(page)
  await expect(page).toHaveURL(/commercial-manager/)
  await expect(page.getByTestId('manager-dashboard-page')).toBeVisible()
})

test('responsable commercial voit suivi et prospects équipe', async ({ page }) => {
  await loginAsCommercialManager(page)
  await page.goto('/commercial-manager/suivi')
  await expect(page.getByTestId('manager-followups-page')).toBeVisible()
  await page.goto('/commercial-manager/prospects')
  await expect(page.getByTestId('manager-prospects-page')).toBeVisible()
})

test('responsable commercial ouvre le mémo ASV', async ({ page }) => {
  await loginAsCommercialManager(page)
  await expect(page.getByTestId('nav-commercial-asv-memo')).toBeVisible()
  await page.goto('/commercial/asv-memo', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('commercial-asv-memo')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('asv-memo-print')).toBeVisible()
  await expect(page.getByTestId('nav-commercial-settings')).toBeVisible()
})

test.describe('manager filiation', { tag: '@p1' }, () => {
  test('responsable ouvre la page filiation', async ({ page }) => {
    await loginAsCommercialManager(page)
    await page.goto('/commercial-manager/filiation', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('manager-filiation-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('filiation-table')).toBeVisible()
    await expect(page.getByTestId('filiation-error')).toHaveCount(0)
    await expect(page.getByTestId('filiation-export-csv')).toBeVisible()
    await expect(page.getByTestId('filiation-history')).toBeVisible()
  })
})
