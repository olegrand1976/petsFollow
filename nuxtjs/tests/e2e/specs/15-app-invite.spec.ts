import { test, expect } from '@playwright/test'
import { loginAsCommercial, nativeClick } from '../helpers/auth'

test('landing invite role=client n’affiche pas le CTA cabinet', async ({ page }) => {
  await page.route('**/api/public/app-invite/**', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          code: 'CLIENT01',
          role: 'client',
          displayName: 'Marie Demo',
          practiceName: '',
          downloadUrl: 'https://example.com/app',
          deepLink: 'petsfollow://invite?code=CLIENT01',
          inviteUrl: 'http://localhost:3002/invite/CLIENT01',
        },
      }),
    })
  })

  // download=1 empêche l’auto-open deep link mobile.
  await page.goto('/invite/CLIENT01?download=1', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('app-invite-landing')).toBeVisible()
  await expect(page.getByTestId('app-invite-ok')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('app-invite-code')).toBeVisible()
  await expect(page.getByTestId('app-invite-cabinet-cta')).toHaveCount(0)
  await expect(page.getByTestId('app-invite-open-app')).toBeVisible()
})

test('commercial modal invite expose le lien cabinet (vetRegisterUrl)', async ({ page }) => {
  await loginAsCommercial(page)

  await page.route('**/api/me/app-invite', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: {
          code: 'ABCD2345',
          role: 'commercial',
          inviteUrl: 'http://localhost:3002/invite/ABCD2345',
          vetRegisterUrl: 'http://localhost:3002/register?invite=ABCD2345',
          displayName: 'Camille Demo',
          practiceName: '',
          qrCodeDataUrl: 'data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAADUlEQVR42mP8z8BQDwAEhQGAhKmMIQAAAABJRU5ErkJggg==',
        },
      }),
    })
  })

  await page.goto('/commercial', { waitUntil: 'networkidle' })
  await nativeClick(page, 'commercial-app-invite-open')
  await expect(page.getByTestId('app-invite-qr')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('app-invite-url')).toBeVisible()
  await expect(page.getByTestId('app-invite-vet-register-url')).toBeVisible({ timeout: 10000 })
  await expect(page.getByTestId('app-invite-copy-vet-register')).toBeVisible()
  await expect(page.getByTestId('app-invite-vet-register-url')).toHaveValue(/\/register\?invite=ABCD2345/)
})
