import { test, expect } from '@playwright/test'

test('login véto et liste clients', async ({ page }) => {
  await page.goto('/login')
  await expect(page.getByTestId('login-form')).toBeVisible()

  const email = page.getByTestId('login-email')
  const password = page.getByTestId('login-password')

  if (await email.inputValue() === '') {
    await email.fill('vet.demo@petsfollow.test')
    await password.fill('VetDemo123!')
  }

  await page.getByTestId('login-submit').click()
  await expect(page).toHaveURL(/dashboard/)
  await expect(page.getByTestId('pro-topbar')).toBeVisible()
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
})
