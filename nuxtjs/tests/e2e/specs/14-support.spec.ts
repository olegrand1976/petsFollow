import { test, expect } from '@playwright/test'
import { loginAsAdmin, loginAsVet, fillField, waitForAuthForm } from '../helpers/auth'
import { deleteSupportTicket } from '../helpers/cleanup'

test.describe('support bug-report + admin inbox', { tag: '@p1' }, () => {
  test('véto envoie un ticket, admin le voit et répond', async ({ page }) => {
    test.setTimeout(90000)
    const subject = `E2E support ${Date.now()}`
    const replyText = `E2E reply ${Date.now()}`
    const messageText = 'Bouton calendrier ne répond plus (e2e).'
    const originPath = '/calendar'
    let ticketId = ''

    try {
      await loginAsVet(page)
      await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 15000 })
      await page.goto(originPath, { waitUntil: 'networkidle' })
      await expect(page).toHaveURL(/\/calendar/, { timeout: 15000 })
      await expect(page.getByTestId('pro-topbar')).toBeVisible({ timeout: 15000 })

      await page.getByTestId('pro-support-btn').click()
      await expect(page).toHaveURL(/\/support/, { timeout: 15000 })
      await waitForAuthForm(page, 'support-form')
      await fillField(page, 'support-subject', subject)
      const message = page.getByTestId('support-message')
      await message.fill(messageText)
      await expect(message).toHaveValue(/calendrier/)

      // UI submit required — origin capture is asserted from the POST body.
      const postWait = page.waitForResponse(
        (r) => r.request().method() === 'POST' && /\/api\/support\/tickets\/?$/.test(new URL(r.url()).pathname),
        { timeout: 15000 },
      ).catch(() => null)

      await page.getByTestId('support-form').evaluate((node) => {
        const form = node as HTMLFormElement
        if (typeof form.requestSubmit === 'function') form.requestSubmit()
        else form.dispatchEvent(new Event('submit', { bubbles: true, cancelable: true }))
      })

      let res = await postWait
      expect(res, 'support ticket must be created via UI form (origin capture)').toBeTruthy()
      const body = res!.request().postDataJSON() as {
        route?: string
        diagnostics?: { config?: { route?: string }; originPage?: { fullPath?: string } }
      }
      expect(body.route, 'ticket.route must be origin page').toMatch(/\/calendar/)
      expect(body.route).not.toBe('/support')
      expect(body.diagnostics?.config?.route).toMatch(/\/calendar/)
      expect(body.diagnostics?.originPage?.fullPath).toMatch(/\/calendar/)

      // BFF historically answered 200; create handlers now forward 201.
      expect([200, 201], `support ticket POST ${res!.status()}`).toContain(res!.status())
      const created = await res!.json().catch(() => null) as { data?: { id?: string }; id?: string } | null
      ticketId = String(created?.data?.id || created?.id || '')
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
      await expect(page.getByTestId('admin-support-origin-route')).toContainText(/\/calendar/)
      await expect(page.getByTestId('admin-support-diagnostics')).toBeVisible()

      const replyBox = page.getByTestId('admin-support-reply')
      await replyBox.fill(replyText)
      await Promise.all([
        page.waitForResponse(
          (r) =>
            r.request().method() === 'POST' &&
            /\/api\/admin\/support\/tickets\/[^/]+\/replies\/?$/.test(new URL(r.url()).pathname) &&
            (r.status() === 200 || r.status() === 201),
          { timeout: 20000 },
        ),
        page.getByTestId('admin-support-reply-submit').click(),
      ])
      await expect(page.getByTestId('admin-support-reply-msg')).toBeVisible({ timeout: 10000 })
      await expect(page.getByTestId('admin-support-replies')).toContainText(replyText)
    }
    finally {
      if (ticketId) {
        // Soft-fail path may still be véto — ensure admin before hard-delete.
        await loginAsAdmin(page).catch(() => undefined)
        await deleteSupportTicket(page, ticketId)
      }
    }
  })
})
