import { test, expect } from '@playwright/test'

test('page publique /dossier affiche meta + CTA register', async ({ page }) => {
  await page.route('**/api/public/pet-dossier/**', async (route) => {
    if (route.request().url().includes('/download')) {
      await route.fulfill({
        status: 200,
        contentType: 'application/zip',
        body: Buffer.from('PK\x03\x04mock'),
      })
      return
    }
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          petName: 'Bella',
          expiresAt: new Date(Date.now() + 3600_000).toISOString(),
          commercialName: 'Camille Demo',
          commercialPhone: '0470 12 34 56',
          registerUrl: 'http://localhost:3002/register?invite=ABCD2345',
          siteUrl: 'http://localhost:3002',
          productsUrl: 'http://localhost:3002/produits',
          downloadUrl: '/api/v1/public/pet-dossier/tok-demo/download',
        },
      }),
    })
  })

  await page.goto('/dossier/tok-demo', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('dossier-public-page')).toBeVisible()
  await expect(page.getByTestId('dossier-title')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('dossier-download')).toBeVisible()
  await expect(page.getByTestId('dossier-commercial-phone')).toContainText('0470')
  await expect(page.getByTestId('dossier-register-cta').first()).toBeVisible()
})

test('page publique /dossier expirée montre CTA register', async ({ page }) => {
  await page.route('**/api/public/pet-dossier/**', async (route) => {
    await route.fulfill({
      status: 410,
      contentType: 'application/json',
      body: JSON.stringify({
        error: { code: 'gone', message: 'dossier_expired' },
      }),
    })
  })

  await page.goto('/dossier/tok-expired', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('dossier-expired-title')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('dossier-register-cta')).toBeVisible()
})
