import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'

test('messagerie affiche les conversations', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/messages')
  await expect(page.getByTestId('messages-page')).toBeVisible()
  await expect(page.getByRole('heading', { name: /messagerie|messaging|berichten/i })).toBeVisible()
})

test('deep-link thread depuis query', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/messages')
  const threadBtn = page.locator('[data-testid^="thread-"]').first()
  if (await threadBtn.count()) {
    const testId = await threadBtn.getAttribute('data-testid')
    const threadId = testId?.replace('thread-', '')
    await page.goto(`/messages?thread=${threadId}`)
    await expect(page.getByTestId('messages-page')).toBeVisible()
    await expect(threadBtn).toHaveClass(/pro-chat__thread-btn--active/)
  }
})
