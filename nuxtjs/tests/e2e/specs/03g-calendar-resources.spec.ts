import { test, expect } from '@playwright/test'
import { loginAsVet, nativeClick } from '../helpers/auth'
import { ensureMultiSite } from '../helpers/sites'

/**
 * C3.17 — calendar resources (rooms + day view people/rooms).
 * Requires API + NUXT_PUBLIC_SITES_UI_ENABLED on. Ensures ≥2 sites on staging.
 */
test.describe('calendar resources rooms + day view', { tag: '@p1' }, () => {
  test('vet.demo — rooms CRUD, SITE_ALL gate, day toggle, create with assignee', async ({ page }) => {
    test.setTimeout(120_000)
    await loginAsVet(page)
    await ensureMultiSite(page)

    await page.goto('/sites')
    await expect(page.getByTestId('sites-page')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('settings-sites')).toBeVisible()

    const siteRows = page.getByTestId('settings-sites-list').locator('[data-testid^="settings-site-row-"]')
    await expect.poll(async () => siteRows.count()).toBeGreaterThanOrEqual(1)
    const rowTestId = await siteRows.first().getAttribute('data-testid')
    const siteId = (rowTestId || '').replace('settings-site-row-', '')
    expect(siteId).toBeTruthy()

    const roomsBlock = page.getByTestId(`settings-rooms-${siteId}`)
    await expect(roomsBlock).toBeVisible({ timeout: 10000 })

    const roomName = `Box e2e ${Date.now()}`
    await page.getByTestId(`settings-room-name-${siteId}`).fill(roomName)
    await page.getByTestId(`settings-room-create-${siteId}`).click()
    await expect(roomsBlock.getByText(roomName, { exact: true })).toBeVisible({ timeout: 10000 })

    const roomRow = roomsBlock.locator('[data-testid^="settings-room-row-"]').filter({ hasText: roomName })
    const roomRowTestId = await roomRow.getAttribute('data-testid')
    const roomId = (roomRowTestId || '').replace('settings-room-row-', '')
    expect(roomId).toBeTruthy()

    await page.getByTestId(`settings-room-deactivate-${roomId}`).click()
    await expect(page.getByTestId(`settings-room-inactive-${roomId}`)).toBeVisible({ timeout: 10000 })
    await page.getByTestId(`settings-room-reactivate-${roomId}`).click()
    await expect(page.getByTestId(`settings-room-inactive-${roomId}`)).toHaveCount(0)
    await expect(page.getByTestId(`settings-room-deactivate-${roomId}`)).toBeVisible()

    await page.goto('/calendar')
    await expect(page.getByTestId('calendar-page')).toHaveAttribute('data-calendar-ready', '1', {
      timeout: 15000,
    })

    const siteSelect = page.getByTestId('pro-site-select')
    await expect(siteSelect).toBeVisible({ timeout: 10000 })
    await siteSelect.selectOption('all')
    await nativeClick(page, 'calendar-view-day')
    await expect(page.getByTestId('calendar-view-day')).toHaveAttribute('aria-pressed', 'true', {
      timeout: 10000,
    })
    await expect(page.getByTestId('calendar-day-needs-site')).toBeVisible({ timeout: 10000 })

    const options = siteSelect.locator('option')
    const count = await options.count()
    let concreteSite = ''
    for (let i = 0; i < count; i++) {
      const val = await options.nth(i).getAttribute('value')
      if (val && val !== 'all' && val !== '') {
        concreteSite = val
        await siteSelect.selectOption(val)
        break
      }
    }
    expect(concreteSite).toBeTruthy()

    await expect(page.getByTestId('calendar-day-needs-site')).toHaveCount(0)
    await expect(page.getByTestId('calendar-day-grid')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('calendar-resource-toggle')).toBeVisible()
    await nativeClick(page, 'calendar-resource-rooms')
    await expect(page.getByTestId('calendar-resource-rooms')).toHaveAttribute('aria-pressed', 'true')
    await nativeClick(page, 'calendar-resource-people')
    await expect(page.getByTestId('calendar-resource-people')).toHaveAttribute('aria-pressed', 'true')
    await expect(page.getByTestId('calendar-day-col-empty')).toBeVisible()

    // Create RDV with assignee + room via modal
    await page.getByTestId('calendar-new-appointment').click()
    await expect(page.getByTestId('new-appointment-modal')).toBeVisible({ timeout: 10000 })
    const clientSelect = page.getByTestId('new-appt-client')
    await expect.poll(async () => clientSelect.locator('option').count(), { timeout: 15000 }).toBeGreaterThan(1)
    const petSelect = page.getByTestId('new-appt-pet')
    const clientOpts = clientSelect.locator('option')
    const clientCount = await clientOpts.count()
    let petReady = false
    for (let i = 0; i < clientCount; i++) {
      const val = await clientOpts.nth(i).getAttribute('value')
      if (!val) continue
      await clientSelect.selectOption(val)
      try {
        await expect.poll(async () => petSelect.locator('option').count(), { timeout: 8000 }).toBeGreaterThan(1)
        petReady = true
        break
      } catch {
        // try next client (seed clients without pets)
      }
    }
    expect(petReady).toBe(true)
    const petOpts = petSelect.locator('option')
    for (let i = 0; i < await petOpts.count(); i++) {
      const val = await petOpts.nth(i).getAttribute('value')
      if (val) {
        await petSelect.selectOption(val)
        break
      }
    }

    const assigneeSelect = page.getByTestId('new-appt-assignee')
    await expect(assigneeSelect).toBeVisible()
    const assigneeOpts = assigneeSelect.locator('option')
    for (let i = 0; i < await assigneeOpts.count(); i++) {
      const val = await assigneeOpts.nth(i).getAttribute('value')
      if (val) {
        await assigneeSelect.selectOption(val)
        break
      }
    }
    const roomSelect = page.getByTestId('new-appt-room')
    await expect(roomSelect).toBeVisible()
    const roomOpts = roomSelect.locator('option')
    let pickedRoom = false
    for (let i = 0; i < await roomOpts.count(); i++) {
      const val = await roomOpts.nth(i).getAttribute('value')
      const label = (await roomOpts.nth(i).textContent()) || ''
      if (val && label.includes(roomName.slice(0, 12))) {
        await roomSelect.selectOption(val)
        pickedRoom = true
        break
      }
    }
    if (!pickedRoom) {
      for (let i = 0; i < await roomOpts.count(); i++) {
        const val = await roomOpts.nth(i).getAttribute('value')
        if (val) {
          await roomSelect.selectOption(val)
          break
        }
      }
    }

    // Far-out weekday + unique afternoon slot to avoid assignee_busy on busy staging.
    const future = new Date()
    future.setDate(future.getDate() + 28 + (Date.now() % 5))
    while (future.getDay() === 0 || future.getDay() === 6) {
      future.setDate(future.getDate() + 1)
    }
    const yyyy = future.getFullYear()
    const mm = String(future.getMonth() + 1).padStart(2, '0')
    const dd = String(future.getDate()).padStart(2, '0')
    await page.getByTestId('new-appt-day').fill(`${yyyy}-${mm}-${dd}`)
    const hour = 14 + (Date.now() % 3)
    const slotMin = String((Date.now() % 12) * 5).padStart(2, '0')
    await page.getByTestId('new-appt-time').fill(`${hour}:${slotMin}`)

    const createRes = page.waitForResponse(
      (r) => /\/api\/pets\/[^/]+\/visits\b/.test(r.url()) && r.request().method() === 'POST',
      { timeout: 20000 },
    )
    await page.getByTestId('new-appt-confirm').click()
    const created = await createRes
    expect(created.status(), await created.text()).toBeLessThan(400)
    await expect(page.getByTestId('new-appointment-modal')).toHaveCount(0, { timeout: 20000 })
  })
})
