import type { Page } from '@playwright/test'

/** Identité mock alignée sur `go/internal/eid/testdata/sample_valid.eid`. */
export const eidMockIdentity = {
  firstname: 'Camille',
  lastname: 'Testeur',
  niss: '96072399886',
  country: 'BE',
  address_street: 'Rue Demo 1',
  address_zip: '1000',
  address_city: 'Bruxelles',
  import_tool: 'eid_viewer_xml',
} as const

export const eidMockWebIdentity = {
  firstname: 'Camille',
  lastname: 'Testeur',
  niss: '96072399886',
  country: 'BE',
  import_tool: 'web_eid',
} as const

/** Stub lib Web eID (pas d’extension / PIN en CI). À poser avant navigation Pro. */
export async function installWebEidLibMock(page: Page) {
  await page.addInitScript(() => {
    ;(window as Window & { __PF_WEB_EID_MOCK__?: unknown }).__PF_WEB_EID_MOCK__ = {
      status: async () => ({ extension: true, nativeApp: true }),
      authenticate: async (nonce: string) => ({ unverifiedCertificate: `e2e-${nonce}` }),
    }
  })
}

export async function mockEidViewerImport(page: Page, identity = eidMockIdentity) {
  await page.route('**/api/vet/eid/import', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: identity }),
    })
  })
}

export async function mockEidWebEidApi(page: Page, identity = eidMockWebIdentity) {
  await page.route('**/api/vet/eid/web-eid/challenge', async (route) => {
    const origin = new URL(page.url()).origin
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({
        data: { nonce: 'e2e-nonce-web-eid', origin },
      }),
    })
  })
  await page.route('**/api/vet/eid/web-eid/verify', async (route) => {
    await route.fulfill({
      status: 200,
      contentType: 'application/json',
      body: JSON.stringify({ data: identity }),
    })
  })
}
