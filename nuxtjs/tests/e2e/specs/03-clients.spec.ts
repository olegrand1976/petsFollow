import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('liste clients avec recherche', async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)

  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()

  const search = page.getByPlaceholder(/nom ou email|name or email|naam of e-mail/i)
  await search.fill('Sophie')
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible()
})
