import { test, expect } from '@playwright/test'
import { loginAsVet, nativeClick } from '../helpers/auth'

test('changement de langue dans paramètres', async ({ page }) => {
  await loginAsVet(page)
  await expect(page).toHaveURL(/dashboard/)

  // Pas de networkidle : la WebSocket notifications du shell reste ouverte en continu.
  await page.goto('/settings')
  // Attendre l’init async (preferredLocale) — peut être fr ou en selon runs précédents.
  const activeLocale = page.locator('[data-testid^="settings-locale-"].pro-toggle-btn--active')
  await expect(activeLocale).toBeVisible({ timeout: 15000 })

  if (await page.getByTestId('settings-locale-en').evaluate((el) => el.classList.contains('pro-toggle-btn--active'))) {
    await nativeClick(page, 'settings-locale-fr')
    await nativeClick(page, 'settings-locale-save')
    await expect(page.getByTestId('settings-locale-fr')).toHaveClass(/pro-toggle-btn--active/, {
      timeout: 15000,
    })
  }

  // Retry : l'init async preferredLocale peut réinitialiser le toggle juste après le clic.
  await expect(async () => {
    await nativeClick(page, 'settings-locale-en')
    await expect(page.getByTestId('settings-locale-en')).toHaveClass(/pro-toggle-btn--active/, {
      timeout: 2000,
    })
  }).toPass({ timeout: 15000 })
  await nativeClick(page, 'settings-locale-save')
  await expect(page.getByTestId('settings-locale-en')).toHaveClass(/pro-toggle-btn--active/, {
    timeout: 15000,
  })
})
