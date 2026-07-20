import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page calendrier accessible', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/calendar')
  await expect(page.getByTestId('calendar-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /calendrier|calendar|agenda/i })).toBeVisible()
})

test('nav calendrier accessible depuis le dashboard', async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)
  await page.getByTestId('nav-calendar').click()
  await expect(page).toHaveURL(/\/calendar/)
  await expect(page.getByTestId('calendar-page')).toBeVisible()
})

test('redirect /requests vers calendrier', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/requests')
  await expect(page).toHaveURL(/\/calendar/)
})

test('confirmer un RDV demandé si présent', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/calendar')
  await expect(page.getByTestId('calendar-page')).toBeVisible()

  const visitRow = page.locator('[data-testid^="visit-request-"]').first()
  if ((await visitRow.count()) === 0) {
    test.skip(true, 'Aucun RDV pending dans le seed')
    return
  }

  const confirmBtn = visitRow.getByRole('button', { name: /confirmer|confirm|bevestigen/i })
  await confirmBtn.click()
  await expect(visitRow).toHaveCount(0, { timeout: 15000 })
})

test('invitations clients dans l’en-tête Clients', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await page.getByTestId('clients-invitations-open').click()
  await expect(page.getByRole('heading', { name: /invitation/i })).toBeVisible()

  const linkRow = page.locator('[data-testid^="link-request-"]').first()
  if ((await linkRow.count()) === 0) {
    test.skip(true, 'Aucune invitation pending dans le seed')
    return
  }

  const acceptBtn = linkRow.getByRole('button', { name: /accepter|accept|aanvaarden/i })
  await acceptBtn.click()
  await expect(linkRow).toHaveCount(0, { timeout: 15000 })
})
