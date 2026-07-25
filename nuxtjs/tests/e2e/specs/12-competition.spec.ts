import { test, expect } from '@playwright/test'
import { loginAsCommercial } from '../helpers/auth'

test.describe('12-competition', () => {
  test('commercial voit la page concurrence et le tab BE', async ({ page }) => {
    await loginAsCommercial(page)
    await page.goto('/commercial/competition')
    await expect(page.getByTestId('commercial-competition-page')).toBeVisible()
    await expect(page.getByTestId('competition-synthesis')).toBeVisible()
    await page.getByTestId('competition-tab-be').click()
    await expect(page.getByTestId('competition-synthesis')).toBeVisible()
  })
})
