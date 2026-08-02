import { test, expect } from '@playwright/test'
import {
  loginAsAdmin,
  loginAsCommercial,
  loginAsCommercialManager,
  loginAsVet,
} from '../helpers/auth'

test('commercial ouvre /presentation via la nav', async ({ page }) => {
  await loginAsCommercial(page)
  await expect(page.getByTestId('nav-presentation')).toBeVisible()
  await page.getByTestId('nav-presentation').click()
  await expect(page).toHaveURL(/\/presentation/, { timeout: 15000 })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('presentation-toc')).toBeVisible()
})

test('commercial navigue les steps et deep-link ?step=', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/presentation?step=welcome', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 15000 })
  const points = page.getByTestId('presentation-points')
  await expect(points).toBeVisible()
  await expect(points.locator('li').first()).toContainText(/\S/)
  await Promise.all([
    page.waitForURL(/[?&]step=ai_in_app(?:&|$)/, { timeout: 10000 }),
    page.getByTestId('presentation-toc-ai_in_app').click(),
  ])
  await expect(page.getByTestId('presentation-open-ai-flows')).toBeVisible()
  await expect(page.getByTestId('presentation-points').locator('li').first()).toContainText(/\S/)

  await page.goto('/presentation?step=close', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('presentation-cta')).toBeVisible({ timeout: 15000 })
})

test('?step invalide est canonisé vers welcome', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/presentation?step=nope', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/[?&]step=welcome(?:&|$)/, { timeout: 15000 })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 15000 })
})

test('redirect /commercial/presentation préserve ?step=', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/presentation?step=offer', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/\/presentation\?step=offer/, { timeout: 15000 })
  await expect(page.getByTestId('presentation-offer')).toBeVisible()
})

test('commercial ouvre /flux-ia (profil vet par défaut)', async ({ page }) => {
  await loginAsCommercial(page)
  await expect(page.getByTestId('nav-flux-ia')).toBeVisible()
  await page.getByTestId('nav-flux-ia').click()
  await expect(page).toHaveURL(/\/flux-ia/, { timeout: 15000 })
  await expect(page.getByTestId('ai-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('ai-flow-profile-vet')).toHaveAttribute('aria-checked', 'true')
  await expect(page.getByTestId('ai-flow-diagram-vet_cr_ai')).toBeVisible()
  await expect(page.getByTestId('mermaid-diagram')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('mermaid-svg-host').locator('svg')).toBeVisible({ timeout: 15000 })
})

test('commercial change de profil AI flows', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/flux-ia?profile=vet', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('ai-flows-page')).toBeVisible({ timeout: 15000 })
  await Promise.all([
    page.waitForURL(/[?&]profile=commercial(?:&|$)/, { timeout: 10000 }),
    page.getByTestId('ai-flow-profile-commercial').click(),
  ])
  await expect(page.getByTestId('ai-flow-diagram-co_encode_activate')).toBeVisible()
})

test('admin accède à /presentation et /flux-ia', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page.getByTestId('nav-presentation')).toBeVisible({ timeout: 10000 })
  await page.goto('/presentation', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('presentation-wizard')).toBeVisible({ timeout: 15000 })
  await page.goto('/flux-ia', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('ai-flows-page')).toBeVisible({ timeout: 15000 })
})

test('responsable commercial accède à /flux-ia', async ({ page }) => {
  await loginAsCommercialManager(page)
  await expect(page.getByTestId('nav-flux-ia')).toBeVisible()
  await page.goto('/flux-ia', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('ai-flows-page')).toBeVisible({ timeout: 15000 })
})

test('véto est redirigé hors de /presentation et /flux-ia', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/presentation', { waitUntil: 'domcontentloaded' })
  await expect(page).not.toHaveURL(/\/presentation/, { timeout: 15000 })
  await page.goto('/flux-ia', { waitUntil: 'domcontentloaded' })
  await expect(page).not.toHaveURL(/\/flux-ia/, { timeout: 15000 })
})
