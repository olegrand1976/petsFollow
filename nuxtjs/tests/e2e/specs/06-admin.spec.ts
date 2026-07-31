import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsCommercial, uniqueE2EEmail, fillField, nativeClick } from '../helpers/auth'

test('admin accède au tableau de bord', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page).toHaveURL(/admin/)
  await expect(page.getByTestId('admin-dashboard-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /admin|dashboard|tableau de bord/i })).toBeVisible()
})

test('admin crée un commercial', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/users?tab=commercial', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('admin-create-commercial')).toBeVisible()
  const email = uniqueE2EEmail('pw-commercial')
  // Remplir le nom en dernier : le fill password peut vider le champ name (autofill navigateur).
  await fillField(page, 'admin-commercial-email', email)
  await fillField(page, 'admin-commercial-password', 'CommercialDemo123!')
  await fillField(page, 'admin-commercial-name', 'E2E Commercial')
  await expect(page.getByTestId('admin-commercial-name')).toHaveValue('E2E Commercial')
  await Promise.all([
    page.waitForResponse(
      (r) => r.request().method() === 'POST' && /\/api\/admin\/commercials\/?$/.test(new URL(r.url()).pathname),
      { timeout: 20000 },
    ),
    nativeClick(page, 'admin-commercial-submit'),
  ])
  await expect(page.getByTestId('admin-commercial-msg')).toBeVisible({ timeout: 15000 })
})

test('admin voit commercials et prospects', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/commercials')
  await expect(page.getByTestId('admin-commercials-page')).toBeVisible()
  await expect(page.getByTestId('admin-assign-vet')).toBeVisible()
  await expect(page.getByTestId('admin-assign-vet-select')).toBeVisible()
  await page.goto('/admin/prospects')
  await expect(page.getByTestId('admin-prospects-page')).toBeVisible()
})

test('admin voit page facturation Billit', async ({ page }) => {
  await loginAsAdmin(page)
  await page.goto('/admin/invoicing', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('admin-invoicing-page')).toBeVisible({ timeout: 15000 })
  // Liste vide OK si aucun cabinet connecté ; erreur seulement si API Billit off.
  const err = page.getByTestId('admin-invoicing-error')
  if (await err.isVisible()) {
    await expect(err).toContainText(/désactiv|disabled|uitgeschakeld|desactiv|välja|disattiv/i)
  } else {
    await expect(page.getByTestId('admin-invoicing-list')).toBeVisible()
    await expect(page.getByTestId('admin-invoicing-export-csv')).toBeVisible()
    await expect(page.getByTestId('admin-invoicing-saas-list')).toBeVisible()
  }
})

test.describe('admin Flux A SaaS I7.9', { tag: ['@p1', '@invoicing'] }, () => {
  test('I7.9 draft puis send Peppol master (mock)', async ({ page }) => {
    await loginAsAdmin(page)
    page.on('dialog', (d) => d.accept())
    await page.goto('/admin/invoicing', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-invoicing-page')).toBeVisible({ timeout: 15000 })

    const err = page.getByTestId('admin-invoicing-error')
    if (await err.isVisible()) {
      test.skip(true, 'Billit API disabled')
    }

    await expect(page.getByTestId('admin-invoicing-saas-list')).toBeVisible()
    const enableBtn = page.locator('[data-testid^="admin-invoicing-saas-enable-"]').first()
    if (await enableBtn.count() > 0) {
      await enableBtn.click()
      await expect(page.getByTestId('admin-invoicing-msg')).toBeVisible({ timeout: 20000 })
    }
    const draftBtn = page.locator('[data-testid^="admin-invoicing-saas-draft-"]').first()
    if (await draftBtn.count() === 0) {
      test.skip(true, 'no SaaS draft button (master off or all delivered)')
    }
    const draftTid = await draftBtn.getAttribute('data-testid')
    const practiceId = draftTid?.replace('admin-invoicing-saas-draft-', '') || ''
    await draftBtn.click()
    await expect(page.getByTestId('admin-invoicing-msg')).toBeVisible({ timeout: 20000 })

    const sendBtn = page.getByTestId(`admin-invoicing-saas-send-${practiceId}`)
    await expect(sendBtn).toBeVisible({ timeout: 10000 })
    await sendBtn.click()
    await expect(page.getByTestId('admin-invoicing-msg')).toContainText(/delivered|envoyé|verzonden|enviad|saadetud|inviat/i, { timeout: 20000 })
  })
})

test.describe('admin filiation', { tag: '@p1' }, () => {
  test('admin ouvre la page filiation avec filtres', async ({ page }) => {
    await loginAsAdmin(page)
    await page.goto('/admin/filiation', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-filiation-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('filiation-table')).toBeVisible()
    await expect(page.getByTestId('filiation-error')).toHaveCount(0)
    await expect(page.getByTestId('filiation-branch')).toBeVisible()
    await expect(page.getByTestId('filiation-commercial')).toBeVisible()
    await expect(page.getByTestId('filiation-export-csv')).toBeVisible()
    await expect(page.getByTestId('filiation-history')).toBeVisible()
  })
})

test('commercial bloqué sur admin', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/admin')
  await expect(page).not.toHaveURL(/\/admin$/)
})
