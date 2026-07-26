import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsVet, fillField } from '../helpers/auth'

test.describe('support bug-report + admin inbox', { tag: '@p1' }, () => {
  test('véto envoie un ticket, admin le voit et répond', async ({ page }) => {
    const subject = `E2E support ${Date.now()}`
    const replyText = `E2E reply ${Date.now()}`

    await loginAsVet(page)
    await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 15000 })
    await page.getByTestId('pro-support-btn').evaluate((node) => (node as HTMLElement).click())
    await expect(page.getByTestId('support-form')).toBeVisible({ timeout: 15000 })
    await fillField(page, 'support-subject', subject)
    const message = page.getByTestId('support-message')
    await message.fill('Bouton calendrier ne répond plus (e2e).')
    await expect(message).toHaveValue(/calendrier/)

    await Promise.all([
      page.waitForResponse(
        (r) =>
          r.request().method() === 'POST' &&
          /\/api\/support\/tickets\/?$/.test(new URL(r.url()).pathname) &&
          r.status() === 201,
        { timeout: 20000 },
      ),
      page.getByTestId('support-submit').click(),
    ])
    await expect(page.getByTestId('support-success')).toBeVisible({ timeout: 10000 })

    await loginAsAdmin(page)
    await page.goto('/admin/support', { waitUntil: 'networkidle' })
    await expect(page.getByTestId('admin-support-page')).toBeVisible()
    await expect(page.getByTestId('admin-support-pager')).toBeVisible()

    const row = page.getByTestId('admin-support-row').filter({ hasText: subject })
    await expect(row).toBeVisible({ timeout: 15000 })
    await row.click()
    await expect(page.getByTestId('admin-support-detail-page')).toBeVisible()
    await expect(page.getByTestId('admin-support-message')).toContainText(/calendrier/i)
    await expect(page.getByTestId('admin-support-diagnostics')).toBeVisible()

    const replyBox = page.getByTestId('admin-support-reply')
    await replyBox.fill(replyText)
    await Promise.all([
      page.waitForResponse(
        (r) =>
          r.request().method() === 'POST' &&
          /\/api\/admin\/support\/tickets\/[^/]+\/replies\/?$/.test(new URL(r.url()).pathname) &&
          r.status() === 201,
        { timeout: 20000 },
      ),
      page.getByTestId('admin-support-reply-submit').click(),
    ])
    await expect(page.getByTestId('admin-support-reply-msg')).toBeVisible({ timeout: 10000 })
    await expect(page.getByTestId('admin-support-replies')).toContainText(replyText)
  })
})
