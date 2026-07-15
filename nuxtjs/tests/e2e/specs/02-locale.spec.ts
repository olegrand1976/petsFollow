import { test, expect } from '@playwright/test'

test('changement de langue dans paramètres', async ({ page }) => {
  await page.goto('/login')
  const email = page.getByTestId('login-email')
  const password = page.getByTestId('login-password')
  if (await email.inputValue() === '') {
    await email.fill('vet.demo@petsfollow.test')
    await password.fill('VetDemo123!')
  }
  await page.getByTestId('login-submit').click()
  await expect(page).toHaveURL(/dashboard/)

  await page.goto('/settings')
  await page.getByTestId('settings-locale-en').click()
  await page.getByRole('button', { name: /save|enregistrer/i }).click()
  await expect(page.getByTestId('settings-locale-en')).toHaveClass(/pro-toggle-btn--active/)
})
