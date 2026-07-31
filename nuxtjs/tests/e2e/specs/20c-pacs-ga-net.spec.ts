import { test, expect } from '@playwright/test'
import {
  openFirstStudyWithFixture,
  pacsPublicFlagOff,
  pacsViewerEngine,
} from '../helpers/pacs'

/**
 * P2.3 GA net — download DICOM, frame OOR → 404, canvas frame clamp.
 * Soft-skip only when public PACS flag off.
 */
test.describe('PACS GA regression net', { tag: '@p0' }, () => {
  test('download .dcm + frame hors plage 404', async ({ page }) => {
    test.setTimeout(180000)
    test.skip(pacsPublicFlagOff(), 'NUXT_PUBLIC_PACS_ENABLED off')

    const engine = pacsViewerEngine()
    const { instanceId } = await openFirstStudyWithFixture(page)

    // API filet: out-of-range preview must be 404 (same cookies as the page).
    const oorStatus = await page.evaluate(async (id) => {
      const r = await fetch(`/api/pacs/instances/${encodeURIComponent(id)}/frames/99/preview`)
      return r.status
    }, instanceId)
    expect(oorStatus, 'frame OOR must map to 404').toBe(404)

    // Download via toolbar (canvas + cornerstone).
    await expect(page.getByTestId('dicom-tool-download')).toBeVisible()
    const [fileRes] = await Promise.all([
      page.waitForResponse(
        (r) => r.url().includes(`/instances/${instanceId}/file`) && r.request().method() === 'GET',
        { timeout: 30000 },
      ),
      page.getByTestId('dicom-tool-download').click(),
    ])
    expect(fileRes.status(), 'DICOM download').toBe(200)
    const buf = Buffer.from(await fileRes.body())
    expect(buf.length).toBeGreaterThan(132)
    expect(buf.subarray(128, 132).toString('ascii')).toBe('DICM')
    await expect(page.getByTestId('dicom-preview-error')).toHaveCount(0)

    // Frame next clamp: canvas only (preview frames). Metadata n=1 → next already disabled.
    if (engine !== 'cornerstone') {
      const nextBtn = page.getByTestId('dicom-tool-frame-next')
      await expect(nextBtn).toBeVisible()
      if (await nextBtn.isEnabled()) {
        const frame1 = page.waitForResponse(
          (r) => r.url().includes('/frames/1/preview') && r.request().method() === 'GET',
          { timeout: 30000 },
        )
        await nextBtn.click()
        const fr = await frame1
        expect(fr.status(), 'frame 1 OOR → 404').toBe(404)
        await expect(page.getByTestId('dicom-preview-error')).toHaveCount(0)
        await expect(nextBtn).toBeDisabled({ timeout: 10000 })
      }
    }
  })
})
