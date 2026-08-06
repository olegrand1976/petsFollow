import { test, expect } from '@playwright/test'
import { fillField, login } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

const STAFF_PASSWORD = 'VetDemo123!'

test.describe('Billit invoicing (mock)', { tag: ['@p1', '@invoicing'] }, () => {
  test.describe.configure({ mode: 'serial' })

  test('I7.0 UI WIP — message en cours de développement', async ({ page }) => {
    if (INVOICING_UI_ENABLED) {
      test.skip(true, 'INVOICING_UI_ENABLED — scénario métier I7.1+')
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
    await expect(page.getByTestId('invoicing-connect-start')).toHaveCount(0)
    await expect(page.getByTestId('invoicing-create')).toHaveCount(0)
  })

  test('I7 connect → facture BE delivered + credit/proforma/IT', async ({ page }) => {
    if (!INVOICING_UI_ENABLED) {
      test.skip(true, 'INVOICING_UI_ENABLED off')
    }

    const { status } = await login(page, 'vet.demo@petsfollow.test', STAFF_PASSWORD)
    expect(status).toBe(200)
    await page.waitForURL((url) => url.pathname.includes('/dashboard'), { timeout: 20000 })

    const probe = await page.request.get('/api/invoicing/connection')
    if (probe.status() === 404) {
      test.skip(true, 'BILLIT_ENABLED off')
    }
    expect(probe.status(), await probe.text()).toBe(200)

    await page.goto('/invoicing', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('invoicing-page')).toBeVisible()
    await expect(page.getByTestId('invoicing-under-development')).toHaveCount(0)

    const startBtn = page.getByTestId('invoicing-connect-start')
    if (await startBtn.isVisible()) {
      await startBtn.click()
      await expect(page.getByTestId('invoicing-complete-form')).toBeVisible({ timeout: 15000 })
      await fillField(page, 'invoicing-complete-party', `party_e2e_${Date.now()}`)
      await fillField(page, 'invoicing-complete-apikey', 'mock-key-e2e')
      await page.getByTestId('invoicing-complete-submit').click()
    }

    await expect(page.getByTestId('invoicing-create')).toBeVisible({ timeout: 20000 })
    await expect(page.getByTestId('invoicing-doc-type')).toHaveValue('invoice')

    await fillField(page, 'invoicing-cp-name', `Client E2E ${Date.now()}`)
    await page.getByTestId('invoicing-country').selectOption('BE')
    await fillField(page, 'invoicing-cp-vat', 'BE1000000021')
    await fillField(page, 'invoicing-cp-street', 'Rue E2E 1')
    await fillField(page, 'invoicing-cp-postal', '1000')
    await fillField(page, 'invoicing-cp-city', 'Bruxelles')
    await fillField(page, 'invoicing-line-desc', 'Consultation E2E')
    await fillField(page, 'invoicing-line-amount', '42')

    const createRes = page.waitForResponse(
      (r) => r.url().includes('/api/invoicing/documents') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('invoicing-new-doc').click()
    const created = await createRes
    expect([200, 201]).toContain(created.status())
    const createdBody = await created.json() as { data?: { id?: string }, id?: string }
    const docId = createdBody.data?.id || createdBody.id
    expect(docId).toBeTruthy()

    const sendBtn = page.getByTestId(`invoicing-send-${docId}`)
    await expect(sendBtn).toBeVisible()
    const sendRes = page.waitForResponse(
      (r) => r.url().includes(`/api/invoicing/documents/${docId}/send`) && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await sendBtn.click()
    const sent = await sendRes
    expect(sent.status(), await sent.text()).toBe(200)
    const sentBody = await sent.json() as { data?: { status?: string }, status?: string }
    // Mock → delivered sync ; staging live sandbox → sending until Peppol webhook.
    const sendStatus = sentBody.data?.status || sentBody.status
    expect(['delivered', 'sending', 'issued']).toContain(sendStatus)
    // Live sandbox stays on sending until Peppol webhook (mock → delivered sync).
    const liveBillit = sendStatus === 'sending'
    await expect(page.getByTestId(`invoicing-doc-${docId}`)).toBeVisible()
    await expect(page.getByTestId(`invoicing-send-${docId}`)).toHaveCount(0)

    const badIt = await page.request.post('/api/invoicing/documents', {
      data: {
        type: 'invoice',
        counterparty: {
          name: 'Cliente IT',
          country: 'IT',
          vatNumber: 'IT12345678901',
          street: 'Via Roma 1',
          city: 'Roma',
          postal: '00100',
        },
        lines: [{ description: 'Visita', quantity: 1, unitPriceExclCents: 5000, vatPercent: 22 }],
      },
    })
    expect(badIt.status()).toBe(400)

    const badCredit = await page.request.post('/api/invoicing/documents', {
      data: {
        type: 'credit_note',
        counterparty: {
          name: 'Client CN',
          country: 'BE',
          vatNumber: 'BE1000000021',
          street: 'Rue 1',
          city: 'Bruxelles',
          postal: '1000',
        },
        lines: [{ description: 'Avoir', quantity: 1, unitPriceExclCents: 1000, vatPercent: 21 }],
      },
    })
    expect(badCredit.status()).toBe(400)

    const okCredit = await page.request.post('/api/invoicing/documents', {
      data: {
        type: 'credit_note',
        relatedDocumentId: docId,
        counterparty: {
          name: 'Client CN',
          country: 'BE',
          vatNumber: 'BE1000000021',
          street: 'Rue 1',
          city: 'Bruxelles',
          postal: '1000',
        },
        lines: [{ description: 'Avoir', quantity: 1, unitPriceExclCents: 1000, vatPercent: 21 }],
      },
    })
    expect([200, 201]).toContain(okCredit.status())

    const okProforma = await page.request.post('/api/invoicing/documents', {
      data: {
        type: 'proforma',
        counterparty: {
          name: 'Client PF',
          country: 'BE',
          vatNumber: 'BE1000000120',
          street: 'Rue 2',
          city: 'Bruxelles',
          postal: '1000',
        },
        lines: [{ description: 'Devis', quantity: 1, unitPriceExclCents: 2000, vatPercent: 21 }],
      },
    })
    expect([200, 201]).toContain(okProforma.status())
    const pfBody = await okProforma.json() as { data?: { id?: string }, id?: string }
    const pfId = pfBody.data?.id || pfBody.id
    expect(pfId).toBeTruthy()

    // ProForma issue hits Billit create; sandbox parties may 502 — mock path asserts issued.
    if (!liveBillit) {
      const sendPf = await page.request.post(`/api/invoicing/documents/${pfId}/send`)
      expect(sendPf.status()).toBe(200)
      const sendPfBody = await sendPf.json() as { data?: { status?: string }, status?: string }
      expect(sendPfBody.data?.status || sendPfBody.status).toBe('issued')

      const resendPf = await page.request.post(`/api/invoicing/documents/${pfId}/send`)
      expect(resendPf.status()).toBeGreaterThanOrEqual(400)
    }

    await page.goto('/invoicing', { waitUntil: 'networkidle' })
    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toBeVisible()
    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toContainText(/proforma/i)
    if (!liveBillit) {
      await expect(page.getByTestId(`invoicing-send-${pfId}`)).toHaveCount(0)
    }
  })
})
