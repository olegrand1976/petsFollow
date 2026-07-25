import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('véto profil incomplet redirigé vers onboarding', async ({ page }) => {
  await loginAsVet(page, 'vet.onboarding@petsfollow.test')
  await expect(page).toHaveURL(/onboarding/)
  await expect(page.getByTestId('onboarding-page')).toBeVisible()
})
