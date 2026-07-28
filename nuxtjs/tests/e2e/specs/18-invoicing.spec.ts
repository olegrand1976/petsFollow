import { test, expect } from '@playwright/test'
import { login } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

const STAFF_PASSWORD = 'VetDemo123!'

/**
 * UI facturation gelée via `INVOICING_UI_ENABLED` (`nuxtjs/utils/invoicing-ui.ts`).
 * Si le flag est `true` : restaurer le scénario connect → create → send (git history) et retirer ce skip.
 */
test.describe('Billit invoicing (UI WIP)', { tag: ['@p1', '@invoicing'] }, () => {
  test('C7.0 page facturation — message en cours de développement', async ({ page }) => {
    if (INVOICING_UI_ENABLED) {
      test.skip(true, 'INVOICING_UI_ENABLED — restaurer le scénario connect/create e2e')
    }

    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    await page.goto('/invoicing', { waitUntil: 'networkidle' })
    if (!page.url().includes('/invoicing')) {
      test.skip(true, 'NUXT_PUBLIC_BILLIT_ENABLED off (redirect)')
    }

    await expect(page.getByTestId('invoicing-page')).toBeVisible()
    await expect(page.getByTestId('invoicing-under-development')).toBeVisible()
    await expect(page.getByTestId('invoicing-under-development')).toContainText(/développement|development|ontwikkeling|desarrollo|arendamisel|sviluppo/i)
    await expect(page.getByTestId('invoicing-connect-start')).toHaveCount(0)
    await expect(page.getByTestId('invoicing-create')).toHaveCount(0)
  })
})
