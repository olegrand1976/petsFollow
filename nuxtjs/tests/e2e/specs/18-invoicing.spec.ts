import { test, expect } from '@playwright/test'
import { fillField, login } from '../helpers/auth'
import { INVOICING_UI_ENABLED } from '../../../utils/invoicing-ui'

const STAFF_PASSWORD = 'VetDemo123!'

/**
 * Billit refuse tout envoi tant que le compte du cabinet n'a pas validé son
 * téléphone ou son IBAN — c'est l'état du compte sandbox branché sur staging,
 * pas une régression produit. On exige quand même que l'API le dise de façon
 * actionnable (409 `invoicing_account_unverified`, jamais un 502 opaque) avant
 * d'abandonner un scénario de livraison injouable ici. En mock (local, CI) le
 * cas ne se présente pas et les assertions de livraison tournent en entier.
 */
async function skipIfBillitAccountUnverified(res: { status: () => number, text: () => Promise<string> }) {
  if (res.status() !== 409) return
  if (!(await res.text()).includes('invoicing_account_unverified')) return
  test.skip(true, 'Compte Billit non vérifié (téléphone/IBAN) — envoi impossible sur cet environnement')
}

/** L'éditeur de types de RDV est dans l'onglet Agenda, replié par défaut. */
async function openVisitTypesSection(page: import('@playwright/test').Page) {
  await page.getByTestId('section-tab-calendar').click()
  const section = page.getByTestId('settings-visit-types')
  await expect(section).toBeVisible({ timeout: 15000 })
  if (!await section.evaluate((el: HTMLDetailsElement) => el.open)) {
    await section.locator('summary').click()
  }
}

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

  // Cœur du métier : un cabinet facture d'abord des particuliers. Pas de TVA,
  // pas de Peppol (Billit refuse un destinataire hors réseau) → envoi email.
  test('I7.10 facture particulier (sans TVA) → email', { tag: ['@p0'] }, async ({ page }) => {
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

    const startBtn = page.getByTestId('invoicing-connect-start')
    if (await startBtn.isVisible()) {
      await startBtn.click()
      await expect(page.getByTestId('invoicing-complete-form')).toBeVisible({ timeout: 15000 })
      await fillField(page, 'invoicing-complete-party', `party_e2e_b2c_${Date.now()}`)
      await fillField(page, 'invoicing-complete-apikey', 'mock-key-e2e')
      await page.getByTestId('invoicing-complete-submit').click()
    }

    await expect(page.getByTestId('invoicing-create')).toBeVisible({ timeout: 20000 })
    // Particulier par défaut : les champs fiscaux disparaissent.
    await expect(page.getByTestId('invoicing-cp-kind')).toHaveValue('individual')
    await expect(page.getByTestId('invoicing-cp-vat')).toHaveCount(0)
    await expect(page.getByTestId('invoicing-cp-company')).toHaveCount(0)

    await fillField(page, 'invoicing-cp-name', `Particulier E2E ${Date.now()}`)
    await fillField(page, 'invoicing-cp-email', 'particulier.e2e@petsfollow.test')
    await page.getByTestId('invoicing-country').selectOption('BE')
    await fillField(page, 'invoicing-cp-street', 'Rue des Fleurs 12')
    await fillField(page, 'invoicing-cp-postal', '1000')
    await fillField(page, 'invoicing-cp-city', 'Bruxelles')
    await fillField(page, 'invoicing-line-0-desc', 'Consultation particulier')
    await fillField(page, 'invoicing-line-0-qty', '1')
    await fillField(page, 'invoicing-line-0-unit', '45')

    const createRes = page.waitForResponse(
      (r) => r.url().includes('/api/invoicing/documents') && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('invoicing-new-doc').click()
    const created = await createRes
    expect([200, 201]).toContain(created.status())
    const createdJson = await created.json() as {
      data?: { id?: string, counterparty?: { customerKind?: string, vatNumber?: string } }
    }
    const createdData = createdJson.data || (createdJson as any)
    const docId = createdData.id
    expect(docId).toBeTruthy()
    expect(createdData.counterparty?.customerKind).toBe('individual')
    expect(createdData.counterparty?.vatNumber || '').toBe('')

    const sendRes = page.waitForResponse(
      (r) => r.url().includes(`/api/invoicing/documents/${docId}/send`) && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId(`invoicing-send-${docId}`).click()
    const sent = await sendRes
    await skipIfBillitAccountUnverified(sent)
    expect(sent.status(), await sent.text()).toBe(200)
    const sentBody = await sent.json() as { data?: { status?: string, peppolStatus?: string } }
    expect(['delivered', 'sending']).toContain(sentBody.data?.status || '')
    // L'audit doit dire email, pas Peppol.
    expect(sentBody.data?.peppolStatus || '').toMatch(/^email_/)

    // …mais le véto lit un libellé traduit, la valeur technique restant en infobulle.
    const delivery = page.getByTestId(`invoicing-delivery-${docId}`)
    await expect(delivery).toBeVisible()
    await expect(delivery).toHaveAttribute('title', /^email_/)
    await expect(delivery).not.toContainText('email_')
    await expect(delivery).toContainText(/E-?mail/)

    // Avoir sur cette facture : la contrepartie reprise reste un particulier.
    await page.getByTestId('invoicing-doc-type').selectOption('credit_note')
    await page.getByTestId('invoicing-related-invoice').selectOption(String(docId))
    await expect(page.getByTestId('invoicing-cp-kind')).toHaveValue('individual')
    await expect(page.getByTestId('invoicing-cp-vat')).toHaveCount(0)
    await page.getByTestId('invoicing-doc-type').selectOption('invoice')

    // Sans email, la facture d'un particulier n'a pas de canal de livraison.
    const noEmail = await page.request.post('/api/invoicing/documents', {
      data: {
        type: 'invoice',
        counterparty: {
          name: 'Particulier sans email',
          customerKind: 'individual',
          country: 'BE',
          street: 'Rue 1',
          city: 'Bruxelles',
          postal: '1000',
        },
        lines: [{ description: 'Consultation', quantity: 1, unitPriceExclCents: 4500, vatPercent: 21 }],
      },
    })
    expect(noEmail.status()).toBe(400)
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
    // Type repris de la fiche client (billingCustomerKind), sans ressaisie.
    await expect(page.getByTestId('invoicing-cp-kind')).toHaveValue('business')
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
    await skipIfBillitAccountUnverified(sent)
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

    // Avoir sur une facture pro : la contrepartie doit rester professionnelle,
    // sinon la TVA saute du document et l'avoir part par email au lieu de Peppol.
    await page.getByTestId('invoicing-doc-type').selectOption('credit_note')
    await page.getByTestId('invoicing-related-invoice').selectOption(String(docId))
    await expect(page.getByTestId('invoicing-cp-kind')).toHaveValue('business')
    await expect(page.getByTestId('invoicing-cp-vat')).toHaveValue('BE1000000021')
    await page.getByTestId('invoicing-doc-type').selectOption('invoice')

    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toBeVisible()
    await expect(page.getByTestId(`invoicing-doc-${pfId}`)).toContainText(/proforma/i)
    await expect(page.getByTestId(`invoicing-send-${pfId}`)).toHaveCount(0)
  })

  // BIL-9 : le tarif de l'acte vit sur le type de RDV. Sans aller-retour fiable
  // ici, la facture de fin de consultation repart vide et le véto ressaisit tout.
  test('I7.15 tarif du type de RDV — aller-retour /settings', async ({ page }) => {
    test.setTimeout(90_000)
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

    await page.goto('/settings', { waitUntil: 'networkidle' })
    await openVisitTypesSection(page)
    const priceInput = page.getByTestId('settings-visit-type-price').first()
    await expect(priceInput).toBeVisible({ timeout: 15000 })
    const initialPrice = await priceInput.inputValue()

    await priceInput.fill('45.50')
    await page.getByTestId('settings-visit-type-vat').first().fill('21')
    await page.getByTestId('settings-visit-types-save').click()

    // Relecture depuis l'API : le tarif est bien persisté en centimes HTVA.
    await expect(async () => {
      const res = await page.request.get('/api/vet/visit-types')
      expect(res.status()).toBe(200)
      const body = await res.json() as { data?: Array<{ priceExclCents?: number, vatPercent?: number }> }
      const first = (body.data ?? [])[0]
      expect(first?.priceExclCents).toBe(4550)
      expect(Number(first?.vatPercent)).toBe(21)
    }).toPass({ timeout: 15000 })

    await page.reload({ waitUntil: 'networkidle' })
    await openVisitTypesSection(page)
    await expect(page.getByTestId('settings-visit-type-price').first()).toHaveValue('45.50')

    // Tarif du seed remis en place : les autres suites (et la démo) comptent dessus.
    await page.getByTestId('settings-visit-type-price').first().fill(initialPrice)
    await page.getByTestId('settings-visit-types-save').click()
    await expect(async () => {
      const res = await page.request.get('/api/vet/visit-types')
      const body = await res.json() as { data?: Array<{ priceExclCents?: number }> }
      const restored = Math.round(Number(initialPrice || 0) * 100)
      expect((body.data ?? [])[0]?.priceExclCents).toBe(restored)
    }).toPass({ timeout: 15000 })
  })

  // Bout de chaîne visible du préremplissage : le tarif saisi dans /settings doit
  // arriver dans le formulaire de facture au retour de consultation. Le contrat
  // d'API est couvert côté Go, la conversion centimes → euros par Vitest ; ici on
  // vérifie que la page câble bien les deux.
  test('I7.14 lignes préremplies depuis la consultation', async ({ page }) => {
    test.setTimeout(90_000)
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

    const typesRes = await page.request.get('/api/vet/visit-types')
    expect(typesRes.status()).toBe(200)
    const types = ((await typesRes.json()) as {
      data?: Array<{ id: string, name: string, priceExclCents?: number }>
    }).data ?? []
    const tariffed = types.find((t) => Number(t.priceExclCents) > 0)
    if (!tariffed) {
      test.skip(true, 'aucun type de RDV tarifé (make seed)')
    }

    const clientsRes = await page.request.get('/api/clients')
    const clients = ((await clientsRes.json()) as {
      data?: Array<{ userId: string, email?: string }>
    }).data ?? []
    const demoClient = clients.find((c) => c.email === 'client.demo@petsfollow.test')
    expect(demoClient, 'client démo introuvable (make seed)').toBeTruthy()
    const petsRes = await page.request.get(`/api/clients/${demoClient!.userId}/pets`)
    const pets = ((await petsRes.json()) as { data?: Array<{ id: string }> }).data ?? []
    expect(pets.length, 'client démo sans animal (make seed)').toBeGreaterThan(0)

    const visitRes = await page.request.post(`/api/pets/${pets[0].id}/visits`, {
      data: {
        scheduledAt: new Date().toISOString(),
        durationMinutes: 30,
        confirmDirect: true,
        silentConfirm: true,
        consultationSession: true,
        visitTypeId: tariffed!.id,
        notes: 'e2e prefill',
      },
    })
    expect([200, 201]).toContain(visitRes.status())
    const visitId = String(((await visitRes.json()) as any)?.data?.id ?? '')
    expect(visitId).toBeTruthy()
    try {
      // `networkidle` est trop strict ici : la page garde des appels en vol
      // (documents, médicaments) et l'attente explicite ci-dessous suffit.
      await page.goto(`/invoicing?visitId=${visitId}&mode=direct`, { waitUntil: 'domcontentloaded' })
      await expect(page.getByTestId('invoicing-page')).toBeVisible({ timeout: 20000 })

      await expect(page.getByTestId('invoicing-line-0-desc')).toHaveValue(tariffed!.name, { timeout: 15000 })
      await expect(page.getByTestId('invoicing-line-0-unit'))
        .toHaveValue((Number(tariffed!.priceExclCents) / 100).toFixed(2))
      await expect(page.getByTestId('invoicing-consultation-context')).toContainText(tariffed!.name)
    }
    finally {
      // Supprimée plutôt qu'annulée : une consultation annulée resterait visible
      // dans l'agenda de démo, une par exécution de la suite.
      await page.request.delete(`/api/visits/${visitId}`).catch(() => undefined)
    }
  })
})
