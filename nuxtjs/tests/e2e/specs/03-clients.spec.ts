import { test, expect, type Page } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import {
  eidMockIdentity,
  installWebEidLibMock,
  mockEidViewerImport,
  mockEidWebEidApi,
} from '../helpers/eid'

/** /clients peut ouvrir un ProModal (invitations) qui bloque les clics. */
async function dismissProModals(page: Page) {
  for (let i = 0; i < 3; i++) {
    const close = page.getByTestId('pro-modal-close')
    if (await close.isVisible().catch(() => false)) {
      await close.first().click({ force: true }).catch(() => {})
      await page.waitForTimeout(200)
    } else {
      break
    }
  }
}

async function openCreateClientWithEid(page: Page) {
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await page.waitForLoadState('networkidle')
  await dismissProModals(page)
  await page.getByTestId('create-client-open').click()
  await expect(page.getByTestId('create-client-modal')).toBeVisible()
  const eid = page.getByTestId('eid-reader')
  if (!(await eid.isVisible().catch(() => false))) {
    test.skip(true, 'eID UI masquée (NUXT_PUBLIC_EID_ENABLED / cabinet BE)')
  }
  return eid
}

test('liste clients avec recherche', { tag: '@p0' }, async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)

  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  await dismissProModals(page)

  const search = page.locator('#client-search')
  await search.fill('Sophie')
  // Timeout large : premier chargement de la page en dev (compilation Vite) sous suite complète.
  // Téléphone seed / edit fiche : couverture Go `TestClientContactPhone*` (staging sans re-seed fiable).
  await expect(page.getByText(/Sophie Demo|client\.demo/i).first()).toBeVisible({ timeout: 15000 })
})

// Modale seule, aucune création : la bascule Particulier doit retirer les champs
// fiscaux, sinon un n° de TVA saisi ici disparaîtrait sans le dire à la facturation.
test('création client : le type particulier masque TVA et n° d\'entreprise', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/clients')
  await expect(page.getByTestId('clients-page')).toBeVisible()
  // La modale « invitations en attente » s'ouvre d'elle-même dès que le seed en
  // contient, après le chargement : attendre puis fermer, sinon elle avale le clic.
  await page.waitForLoadState('networkidle')
  await dismissProModals(page)

  await page.getByTestId('create-client-open').click()
  await expect(page.getByTestId('create-client-billing-section')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('create-client-billing-vat')).toBeVisible()

  await page.getByTestId('create-client-billing-customer-kind').selectOption('individual')
  await expect(page.getByTestId('create-client-billing-vat')).toHaveCount(0)
  await expect(page.getByTestId('create-client-billing-company')).toHaveCount(0)

  await page.getByTestId('create-client-billing-customer-kind').selectOption('business')
  await expect(page.getByTestId('create-client-billing-vat')).toBeVisible()
})

test.describe('eID BE prefill', { tag: '@p0' }, () => {
  test('création client : prefill eID Viewer (mock BFF)', async ({ page }) => {
    await loginAsVet(page)
    await mockEidViewerImport(page)
    await openCreateClientWithEid(page)

    await expect(page.getByTestId('eid-read-card')).toBeVisible()
    await expect(page.getByTestId('eid-import-input')).toBeAttached()
    await expect(page.getByTestId('eid-local-hint')).toHaveCount(0)
    await expect(page.getByTestId('eid-dev-badge')).toBeVisible()

    await page.getByTestId('eid-import-input').setInputFiles({
      name: 'sample_valid.eid',
      mimeType: 'application/xml',
      buffer: Buffer.from(
        '<?xml version="1.0"?><Export><surname>Testeur</surname><firstname>Camille</firstname><nationalnumber>96072399886</nationalnumber></Export>',
      ),
    })
    await expect(page.getByTestId('create-client-first-name')).toHaveValue(eidMockIdentity.firstname, {
      timeout: 10000,
    })
    await expect(page.getByTestId('create-client-last-name')).toHaveValue(eidMockIdentity.lastname)
    await expect(page.getByTestId('create-client-niss')).toHaveValue(eidMockIdentity.niss)
    await expect(page.getByTestId('create-client-billing-country')).toHaveValue('BE')
    await expect(page.getByTestId('create-client-billing-postal')).toHaveValue(eidMockIdentity.address_zip)
    await expect(page.getByTestId('eid-reader-msg')).toBeVisible()
  })

  test('création client : prefill Web eID (mock lib + BFF)', async ({ page }) => {
    await page.addInitScript(() => {
      ;(window as Window & { __PF_WEB_EID_MOCK__?: unknown }).__PF_WEB_EID_MOCK__ = {
        status: async () => ({ extension: true, nativeApp: true }),
        authenticate: async (nonce: string) => {
          await new Promise((r) => setTimeout(r, 80))
          return { unverifiedCertificate: `e2e-${nonce}` }
        },
      }
    })
    await loginAsVet(page)
    await mockEidWebEidApi(page)
    await openCreateClientWithEid(page)

    await page.getByTestId('eid-read-card').click()
    await expect(page.getByTestId('eid-reader-busy')).toBeVisible()
    await expect(page.getByTestId('create-client-first-name')).toHaveValue('Camille', { timeout: 10000 })
    await expect(page.getByTestId('create-client-last-name')).toHaveValue('Testeur')
    await expect(page.getByTestId('create-client-niss')).toHaveValue('96072399886')
    await expect(page.getByTestId('create-client-billing-country')).toHaveValue('BE')
    await expect(page.getByTestId('eid-reader-msg')).toBeVisible()
    await expect(page.getByTestId('eid-reader-error')).toHaveCount(0)
    await expect(page.getByTestId('eid-reader-busy')).toHaveCount(0)
  })

  test('création client : Web eID origin mismatch affiche l’erreur', async ({ page }) => {
    await installWebEidLibMock(page)
    await loginAsVet(page)
    await page.route('**/api/vet/eid/web-eid/challenge', async (route) => {
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({
          data: { nonce: 'e2e-nonce-mismatch', origin: 'https://evil.example' },
        }),
      })
    })
    await openCreateClientWithEid(page)

    await page.getByTestId('eid-read-card').click()
    await expect(page.getByTestId('eid-reader-error')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('create-client-first-name')).toHaveValue('')
  })

  test('création client : Web eID timeout/erreur lib affiche l’alerte', async ({ page }) => {
    await page.addInitScript(() => {
      ;(window as Window & { __PF_WEB_EID_MOCK__?: unknown }).__PF_WEB_EID_MOCK__ = {
        status: async () => ({ extension: true, nativeApp: true }),
        authenticate: async () => {
          throw Object.assign(new Error('user_timeout'), { code: 'user_timeout' })
        },
      }
    })
    await loginAsVet(page)
    await page.route('**/api/vet/eid/web-eid/challenge', async (route) => {
      const origin = new URL(page.url()).origin
      await route.fulfill({
        status: 200,
        contentType: 'application/json',
        body: JSON.stringify({ data: { nonce: 'e2e-timeout', origin } }),
      })
    })
    await openCreateClientWithEid(page)
    await page.getByTestId('eid-read-card').click()
    await expect(page.getByTestId('eid-reader-error')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('create-client-first-name')).toHaveValue('')
  })
})
