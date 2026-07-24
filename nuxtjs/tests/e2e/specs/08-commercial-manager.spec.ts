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
