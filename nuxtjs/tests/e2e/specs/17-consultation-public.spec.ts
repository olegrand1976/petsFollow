import { test, expect } from '@playwright/test'

test.describe('consultation public', { tag: '@p1' }, () => {
  test('page publique /consultation affiche meta + CTA register', async ({ page }) => {
    await page.route('**/api/public/consultation/**', async (route) => {
      if (route.request().url().includes('/download')) {
        await route.fulfill({
          status: 200,
          contentType: 'application/pdf',
          body: Buffer.from('%PDF-1.4 mock'),
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
            downloadUrl: '/api/v1/public/consultation/tok-demo/download',
          },
        }),
      })
    })

    await page.goto('/consultation/tok-demo', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('consultation-public-page')).toBeVisible()
    await expect(page.getByTestId('consultation-title')).toBeVisible({ timeout: 10000 })
    const download = page.getByTestId('consultation-download')
    await expect(download).toBeVisible()
    await expect(download).toHaveAttribute('href', /\/api\/public\/consultation\/tok-demo\/download/)
    await expect(download).toHaveAttribute('target', '_blank')
    await expect(page.getByTestId('consultation-commercial-phone')).toContainText('0470')
    await expect(page.getByTestId('consultation-register-cta').first()).toBeVisible()
  })

  test('page publique /consultation expirée montre CTA register', async ({ page }) => {
    await page.route('**/api/public/consultation/**', async (route) => {
      await route.fulfill({
        status: 410,
        contentType: 'application/json',
        body: JSON.stringify({
          error: { code: 'gone', message: 'consultation_expired' },
        }),
      })
    })

    await page.goto('/consultation/tok-expired', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('consultation-expired-title')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('consultation-register-cta')).toBeVisible()
  })
})
