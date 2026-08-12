import type { APIRequestContext, Page } from '@playwright/test'

type Deleter = Pick<APIRequestContext, 'delete'>

function asDeleter(pageOrReq: Page | APIRequestContext): Deleter {
  if ('request' in pageOrReq && pageOrReq.request) {
    return pageOrReq.request
  }
  return pageOrReq as APIRequestContext
}

/** Best-effort DELETE: never throws; logs non-404 failures so pollution is visible. */
async function bestEffortDelete(deleter: Deleter, url: string, label: string) {
  try {
    const res = await deleter.delete(url)
    if (!res.ok() && res.status() !== 404) {
      console.warn(`[e2e cleanup] ${label} ${url} → HTTP ${res.status()}`)
    }
  } catch (err) {
    console.warn(`[e2e cleanup] ${label} ${url} failed:`, err)
  }
}

/** Soft-delete walk-in consultation (hidden from /consultations). Best-effort. */
export async function softDeleteVisit(pageOrReq: Page | APIRequestContext, visitId: string) {
  const id = String(visitId || '').trim()
  if (!id) return
  await bestEffortDelete(asDeleter(pageOrReq), `/api/visits/${id}`, 'softDeleteVisit')
}

export async function softDeleteVisits(
  pageOrReq: Page | APIRequestContext,
  visitIds: Iterable<string>,
) {
  for (const id of visitIds) {
    await softDeleteVisit(pageOrReq, id)
  }
}

/** Hard-delete support ticket (admin/dev session). Best-effort. */
export async function deleteSupportTicket(pageOrReq: Page | APIRequestContext, ticketId: string) {
  const id = String(ticketId || '').trim()
  if (!id) return
  await bestEffortDelete(
    asDeleter(pageOrReq),
    `/api/admin/support/tickets/${id}`,
    'deleteSupportTicket',
  )
}
