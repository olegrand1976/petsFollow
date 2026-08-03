import { test, expect } from '@playwright/test'
import { loginAsVet, nativeClick } from '../helpers/auth'

test('page calendrier accessible', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/calendar')
  await expect(page.getByTestId('calendar-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /calendrier|calendar|agenda/i })).toBeVisible()
  await expect(page.getByTestId('calendar-grid')).toBeVisible()
})

test('bascule semaine / mois et aujourd’hui', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/calendar')
  await page.evaluate(() => localStorage.removeItem('pf-calendar-view'))
  await page.reload({ waitUntil: 'networkidle' })
  await expect(page.getByTestId('calendar-grid')).toBeVisible()

  await page.getByTestId('calendar-view-week').dispatchEvent('click')
  await expect(page.getByTestId('calendar-view-week')).toHaveAttribute('aria-pressed', 'true')

  await page.getByTestId('calendar-view-month').dispatchEvent('click')
  await expect(page.getByTestId('calendar-view-month')).toHaveAttribute('aria-pressed', 'true')
  await expect(page.getByTestId('calendar-grid')).toBeVisible()

  await page.getByTestId('calendar-view-week').dispatchEvent('click')
  await expect(page.getByTestId('calendar-view-week')).toHaveAttribute('aria-pressed', 'true')

  await page.getByTestId('calendar-today').click({ force: true })
  await expect(page.getByTestId('calendar-grid')).toBeVisible()
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
  await page.goto('/clients', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await nativeClick(page, 'clients-invitations-open')
  const dialog = page.getByTestId('pro-modal')
  await expect(dialog).toBeVisible({ timeout: 10000 })
  await expect(dialog.getByRole('heading')).toContainText(/invitation/i)

  const rows = page.locator('[data-testid^="link-request-"]')
  if ((await rows.count()) === 0) {
    test.skip(true, 'Aucune invitation pending dans le seed')
    return
  }

  // On accepte toutes les invitations : chaque ligne traitée disparaît…
  for (let i = 0; i < 20 && (await rows.count()) > 0; i++) {
    const row = rows.first()
    const rowId = await row.getAttribute('data-testid')
    await row.getByRole('button', { name: /accepter|accept|aanvaarden/i }).click()
    await expect(page.locator(`[data-testid="${rowId}"]`)).toHaveCount(0, { timeout: 15000 })
  }

  // …et quand il n'en reste plus aucune, la modale se ferme d'elle-même.
  await expect(dialog).toBeHidden({ timeout: 10000 })
})
