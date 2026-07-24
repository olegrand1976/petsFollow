import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import fs from 'node:fs'
import os from 'node:os'
import path from 'node:path'

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

test('envoi pièce jointe image sur un thread', async ({ page }) => {
  await loginAsVet(page)
  await page.goto('/messages')
  await expect(page.getByTestId('messages-page')).toBeVisible()

  const threadBtn = page.locator('[data-testid^="thread-"]').first()
  if ((await threadBtn.count()) === 0) {
    test.skip(true, 'Aucun thread de messagerie')
    return
  }
  await threadBtn.click()

  const attachBtn = page.getByTestId('messages-attach')
  await expect(attachBtn).toBeVisible()

  // Minimal valid 1x1 JPEG
  const jpeg = Buffer.from(
    '/9j/4AAQSkZJRgABAQAAAQABAAD/2wBDAAgGBgcGBQgHBwcJCQgKDBQNDAsLDBkSEw8UHRofHh0aHBwgJC4nICIsIxwcKDcpLDAxNDQ0Hyc5PTgyPC4zNDL/2wBDAQkJCQwLDBgNDRgyIRwhMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjIyMjL/wAARCAABAAEDASIAAhEBAxEB/8QAFQABAQAAAAAAAAAAAAAAAAAAAAn/xAAUEAEAAAAAAAAAAAAAAAAAAAAA/8QAFQEBAQAAAAAAAAAAAAAAAAAAAAX/xAAUEQEAAAAAAAAAAAAAAAAAAAAA/9oADAMBAAIQAxAAAAGfAP/EABQQAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQEAAQUCf//EABQRAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQMBAT8Bf//EABQRAQAAAAAAAAAAAAAAAAAAAAD/2gAIAQIBAT8Bf//Z',
    'base64',
  )
  const tmp = path.join(os.tmpdir(), `pf-e2e-media-${Date.now()}.jpg`)
  fs.writeFileSync(tmp, jpeg)

  try {
    const fileInput = page.locator('input[type="file"]')
    await fileInput.setInputFiles(tmp)
    await expect(page.locator('.pro-chat__media-img').last()).toBeVisible({ timeout: 20000 })
  } finally {
    fs.unlinkSync(tmp)
  }
})
