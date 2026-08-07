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
    test.setTimeout(120_000)
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

    // Autocomplete client.demo (billing seed) → préremplit la contrepartie
    const clientSearch = page.getByTestId('invoicing-client-search').getByTestId('pro-combobox-input')
    await clientSearch.fill('Sophie')
    await expect(page.getByTestId('pro-combobox-list')).toBeVisible({ timeout: 8000 })
    await page.getByTestId('pro-combobox-list').locator('[role="option"]').first().click()
    await expect(page.getByTestId('invoicing-cp-name')).toHaveValue(/Sophie/)
    await expect(page.getByTestId('invoicing-cp-vat')).toHaveValue('BE1000000021')
    await expect(page.getByTestId('invoicing-cp-street')).toHaveValue(/Loi/)
    await expect(page.getByTestId('invoicing-cp-city')).toHaveValue('Bruxelles')

    await fillField(page, 'invoicing-cp-name', `Client E2E ${Date.now()}`)
    await page.getByTestId('invoicing-country').selectOption('BE')
    await fillField(page, 'invoicing-cp-vat', 'BE1000000021')
    await fillField(page, 'invoicing-cp-street', 'Rue E2E 1')
    await fillField(page, 'invoicing-cp-postal', '1000')
    await fillField(page, 'invoicing-cp-city', 'Bruxelles')
    await expect(page.getByTestId('invoicing-lines')).toBeVisible()
    await fillField(page, 'invoicing-line-0-desc', 'Consultation E2E')
    await fillField(page, 'invoicing-line-0-qty', '1')
    await fillField(page, 'invoicing-line-0-unit', '42')
    await page.getByTestId('invoicing-add-line').click()
    await fillField(page, 'invoicing-line-1-desc', 'Acte complémentaire')
    await fillField(page, 'invoicing-line-1-qty', '2')
    await fillField(page, 'invoicing-line-1-unit', '10')
    await page.getByTestId('invoicing-line-1-vat').selectOption('6')
    await expect(page.getByTestId('invoicing-totals')).toContainText('62.00')
    // 42 + 2*10 = 62 HT ; TVA 21%*42 + 6%*20 = 8.82+1.20 = 10.02 ; TTC 72.02
    await expect(page.getByTestId('invoicing-totals')).toContainText('72.02')

    const createRes = page.waitForResponse(
      (r) => r.url().includes('/api/invoicing/documents') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('invoicing-new-doc').click()
    const created = await createRes
    expect([200, 201]).toContain(created.status())
    const createdJson = await created.json() as {
      data?: { id?: string, totalExclCents?: number, totalInclCents?: number, lines?: unknown[] }
      id?: string
      totalExclCents?: number
      totalInclCents?: number
      lines?: unknown[]
    }
    const createdData = createdJson.data || createdJson
    const docId = createdData.id
    expect(docId).toBeTruthy()
    expect(createdData.totalExclCents).toBe(6200)
    expect(createdData.totalInclCents).toBe(7202)

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
          email: 'client.pf@petsfollow.test',
        },
        lines: [{ description: 'Devis', quantity: 1, unitPriceExclCents: 2000, vatPercent: 21 }],
      },
    })
    expect([200, 201]).toContain(okProforma.status())
    const pfBody = await okProforma.json() as { data?: { id?: string }, id?: string }
    const pfId = pfBody.data?.id || pfBody.id
    expect(pfId).toBeTruthy()

    // Proforma = email + magic-link (no Billit Offer). Accept → invoice Peppol.
    const sendPf = await page.request.post(`/api/invoicing/documents/${pfId}/send`)
    expect(sendPf.status()).toBe(200)
    const sendPfBody = await sendPf.json() as {
      data?: { status?: string, acceptPath?: string, billitOrderId?: string }
      status?: string
      acceptPath?: string
    }
    expect(sendPfBody.data?.status || sendPfBody.status).toBe('issued')
    expect(sendPfBody.data?.billitOrderId || '').toBe('')

    const acceptPath = sendPfBody.data?.acceptPath || sendPfBody.acceptPath || ''
    if (acceptPath) {
      const token = acceptPath.replace(/^\/proforma\//, '')
      const accept = await page.request.post(`/api/public/proforma/${token}/accept`)
      expect(accept.status()).toBe(200)
      const acceptBody = await accept.json() as { data?: { proforma?: { status?: string }, invoice?: { status?: string } } }
      expect(acceptBody.data?.proforma?.status).toBe('accepted')
      expect(['delivered', 'sending', 'issued']).toContain(acceptBody.data?.invoice?.status || '')
    }

    const resendPf = await page.request.post(`/api/invoicing/documents/${pfId}/send`)
    expect(resendPf.status()).toBeGreaterThanOrEqual(400)

    await page.goto('/invoicing', { waitUntil: 'networkidle' })
    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toBeVisible()
    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toContainText(/proforma/i)
    await expect(page.getByTestId(`invoicing-send-${pfId}`)).toHaveCount(0)
  })
})
