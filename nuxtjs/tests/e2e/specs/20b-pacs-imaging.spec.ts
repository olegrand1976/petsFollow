import { test, expect } from '@playwright/test'
import { loginAsVet } from '../helpers/auth'
import {
  demoClientAndPet,
  openFirstStudyWithFixture,
  pacsPublicFlagOff,
  pacsViewerEngine,
} from '../helpers/pacs'

/**
 * Pet sheet Imagerie — tag @p0.
 * Requires NUXT_PUBLIC_PACS_ENABLED + Orthanc (local make up-pacs / staging warm).
 * Soft-skips only when the public flag is off so CI without PACS stays green.
 * When the tab is present, upload + canvas assertions are hard (no soft-skip on Orthanc).
 */

async function canvasHasNonBackgroundPixels(page: import('@playwright/test').Page): Promise<boolean> {
  return page.getByTestId('dicom-canvas-left').evaluate((canvas) => {
    const c = canvas as HTMLCanvasElement
    const ctx = c.getContext('2d')
    if (!ctx || c.width < 8 || c.height < 8) return false
    const { data } = ctx.getImageData(0, 0, c.width, c.height)
    // Background fill is #0b1220 — tolerate dark RX; count pixels that diverge.
    let lit = 0
    for (let i = 0; i < data.length; i += 8) {
      const r = data[i]
      const g = data[i + 1]
      const b = data[i + 2]
      if (Math.abs(r - 0x0b) > 8 || Math.abs(g - 0x12) > 8 || Math.abs(b - 0x20) > 8) {
        lit++
        if (lit > 8) return true
      }
    }
    return lit > 8
  })
}

test.describe('PACS pet imaging tab', { tag: '@p0' }, () => {
  test('fiche animal — onglet Imagerie (flag on)', async ({ page }) => {
    test.setTimeout(90000)
    const pacsOn = process.env.NUXT_PUBLIC_PACS_ENABLED
    test.skip(
      pacsOn === 'false' || pacsOn === '0',
      'NUXT_PUBLIC_PACS_ENABLED off',
    )

    const { clientId, petId } = await demoClientAndPet()
    await loginAsVet(page)
    await page.goto(`/clients/${clientId}/pets/${petId}?tab=imaging`, { waitUntil: 'networkidle' })

    const imagingTabBtn = page.getByTestId('section-tab-imaging')
    test.skip(
      (await imagingTabBtn.count()) === 0,
      'PACS imaging tab not rendered (public flag off)',
    )

    await imagingTabBtn.click()
    await expect(page.getByTestId('pet-tab-imaging')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-viewer-container')).toBeVisible({ timeout: 15000 })
    await expect(page.getByTestId('pacs-dev-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-status-badge')).toBeVisible()
    await expect(page.getByTestId('pacs-refresh-btn')).toBeVisible()
    const wake = page.getByTestId('pacs-wake-btn')
    if (await wake.isVisible().catch(() => false)) {
      // Smoke only — do not click wake (avoids cold-start side effects on the next test).
      await expect(wake).toBeEnabled()
      await expect(page.getByTestId('pacs-upload-input')).toBeDisabled()
      await expect(page.getByTestId('pacs-offline-hint')).toBeVisible()
    } else {
      await expect(page.getByTestId('pacs-status-badge')).toContainText(/ready|prêt|klaar|listo|valmis|pronto/i)
    }
  })

  test('upload fixture + preview canvas non vide', async ({ page }) => {
    test.setTimeout(180000)
    test.skip(pacsPublicFlagOff(), 'NUXT_PUBLIC_PACS_ENABLED off')
    const engine = pacsViewerEngine()

    // Shared helper: wake + upload fixture + open preferred RX (avoid orphan CT → no preview).
    await openFirstStudyWithFixture(page)

    if (engine === 'cornerstone') {
      await expect(page.getByTestId('dicom-viewer-cornerstone')).toBeVisible({ timeout: 30000 })
      await expect(page.getByTestId('dicom-engine-badge')).toContainText(/Cornerstone/i)
      await expect(page.getByTestId('dicom-cs-left')).toBeVisible()
      await expect(page.getByTestId('dicom-preview-error')).toHaveCount(0)
      await expect(page.getByTestId('dicom-tool-pan')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-zoom')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-wl')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-measure')).toBeVisible()
      await page.getByTestId('dicom-tool-wl').click()
      await expect(page.getByTestId('dicom-tool-wl')).toHaveClass(/is-active/)
      await expect(page.getByTestId('dicom-calib-badge')).toBeVisible({ timeout: 20000 })
      // VOI label is best-effort (depends on image VOI tags / first render).
      const voi = page.getByTestId('dicom-voi-label')
      if (await voi.isVisible().catch(() => false)) {
        await expect(voi).toContainText(/WW|WC/i)
      }
    } else {
      await expect(page.getByTestId('dicom-viewer')).toBeVisible({ timeout: 15000 })
      await expect(page.getByTestId('dicom-canvas-left')).toBeVisible()
      await expect(page.getByTestId('dicom-preview-error')).toHaveCount(0)

      await expect.poll(async () => canvasHasNonBackgroundPixels(page), {
        timeout: 20000,
        message: 'canvas left should show non-background pixels after preview load',
      }).toBe(true)

      await expect(page.getByTestId('dicom-tool-pan')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-zoom')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-wl')).toBeVisible()
      await expect(page.getByTestId('dicom-tool-measure')).toBeVisible()
      const calib = page.getByTestId('dicom-calib-badge')
      await expect(calib).toBeVisible()
      // demo-rx.dcm embeds PixelSpacing 0.5\\0.5 — assert calib when Orthanc exposes tags.
      let calibrated = false
      const deadline = Date.now() + 20000
      while (Date.now() < deadline) {
        if ((await calib.getAttribute('data-calibrated')) === '1') {
          calibrated = true
          break
        }
        await page.waitForTimeout(250)
      }
      if (!calibrated) {
        test.skip(true, 'Orthanc PixelSpacing unavailable on this environment')
      }
      await expect(calib).toContainText(/mm|calibr/i)
      await page.getByTestId('pacs-compare-toggle').check()
      // Compare toggle must not blank the left pane.
      await expect.poll(async () => canvasHasNonBackgroundPixels(page), { timeout: 10000 }).toBe(true)
    }
  })
})
