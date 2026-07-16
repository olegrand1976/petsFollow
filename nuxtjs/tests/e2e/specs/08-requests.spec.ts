import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('page demandes affiche invitations et RDV', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/requests')
  await expect(page.getByTestId('requests-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /demandes|requests|aanvragen/i })).toBeVisible()
})

test('nav demandes accessible depuis le dashboard', async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)
  await page.getByTestId('nav-requests').click()
  await expect(page).toHaveURL(/\/requests/)
  await expect(page.getByTestId('requests-page')).toBeVisible()
})

test('confirmer un RDV demandé si présent', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/requests')
  await expect(page.getByTestId('requests-page')).toBeVisible()

  const visitRow = page.locator('[data-testid^="visit-request-"]').first()
  if ((await visitRow.count()) === 0) {
    test.skip(true, 'Aucun RDV requested dans le seed')
    return
  }

  const confirmBtn = visitRow.getByRole('button', { name: /confirmer|confirm|bevestigen/i })
  await confirmBtn.click()
  await expect(visitRow).toHaveCount(0, { timeout: 15000 })
})

test('accepter une invitation de liaison si présente', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/requests')
  await expect(page.getByTestId('requests-page')).toBeVisible()

  const linkRow = page.locator('[data-testid^="link-request-"]').first()
  if ((await linkRow.count()) === 0) {
    test.skip(true, 'Aucune invitation pending dans le seed')
    return
  }

  const acceptBtn = linkRow.getByRole('button', { name: /accepter|accept|aanvaarden/i })
  await acceptBtn.click()
  await expect(linkRow).toHaveCount(0, { timeout: 15000 })
})
