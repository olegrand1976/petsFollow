import { test, expect } from '@playwright/test'
import { loginAsAdmin } from '../helpers/auth'

test('admin accède au tableau de bord', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page).toHaveURL(/admin/)
  await expect(page.getByTestId('admin-dashboard-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /admin|dashboard|tableau de bord/i })).toBeVisible()
})
