import { test, expect } from '@playwright/test'
import {
  loginAsAdmin,
  loginAsCommercial,
  loginAsCommercialManager,
  loginAsVet,
} from '../helpers/auth'

test('commercial ouvre /flux via la nav', async ({ page }) => {
  await loginAsCommercial(page)
  await expect(page.getByTestId('nav-flux')).toBeVisible()
  await page.getByTestId('nav-flux').click()
  await expect(page).toHaveURL(/\/flux/, { timeout: 15000 })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('product-flows-profiles')).toBeVisible()
})

test('commercial change de profil et deep-link ?profile=', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/flux', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await Promise.all([
    page.waitForURL(/[?&]profile=commercial(?:&|$)/, { timeout: 10000 }),
    page.getByTestId('flow-profile-commercial').click(),
  ])
  await expect(page.getByTestId('flow-profile-commercial')).toHaveAttribute('aria-checked', 'true', {
    timeout: 10000,
  })

  await page.goto('/flux?profile=client', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('flow-profile-client')).toHaveAttribute('aria-checked', 'true')
  await expect(page.getByTestId('product-flows-steps')).toBeVisible()
})

test('?profile invalide est canonisé vers vet', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/flux?profile=nope', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/[?&]profile=vet(?:&|$)/, { timeout: 15000 })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('flow-profile-vet')).toHaveAttribute('aria-checked', 'true')
})

test('redirect /commercial/flux préserve ?profile=', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/commercial/flux?profile=client', { waitUntil: 'domcontentloaded' })
  await expect(page).toHaveURL(/\/flux\?profile=client/, { timeout: 15000 })
  await expect(page.getByTestId('flow-profile-client')).toHaveAttribute('aria-checked', 'true')
})

test('lien croisé bascule le profil', async ({ page }) => {
  await loginAsCommercial(page)
  await page.goto('/flux?profile=commercial', { waitUntil: 'networkidle' })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await Promise.all([
    page.waitForURL(/[?&]profile=vet(?:&|$)/, { timeout: 10000 }),
    page.getByTestId('flow-link-encode_co_vet').click(),
  ])
  await expect(page.getByTestId('flow-profile-vet')).toHaveAttribute('aria-checked', 'true')
})

test('admin accède à /flux', async ({ page }) => {
  await loginAsAdmin(page)
  await expect(page.getByTestId('nav-flux')).toBeVisible({ timeout: 10000 })
  await page.goto('/flux', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('product-flows-profiles')).toBeVisible()
})

test('responsable commercial accède à /flux', async ({ page }) => {
  await loginAsCommercialManager(page)
  await expect(page.getByTestId('nav-flux')).toBeVisible()
  await page.goto('/flux', { waitUntil: 'domcontentloaded' })
  await expect(page.getByTestId('product-flows-page')).toBeVisible({ timeout: 15000 })
  await expect(page.getByTestId('flow-profile-commercial_manager')).toBeVisible()
})

test('véto est redirigé hors de /flux', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/flux', { waitUntil: 'domcontentloaded' })
  await expect(page).not.toHaveURL(/\/flux/, { timeout: 15000 })
  await expect(page).toHaveURL(/\/(dashboard|onboarding|welcome)/, { timeout: 15000 })
  await expect(page.getByTestId('product-flows-page')).toHaveCount(0)
})
