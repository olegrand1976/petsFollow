import type { Page } from '@playwright/test'

export async function login(page: Page, email: string, password: string) {
  await page.goto('/login')
  await page.getByTestId('login-email').fill(email)
  await page.getByTestId('login-password').fill(password)
  await page.getByTestId('login-submit').click()
}

export async function loginAsVet(page: Page, email = 'vet.demo@petsfollow.test', password = 'VetDemo123!') {
  await login(page, email, password)
}

export async function loginAsAdmin(page: Page, email = 'admin.demo@petsfollow.test', password = 'AdminDemo123!') {
  await login(page, email, password)
}
